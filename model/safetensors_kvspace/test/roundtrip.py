#!/usr/bin/env python3
"""离线自测: 合成小 safetensors(qwen3/llama3 命名 + bf16/f32) → 导入 kvspace → 读回校验。

校验每个张量: langtype == "[dims]dtype"、body 字节 == 原始 safetensors 切片(bf16 保真零拷贝)。
需 KVSPACE(默认 redis://127.0.0.1:6379) 与 libkvspace.so。
"""
import ctypes
import json
import os
import struct
import subprocess
import sys
import tempfile

import torch
from safetensors.torch import save_file

HERE = os.path.dirname(os.path.abspath(__file__))
IMPORTER = os.path.join(os.path.dirname(HERE), "import_kvspace.py")
U8P = ctypes.POINTER(ctypes.c_uint8)


class Head(ctypes.Structure):
    _fields_ = [
        ("headlen", ctypes.c_uint16), ("ref", ctypes.c_uint8), ("storetype", ctypes.c_uint8),
        ("ro", ctypes.c_uint8), ("vid", ctypes.c_uint32), ("body_len", ctypes.c_int32),
        ("ndim", ctypes.c_int32), ("dims", ctypes.c_int32 * 8), ("langtype", ctypes.c_char * 256),
        ("langtype_len", ctypes.c_int32), ("body_offset", ctypes.c_int32),
    ]


def build_model(d):
    tensors = {
        "model.embed_tokens.weight": torch.randn(128, 32).to(torch.bfloat16),
        "model.layers.0.self_attn.q_proj.weight": torch.randn(32, 32).to(torch.bfloat16),
        "model.layers.0.mlp.gate_proj.weight": torch.randn(64, 32).to(torch.bfloat16),
        "model.norm.weight": torch.randn(32).to(torch.float32),
        "lm_head.weight": torch.randn(128, 32).to(torch.bfloat16),
    }
    save_file(tensors, os.path.join(d, "model.safetensors"))
    json.dump({"model_type": "qwen3", "hidden_size": 32, "num_hidden_layers": 1},
              open(os.path.join(d, "config.json"), "w"))
    return tensors


def parse_st(path):
    mm = open(path, "rb").read()
    n = struct.unpack("<Q", mm[:8])[0]
    hdr = json.loads(mm[8:8 + n].decode())
    base = 8 + n
    out = {}
    for k, m in hdr.items():
        if k == "__metadata__":
            continue
        b0, b1 = m["data_offsets"]
        out[k] = (m["dtype"], m["shape"], mm[base + b0:base + b1])
    return out


def main():
    dsn = os.environ.get("KVSPACE", "redis://127.0.0.1:6379")
    os.environ["KVSPACE"] = dsn
    lib = ctypes.CDLL(os.environ.get("KVSPACE_SO", "libkvspace.so"))
    lib.kvspaceConnect.argtypes = [ctypes.c_char_p]; lib.kvspaceConnect.restype = ctypes.c_void_p
    lib.kvspaceClose.argtypes = [ctypes.c_void_p]; lib.kvspaceClose.restype = None
    lib.kvspaceGet.argtypes = [ctypes.c_void_p, ctypes.c_char_p, ctypes.c_int,
                               ctypes.POINTER(U8P), ctypes.POINTER(ctypes.c_uint32)]
    lib.kvspaceGet.restype = ctypes.c_int
    lib.kvspaceDecodeHead.argtypes = [U8P, ctypes.c_uint32, ctypes.POINTER(Head)]
    lib.kvspaceDecodeHead.restype = ctypes.c_int

    d = tempfile.mkdtemp(prefix="stkv_")
    tensors = build_model(d)
    expect = parse_st(os.path.join(d, "model.safetensors"))
    root = "/model/synth/"

    r = subprocess.run([sys.executable, IMPORTER, d, "--root", root],
                       capture_output=True, text=True)
    print(r.stdout, r.stderr)
    if r.returncode != 0:
        sys.exit("导入失败")

    kv = lib.kvspaceConnect(dsn.encode())
    ok = 0
    fail = 0
    for name, (dt, shape, raw) in expect.items():
        kind = {"BF16": "bfloat16", "F32": "float32", "F16": "float16"}[dt]
        want_lt = "[" + ",".join(map(str, shape)) + "]" + kind
        out, n = U8P(), ctypes.c_uint32()
        lib.kvspaceGet(kv, (root + name).encode(), 0, ctypes.byref(out), ctypes.byref(n))
        if not out or n.value == 0:
            print(f"FAIL {name}: 空值"); fail += 1; continue
        data = ctypes.string_at(out, n.value)
        buf = (ctypes.c_uint8 * len(data)).from_buffer_copy(data)
        h = Head()
        lib.kvspaceDecodeHead(buf, len(data), ctypes.byref(h))
        got_lt = h.langtype.decode()
        got_dims = list(h.dims[:h.ndim])
        body = data[h.body_offset:h.body_offset + h.body_len]
        if got_lt == want_lt and got_dims == shape and body == raw:
            ok += 1
        else:
            fail += 1
            print(f"FAIL {name}: lt={got_lt!r} want {want_lt!r} dims={got_dims} shape={shape} "
                  f"bytes={'==' if body == raw else '!='}({len(body)}/{len(raw)})")
    lib.kvspaceClose(kv)
    print(f"=== roundtrip PASS={ok} FAIL={fail} ===")
    sys.exit(1 if fail else 0)


if __name__ == "__main__":
    main()
