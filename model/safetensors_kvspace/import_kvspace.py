#!/usr/bin/env python3
"""safetensors → kvspace 全局变量导入器。

每个张量以 XValue(ref=inline, storetype=ARRAYND, langtype="[dims]dtype") 写入
全局路径 <root><name>。按 langtype 一次性预留 head，WriteNewPlace 返回 body 偏移指针后
直接灌原始字节——head 长度先行占定，body 不再挪动（zero-move）。

用法:
    KVSPACE=redis://127.0.0.1:6379 \
    python3 import_kvspace.py <模型目录> [--root /model/<name>/] [--upcast-fp32]

<模型目录> 含 model.safetensors 或 model-*-of-*.safetensors（自动识别分片）与 config.json。
"""
import argparse
import ctypes
import glob
import json
import mmap
import os
import re
import struct
import sys

ST_KIND = {
    "F64": "float64", "F32": "float32", "F16": "float16", "BF16": "bfloat16",
    "I64": "int64", "I32": "int32", "I16": "int16", "I8": "int8",
    "U64": "uint64", "U32": "uint32", "U16": "uint16", "U8": "uint8", "BOOL": "bool",
}
REF_INLINE = 0
STORETYPE_ARRAYND = 2
U8P = ctypes.POINTER(ctypes.c_uint8)


def load_lib():
    lib = ctypes.CDLL(os.environ.get("KVSPACE_SO", "libkvspace.so"))
    lib.kvspaceConnect.argtypes = [ctypes.c_char_p]
    lib.kvspaceConnect.restype = ctypes.c_void_p
    lib.kvspaceClose.argtypes = [ctypes.c_void_p]
    lib.kvspaceClose.restype = None
    lib.kvspaceMkindex.argtypes = [ctypes.c_void_p, ctypes.c_char_p, ctypes.c_uint32,
                                   ctypes.c_char_p, ctypes.c_uint32]
    lib.kvspaceMkindex.restype = ctypes.c_int
    lib.kvspaceWriteNewPlace.argtypes = [ctypes.c_void_p, ctypes.c_char_p, ctypes.c_uint8,
                                         ctypes.c_uint8, ctypes.c_uint8, ctypes.c_uint32,
                                         ctypes.c_char_p, ctypes.c_uint32, ctypes.POINTER(U8P),
                                         ctypes.c_char_p, ctypes.c_uint32]
    lib.kvspaceWriteNewPlace.restype = ctypes.c_int
    return lib


def find_shards(model_dir):
    single = os.path.join(model_dir, "model.safetensors")
    if os.path.exists(single):
        return [single]
    pat = re.compile(r"model-(\d+)-of-(\d+)\.safetensors")
    shards = [(int(pat.search(os.path.basename(f)).group(1)), f)
              for f in glob.glob(os.path.join(model_dir, "model-*-of-*.safetensors"))
              if pat.search(os.path.basename(f))]
    if not shards:
        raise FileNotFoundError(f"无 safetensors 于 {model_dir}")
    shards.sort()
    return [f for _, f in shards]


def parse_header(mm):
    n = struct.unpack("<Q", mm[:8])[0]
    return json.loads(mm[8:8 + n].decode()), 8 + n


def mkindex_chain(lib, kv, root):
    err = ctypes.create_string_buffer(256)
    parts = [p for p in root.split("/") if p]
    cur = "/"
    for p in parts:
        cur += p + "/"
        lib.kvspaceMkindex(kv, cur.encode(), 0, err, 256)


def new_place(lib, kv, key, langtype, n):
    err = ctypes.create_string_buffer(256)
    dst = U8P()
    rc = lib.kvspaceWriteNewPlace(kv, key.encode(), REF_INLINE, STORETYPE_ARRAYND,
                                  0, 0, langtype.encode(), n, ctypes.byref(dst), err, 256)
    if rc != 0:
        raise RuntimeError(f"WriteNewPlace {key}: {err.value.decode()}")
    return dst


def write_value(lib, kv, key, langtype, src_addr, n):
    dst = new_place(lib, kv, key, langtype, n)
    if n > 0 and dst:
        ctypes.memmove(dst, src_addr, n)


def import_file(lib, kv, path, root, upcast, entries):
    import gc
    import numpy as np
    f = open(path, "rb")
    mm = mmap.mmap(f.fileno(), 0, access=mmap.ACCESS_READ)
    hdr, base = parse_header(mm)
    arr = np.frombuffer(mm, dtype=np.uint8)  # mmap 零拷贝视图，仅取基址
    baseaddr = arr.ctypes.data
    count = 0
    try:
        for name, meta in hdr.items():
            if name == "__metadata__":
                continue
            dt, shape = meta["dtype"], meta["shape"]
            b0, b1 = meta["data_offsets"]
            if upcast and dt in ("F16", "BF16"):
                import torch
                tdt = torch.float16 if dt == "F16" else torch.bfloat16
                wide = torch.frombuffer(arr[base + b0:base + b1], dtype=tdt).to(torch.float32).numpy()
                langtype = "[" + ",".join(map(str, shape)) + "]float32"
                write_value(lib, kv, root + name, langtype, wide.ctypes.data, wide.nbytes)
                del wide
            else:
                langtype = "[" + ",".join(map(str, shape)) + "]" + ST_KIND[dt]
                # memmove 直接从 mmap 页地址灌入 kvspace body（无中间对象）
                write_value(lib, kv, root + name, langtype, baseaddr + base + b0, b1 - b0)
            entries.append((root + name, langtype))
            count += 1
    finally:
        del arr
        gc.collect()
        mm.close()
        f.close()
    return count


def emit_kv(path, model_name, root, entries, config):
    lines = [f"// {model_name} —— safetensors 导入 kvspace 的模型表示（{os.path.basename(__file__)} 自动生成，勿手改）",
             f"// 权重已写入下列全局路径为 XValue(ref=inline, storetype=ARRAYND)。张量数 {len(entries)}。"]
    if config:
        keys = ("model_type", "hidden_size", "num_hidden_layers", "num_attention_heads",
                "num_key_value_heads", "head_dim", "intermediate_size", "vocab_size", "torch_dtype")
        summ = " ".join(f"{k}={config[k]}" for k in keys if k in config)
        lines.append(f"// config: {summ}")
    lines.append("")
    for key, lt in entries:
        lines.append(f"{key} : {lt}")
    lines.append("")
    os.makedirs(os.path.dirname(path), exist_ok=True)
    open(path, "w").write("\n".join(lines))
    print(f"kv 表示 → {path}")


def import_aux(lib, kv, model_dir, root):
    for fn in ("config.json", "generation_config.json", "tokenizer_config.json"):
        p = os.path.join(model_dir, fn)
        if os.path.exists(p):
            raw = open(p, "rb").read()
            write_value(lib, kv, root + fn, f"[{len(raw)}]char/utf8", raw, len(raw))


def main():
    ap = argparse.ArgumentParser(description="safetensors → kvspace 全局变量导入器")
    ap.add_argument("model_dir")
    ap.add_argument("--root", default=None, help="全局根，默认 /model/<目录名>/")
    ap.add_argument("--upcast-fp32", action="store_true", help="F16/BF16 升为 float32")
    ap.add_argument("--emit-kv", default=None, help="额外输出 kvlang 模型表示到该 .kv 路径")
    args = ap.parse_args()

    dsn = os.environ.get("KVSPACE")
    if not dsn:
        sys.exit("需设置 KVSPACE 环境变量（redis://... 或 shm://...）")
    root = args.root or f"/model/{os.path.basename(os.path.normpath(args.model_dir))}/"
    if not root.endswith("/"):
        root += "/"

    lib = load_lib()
    kv = lib.kvspaceConnect(dsn.encode())
    if not kv:
        sys.exit(f"connect {dsn} 失败")
    entries = []
    try:
        mkindex_chain(lib, kv, root)
        total = 0
        for shard in find_shards(args.model_dir):
            n = import_file(lib, kv, shard, root, args.upcast_fp32, entries)
            total += n
            print(f"  {os.path.basename(shard)}: {n} 张量")
        import_aux(lib, kv, args.model_dir, root)
        print(f"导入完成: {total} 张量 → {root}")
    finally:
        lib.kvspaceClose(kv)

    if args.emit_kv:
        cfg_path = os.path.join(args.model_dir, "config.json")
        config = json.load(open(cfg_path)) if os.path.exists(cfg_path) else {}
        name = os.path.basename(os.path.normpath(args.model_dir))
        emit_kv(args.emit_kv, name, root, sorted(entries), config)


if __name__ == "__main__":
    main()
