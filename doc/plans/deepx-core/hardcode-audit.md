# deepx-core Hardcode 审计报告

> 日期: 2026-05-01 | 扫描范围: `executor/deepx-core/`

## 严重程度定义

| 级别 | 含义 |
|------|------|
| **P0** | 阻塞跨平台构建 / 运行时错误 |
| **P1** | 平台假设泄漏，违反架构铁律 |
| **P2** | 代码质量问题（命名不一致、潜在 bug、死代码） |
| **P3** | 风格问题（namespace 污染等） |

---

## P0 — 硬编码路径（阻塞跨平台构建）

### 1. CMakeLists.txt — nlohmann-json include 路径硬编码

| 项目 | 值 |
|------|-----|
| **文件** | `executor/deepx-core/CMakeLists.txt:10` |
| **内容** | `include_directories(/opt/homebrew/opt/nlohmann-json/include)` |
| **问题** | macOS Homebrew 专用路径，Linux/Windows 无法编译 |
| **影响面** | 所有依赖 deepx-core 的 executor 间接继承此路径要求 |
| **建议** | 使用 `find_package(nlohmann_json REQUIRED)` 或 CMake `FetchContent`；或作为 cmake cache 变量由顶层项目传入 |

---

## P1 — 平台假设泄漏

### 2. shm_tensor.cpp — Apple Silicon 硬编码注释 + 魔数

| 项目 | 值 |
|------|-----|
| **文件** | `src/shmem/shm_tensor.cpp:15` |
| **内容** | `if (ps <= 0) ps = 16384; // Apple Silicon default` |
| **问题** | ① `16384` (16KB) 作为 fallback 页大小，注释明确写 "Apple Silicon" — 但此文件应平台无关；② x86 Linux 页大小通常是 4096，默认值 16384 会导致内存浪费 |
| **建议** | fallback 改为 `4096`（POSIX 最小页大小），移除 Apple 注释 |

### 3. shm_tensor.h — Apple UMA 注释

| 项目 | 值 |
|------|-----|
| **文件** | `include/deepx/shm_tensor.h:10-11` |
| **内容** | `On Apple Silicon (UMA), this memory is directly GPU-accessible via MTLBuffer` |
| **问题** | POSIX 共享内存头文件不应提及 Metal/Apple 特定 API |
| **建议** | 删除或改为通用注释（"On UMA architectures, this memory may be GPU-accessible"） |

---

## P2 — 代码质量问题

### 4. shape.hpp — operator== 有 bug

| 项目 | 值 |
|------|-----|
| **文件** | `include/deepx/shape.hpp:32` |
| **内容** | `bool operator==(const Shape &shape) const { return shape.shape == shape.shape; }` |
| **问题** | 参数名 `shape` 遮蔽成员 `this->shape`，导致两侧都读取参数的 `shape` 成员，**始终返回 true** |
| **修复** | `return this->shape == shape.shape && this->dtype == shape.dtype;`（参数改名或加 this 前缀） |

### 5. include guard 命名不一致（4 处）

| 文件 | 当前 guard | 问题 |
|------|-----------|------|
| `include/deepx/tensor.hpp` | `TENSOR_HPP` | 缺少 `DEEPX_` 前缀，与其他 28 个 guard 不一致 |
| `include/deepx/shape_reduce.hpp` | `DEEPX_SHAPE_SUM_HPP` | 文件名为 `shape_reduce`，guard 却写 `SHAPE_SUM` |
| `include/deepx/typespec.hpp` | `DEEPX_DTYPE_TYPEDEF_HPP` | 类名是 `TypeSpec`，guard 却写 `TYPEDEF`（旧名残留） |
| `include/stdutil/print.hpp` | `STDUTIL_PRINT_HPP__` | 尾部多一个 `_`（`HPP__` vs `HPP`） |

### 6. include 指令写法不一致

| 文件 | 内容 | 问题 |
|------|------|------|
| `include/tensorfunc/authors.hpp:4` | `#include "string"` | 标准库头文件应用 `<>` 而非 `""` |
| `include/mem/mem.hpp:8` | `#include "iostream"` | 同上 |
| `include/stdutil/fs.hpp` | guard 前缀 `DEEPX_STDUTIL_` | 其他 stdutil 用 `STDUTIL_`（无 DEEPX_ 前缀）|

### 7. shape.cpp — YAML 命名残留

| 项目 | 值 |
|------|-----|
| **文件** | `include/deepx/shape.hpp:41-42` |
| **内容** | `std::string toYaml() const;` / `void fromYaml(const std::string &yaml);` |
| **问题** | 实际已使用 nlohmann/json 序列化，函数名仍叫 `*Yaml`，形参名也叫 `yaml` |
| **建议** | 重命名为 `toJson()` / `fromJson()` |

### 8. tensor.hpp — 硬编码文件扩展名

| 项目 | 值 |
|------|-----|
| **文件** | `include/deepx/tensor.hpp:168` |
| **内容** | `saver(data, shape.size, path+".data");` |
| **问题** | `.data` 扩展名硬编码在库头文件中 |

### 9. mem.hpp — 死代码

| 项目 | 值 |
|------|-----|
| **文件** | `include/mem/mem.hpp:20` |
| **内容** | `int tempidx = 0;` |
| **问题** | 声明但从未使用，死字段 |

---

## P3 — Namespace 污染

### 10. `using namespace std;` 在头文件中（10 处）

所有以下头文件在其 namespace 内 `using namespace std;`，导致任何包含这些头文件的翻译单元被强制污染 std 命名空间：

| 文件 |
|------|
| `include/deepx/data_category.hpp` |
| `include/deepx/shape_changeshape.hpp` |
| `include/deepx/tensor.hpp` |
| `include/deepx/vector_combination.hpp` |
| `include/mem/mem.hpp` |
| `include/stdutil/fs.hpp` |
| `include/tensorfunc/changeshape.hpp` |
| `include/tensorfunc/authors.hpp` |
| `include/tf/tf.hpp` |

另有 2 个 `.cpp` 文件中存在同样问题（影响较小）。

---

## 汇总

| 级别 | 数量 | 说明 |
|------|------|------|
| P0 | 1 | CMake 路径硬编码，阻塞 Linux/Windows 构建 |
| P1 | 2 | Apple Silicon 注释/魔数泄漏到平台无关代码 |
| P2 | 9 | include guard 不一致(4) + 写法错误(2) + 命名残留(1) + 扩展名硬编码(1) + 死代码(1) + operator==bug(1) |
| P3 | 10 | `using namespace std;` 污染头文件 |

---

*生成方式: `grep -rn` 全量扫描 + 人工逐文件审查*
