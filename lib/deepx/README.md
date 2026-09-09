# deepx 模型前端表示

用 kvlang 描述神经网络的前端表示，组织参照 pytorch `torch.nn`：模块是 `rwfunc`，算子调用是厂商中立的前端抽象算子 `deepx·{op}`，权重是 `/` 全局变量。

## 两层算子（对齐 kvlang spec 卷 06）

一个张量运算有两层键，前端代码只写第一层：

| 层 | 键形态 | 例 | 谁提供 |
|----|--------|-----|--------|
| 前端抽象算子 | `/lib/deepx·{op}` | `deepx·matmul` | 稳定算子面，无实现、无 author |
| 后端实现算子 | `/lib/deepx/{author}·{op}` | `deepx/miaobyte·matmul`、`deepx/cblas·matmul` | 计算 runtime 启动注册 |

前端表示一律写 `deepx·{op}`；扩展编译器扫 `/lib`，按已注册后端把每个调用点绑定到具体实现（`deepx·matmul → deepx/cblas·matmul`），复合算子则分解/融合，产物交 runtime 纯解释。未经编译直接执行抽象算子 → runtime 报「抽象算子未绑定后端」。

## 目录

```
deepx/
  nn.kv              # 对标 torch.nn.functional：linear / embedding / rmsnorm / silu / softmax
  models/qwen3.kv    # 对标 transformers.Qwen3：mlp / attention / decoder_layer
```

- `deepx/nn` 模块是薄封装，把 pytorch 语义的模块名映射到抽象算子组合。
- `deepx/models/qwen3` 逐层组装 Qwen3；跨包调用走全路径 `deepx/nn·linear(...)`（kvlang 无 import）。

## 权重：`/` 全局变量，pytorch state_dict 点分名

权重由 `deepx/model/safetensors_kvspace/import_kvspace.py` 从 safetensors 零拷贝导入，键为 pytorch state_dict 名：

```
/model/Qwen3-0.6B/model.embed_tokens.weight              [151936,1024]bfloat16
/model/Qwen3-0.6B/model.layers.0.self_attn.q_proj.weight [2048,1024]bfloat16
/model/Qwen3-0.6B/model.norm.weight                      [1024]bfloat16
```

forward 以绝对路径字面量引用之，或按层号动态拼路径经 `kv·get` 取值传入模块 rwfunc。`.` 在 kvlang 路径字面量中是子键分隔符，天然对应 pytorch 的点分层级。

## 前端算子清单与后端缺口

`deepx/nn`、`deepx/models/qwen3` 依赖的抽象算子：

| 抽象算子 | 已有后端（deepx-cpu-compute） | 说明 |
|----------|------------------------------|------|
| `deepx·matmul` `deepx·add` `deepx·mul` `deepx·transpose` `deepx·sum` | ✅ `deepx/miaobyte·*`、`deepx/cblas·matmul` | 基础算子已就位 |
| `deepx·rmsnorm` `deepx·silu` `deepx·softmax` | ⏳ 复合，可由基础算子分解（exp/sum/div、mean/rsqrt/mul） | 编译器分解或后端直接实现 |
| `deepx·embedding` | ⏳ gather 行，待后端 | — |
| `deepx·rope` `deepx·sdpa` | ⏳ 高层注意力算子，待后端 | 内部处理多头 reshape / 因果 mask |

`⏳` 项是稳定前端算子面的一部分，前端表示已按其语义书写；后端由计算引擎逐步兑现或编译器分解，不影响本层表示。

## 文法约束：调用实参不跨行

kvlang 以**换行（或 `;`）为语句分隔符**（文法 `stmt_sep = newline | ";"`，见 spec 01-词法/01-源码结构），且无「括号内换行抑制」规则。故一条**调用的实参列表必须写在一行**——跨行会被词法器拆成两条语句。本目录一律遵守（`rwfunc` 签名折行不受此限，签名解析跟踪括号/方括号深度）。文法未显式写明此点，此处备注。
