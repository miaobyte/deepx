#!/usr/bin/env python3
"""真实模型抽检: 连 kvspace 读回若干关键张量, 校验 dims/langtype 与 config 推导的期望形状一致。

用法:
    KVSPACE=redis://127.0.0.1:6379 \
    python3 verify_model.py --root /model/Qwen3-0.6B/ --config /home/.../Qwen3-0.6B/config.json
"""
import argparse
import ctypes
import json
import os
import sys

U8P = ctypes.POINTER(ctypes.c_uint8)


class Head(ctypes.Structure):
    _fields_ = [
        ("headlen", ctypes.c_uint16), ("ref", ctypes.c_uint8), ("storetype", ctypes.c_uint8),
        ("ro", ctypes.c_uint8), ("vid", ctypes.c_uint32), ("body_len", ctypes.c_int32),
        ("ndim", ctypes.c_int32), ("dims", ctypes.c_int32 * 8), ("langtype", ctypes.c_char * 256),
        ("langtype_len", ctypes.c_int32), ("body_offset", ctypes.c_int32),
    ]


def expected(cfg):
    h = cfg["hidden_size"]
    heads = cfg["num_attention_heads"]
    kvh = cfg["num_key_value_heads"]
    hd = cfg.get("head_dim", h // heads)
    ffn = cfg["intermediate_size"]
    vocab = cfg["vocab_size"]
    dt = {"bfloat16": "bfloat16", "float16": "float16", "float32": "float32"}[cfg["torch_dtype"]]
    q, kv = heads * hd, kvh * hd
    e = {
        "model.embed_tokens.weight": ([vocab, h], dt),
        "model.layers.0.self_attn.q_proj.weight": ([q, h], dt),
        "model.layers.0.self_attn.k_proj.weight": ([kv, h], dt),
        "model.layers.0.self_attn.v_proj.weight": ([kv, h], dt),
        "model.layers.0.self_attn.o_proj.weight": ([h, q], dt),
        "model.layers.0.mlp.gate_proj.weight": ([ffn, h], dt),
        "model.layers.0.mlp.up_proj.weight": ([ffn, h], dt),
        "model.layers.0.mlp.down_proj.weight": ([h, ffn], dt),
        "model.layers.0.input_layernorm.weight": ([h], dt),
        "model.norm.weight": ([h], dt),
    }
    if not cfg.get("tie_word_embeddings"):
        e["lm_head.weight"] = ([vocab, h], dt)
    return e


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--root", required=True)
    ap.add_argument("--config", required=True)
    args = ap.parse_args()
    dsn = os.environ.get("KVSPACE") or sys.exit("需 KVSPACE")
    cfg = json.load(open(args.config))
    exp = expected(cfg)

    lib = ctypes.CDLL(os.environ.get("KVSPACE_SO", "libkvspace.so"))
    lib.kvspaceConnect.argtypes = [ctypes.c_char_p]; lib.kvspaceConnect.restype = ctypes.c_void_p
    lib.kvspaceClose.argtypes = [ctypes.c_void_p]; lib.kvspaceClose.restype = None
    lib.kvspaceGet.argtypes = [ctypes.c_void_p, ctypes.c_char_p, ctypes.c_int,
                               ctypes.POINTER(U8P), ctypes.POINTER(ctypes.c_uint32)]
    lib.kvspaceGet.restype = ctypes.c_int
    lib.kvspaceDecodeHead.argtypes = [U8P, ctypes.c_uint32, ctypes.POINTER(Head)]
    lib.kvspaceDecodeHead.restype = ctypes.c_int

    root = args.root if args.root.endswith("/") else args.root + "/"
    kv = lib.kvspaceConnect(dsn.encode())
    ok = fail = 0
    for name, (shape, dt) in exp.items():
        want = "[" + ",".join(map(str, shape)) + "]" + dt
        out, n = U8P(), ctypes.c_uint32()
        lib.kvspaceGet(kv, (root + name).encode(), 0, ctypes.byref(out), ctypes.byref(n))
        if not out or n.value == 0:
            print(f"FAIL {name}: 空值"); fail += 1; continue
        data = ctypes.string_at(out, n.value)
        buf = (ctypes.c_uint8 * len(data)).from_buffer_copy(data)
        h = Head()
        lib.kvspaceDecodeHead(buf, len(data), ctypes.byref(h))
        got_lt, got_dims = h.langtype.decode(), list(h.dims[:h.ndim])
        nbytes = 1
        for x in shape:
            nbytes *= x
        nbytes *= 2 if dt in ("bfloat16", "float16") else 4
        if got_lt == want and got_dims == shape and h.body_len == nbytes:
            ok += 1; print(f"OK   {name}  {got_lt}  body={h.body_len}")
        else:
            fail += 1
            print(f"FAIL {name}: lt={got_lt!r} want {want!r} dims={got_dims} "
                  f"body={h.body_len} want {nbytes}")
    lib.kvspaceClose(kv)
    print(f"=== verify PASS={ok} FAIL={fail} ===")
    sys.exit(1 if fail else 0)


if __name__ == "__main__":
    main()
