# codegen — Redis Key 多语言 SDK 生成器

> 单一真源: `spec/keys.yaml` → Go / C++ / TypeScript

## 用法

```
cd tool/codegen && go run . <project_root>
```

一键命令: `make gen-keys`

## 输入输出

| 角色 | 路径 |
|------|------|
| 输入 (规范) | `spec/keys.yaml` |
| Go SDK | `executor/vm/internal/keys/keys.go` |
| C++ SDK | `executor/deepx-core/include/deepx/key_defs.h` |
| TypeScript SDK | `tool/dashboard/frontend/src/api/keys.ts` |

## 设计

- 零网络依赖：yaml.v2 已在本地模块缓存
- 纯 Go 模板引擎（`text/template`），无外部模板语言
- `pattern` → 三语言表达式自动翻译:
  - Go: `fmt.Sprintf`
  - C++: `std::string` 拼接
  - TypeScript: 模板字面量

## 生成内容

- **前缀常量**: `const` / `constexpr` / 内联字面量
- **Key 构造函数**: 参数化函数，编译器保证类型安全
- **枚举类型**: vthread 状态等
- **常量**: `Backends[]`、`DeviceGPU0`、`ShmPrefix` 等
- **工具函数** (Go only): `IsRelative`、`ResolveRelative`、`FuncMain`

## keys.yaml 扩展指南

新增 key 定义只需在 `spec/keys.yaml` 的 `keys:` 下列出：

```yaml
- name: NewKey
  space: sys
  doc: "新 key 说明"
  params:
    - {name: id, type: string}
  pattern: "{sys_xxx}{id}"
```

然后运行 `make gen-keys`，三语言 SDK 自动包含 `NewKey(id)` 构造函数。

## 与 hardcode-audit 的关系

此工具直接解决 `doc/plans/hardcode-audit.md` 的 **根因 A**（80+ 处裸 Redis key 字符串），是审计报告 §5.1 方案的落地实现。
