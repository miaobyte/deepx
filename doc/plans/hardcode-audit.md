# deepx 全项目硬编码审计汇总

> 日期: 2026-05-01 | 来源: 5 份组件审计报告合并

---

## 一、总体概览

| 组件 | 语言 | 硬编码数 | 最高严重度 |
|------|------|----------|-----------|
| deepx-core | C++ | 23 | 🔴 P0 |
| executor (heap/exop) | C++ | ~55 | 🔴 P0 |
| VM | Go | 80+ | 🔴 P0 |
| deepxctl | Go | 60+ | 🔴 P0 |
| dashboard | Go/TS/CSS | 50+ | 🔴 P0 |

**合计 ~270 处硬编码**，按类型分布：

| 类型 | 数量 | 典型问题 |
|------|------|----------|
| Redis Key/路径 | 80+ | `/vthread/`, `/src/func/`, `/sys/` 等字符串散落 |
| 超时/间隔/魔数 | 50+ | `200ms`, `5s`, `60`, `16384` 未命名 |
| 文件/目录路径 | 20+ | `/tmp/deepx/`, `/opt/homebrew/` 硬编码 |
| 网络地址/端口 | 10+ | `127.0.0.1:6379` vs `16379` 不一致 |
| 代码重复 | ~400 行 | parser↔ir, heap-cpu↔heap-metal |
| 字符串 dispatch | ~600 行 | if-else opcode/dtype 比对 |
| 其他 | 30+ | namespace 污染、include guard 不一致 |

---

## 二、跨组件共性问题

### 2.1 Redis Key 无集中定义（影响全部 5 个组件）

所有组件的 Redis key (`/vthread/`, `/src/func/`, `/sys/`, `/func/main`, `/op/*/list`, `cmd:`, `notify:vm`, `done:`) 均以字符串字面量散落。重构 key 格式时需改动 15+ 文件，极易遗漏。

### 2.2 Redis 默认端口不一致

| 组件 | 默认端口 |
|------|----------|
| VM server mode | `6379` |
| VM single mode | `6379` |
| loader | `16379` |
| deepxctl | `16379` |
| executor (C++) | `6379` |

### 2.3 超时值散落、无分组

各组件独立定义 `200ms`/`5s`/`30s` 等超时，调优需 grep 全仓库。

### 2.4 构建路径硬编码 `/tmp`

- CMake: `/opt/homebrew/opt/...` (macOS Homebrew 专用，Linux 不可编译)
- CMake: `/usr/lib/x86_64-linux-gnu/...` (x86_64 专用，arm64 不可编译)
- build.sh: `/tmp/deepx/<component>/` (重启可能被清空)
- deepxctl: `/tmp/deepx/...` 二进制路径与 Makefile 耦合

### 2.5 代码重复

| 重复内容 | 位置 | 行数 |
|----------|------|------|
| `parseInfix`/`parseParamList`/`stripQuotes` 等 | `parser/parser.go` ↔ `ir/instruction.go` | ~200 |
| `lifecycle.cpp`/`lifecycle.h`/`registry_file.h`/`main.cpp` | `heap-cpu/` ↔ `heap-metal/` | ~300 |
| Backend list `["op-metal","op-cuda","op-cpu"]` | `vm/vm.go` + `platform/router.go` | 2 处 |
| `isRelative` | `termio/native.go` + `platform/dispatch.go` | 2 处 |
| 服务列表 `{"op-metal","heap-metal","vm","dashboard"}` | `boot.go` + `shutdown.go` (4 处独立副本) | 4 处 |

---

## 三、各组件详情

### 3.1 deepx-core (C++ 公共库)

| # | 严重度 | 问题 | 文件 |
|---|--------|------|------|
| 1 | P0 | CMake `include_directories(/opt/homebrew/...)` — Linux 不可构建 | CMakeLists.txt:10 |
| 2 | P1 | `shm_tensor.cpp` 页大小 fallback `16384` + "Apple Silicon" 注释 | shm_tensor.cpp:15 |
| 3 | P1 | `shm_tensor.h` 注释提及 `MTLBuffer` — Metal API 泄漏到 POSIX 头文件 | shm_tensor.h:10 |
| 4 | P2 | `Shape::operator==` 参数名遮蔽成员，**始终返回 true** | shape.hpp:32 |
| 5 | P2 | include guard 命名不一致 (4 处: 缺少 DEEPX_ 前缀/旧名残留/多余下划线) | 4 个 .hpp |
| 6 | P2 | `#include "string"` / `#include "iostream"` — 标准库应用 `<>` | 2 处 |
| 7 | P2 | `toYaml()`/`fromYaml()` — 实际用 nlohmann/json，函数名残留 | shape.hpp:41 |
| 8 | P2 | `.data` 扩展名硬编码在头文件 | tensor.hpp:168 |
| 9 | P2 | `int tempidx = 0;` 死字段 | mem.hpp:20 |
| 10 | P3 | `using namespace std;` 在头文件中 (10 处) | 9 个 .hpp |

### 3.2 executor (heap-cpu / heap-metal / exop-cpu / exop-metal)

| # | 严重度 | 问题 | 影响面 |
|---|--------|------|--------|
| 1 | 🔴 BUG | `byte_size = element_count * 4` — 非 f32 类型写错误 byte_size | heap-cpu + heap-metal (3 处) |
| 2 | 🔴 BUG | exop-cpu `Dockerfile` 全部路径过时 (ubuntu:18.04 EOL, 目录不存在) | exop-cpu |
| 3 | 🔴 BUG | changeshape ops stub — 始终返回 "not available" | exop-cpu |
| 4 | 🔴 | 4 组 Redis Key/Queue 无参数化 (`cmd:heap-metal:0` 等) | 全部 4 个进程 |
| 5 | 🔴 | CMake 绝对路径: `/opt/homebrew/` + `/usr/lib/x86_64-linux-gnu/` | 全部 CMakeLists.txt |
| 6 | 🟡 | Redis 默认地址 `127.0.0.1:6379` — 无环境变量 fallback | 全部 4 个 main.cpp |
| 7 | 🟡 | 超时/心跳硬编码: `BLOCK_TIMEOUT=5`, `HEARTBEAT=2s`, `retry sleep(1)` | 各 4 份副本 |
| 8 | 🟡 | `device` 字段硬编码: CPU 进程写 `"gpu0"` ⚠️ | heap-cpu |
| 9 | 🟡 | `node="n1"` 硬编码 | heap-cpu + heap-metal |
| 10 | 🟡 | ~600 行 if-else 字符串 opcode/dtype dispatch | exop-cpu + exop-metal |
| 11 | 🟡 | heap-cpu ↔ heap-metal 代码完全重复 (lifecycle.cpp/.h, registry_file.h, main.cpp) | 2 个目录 |
| 12 | 🟢 | `"done:"` / `"/deepx_t_"` 魔法前缀 + `"ok"/"error"/"HEAP_ERROR"` 裸字符串 | 全部 |

### 3.3 VM (Go)

| # | 严重度 | 问题 | 影响面 |
|---|--------|------|--------|
| 1 | 🔴 | `KEYS /vthread/*` — 生产代码用 KEYS O(N) 全库扫描 | sched.go |
| 2 | 🔴 | `KEYS /vthread/<vtid>/*` — HandleReturn 扫描子栈 | codegen.go + cache.go |
| 3 | 🔴 | `KEYS /op/*/list` + `KEYS /sys/op-plat/*` — 每次 Select 扫描 | router.go |
| 4 | 🔴 | Redis key 前缀散落 15+ 文件 (`/vthread/`, `/src/func/`, `/sys/`) | 全局 |
| 5 | 🔴 | `cmd:heap-metal:0` — heap queue 硬编码实例 0，不支持多实例 | dispatch.go:207 |
| 6 | 🔴 | Redis 端口不一致: server `6379` vs loader `16379` | main.go 2 处 |
| 7 | 🟡 | ~200 行 parser↔ir 代码完全重复 | parser.go + instruction.go |
| 8 | 🟡 | 12 处 magic timeout 散落 (`2s` heartbeat, `30s` BLPOP, `5s` wait...) | 8 个文件 |
| 9 | 🟡 | 全局 WebSocket 连接池 + 文件句柄 map 无 TTL | term_ws.go |
| 10 | 🟢 | 状态字符串 `"init"/"running"/"wait"/"done"/"error"` 遍布 8 个文件 | 全局 |
| 11 | 🟢 | Backend list 重复定义 | vm.go + router.go |

### 3.4 deepxctl (Go CLI)

| # | 严重度 | 问题 |
|---|--------|------|
| 1 | 🔴 | PID 文件 `/tmp/deepx-boot.json` — 多用户冲突、重启丢失 |
| 2 | 🔴 | 日志目录 `/tmp/deepx/logs` — 多实例覆盖 |
| 3 | 🔴 | 二进制路径全部硬编码 `/tmp/deepx/...` — 与 Makefile 无一致性校验 |
| 4 | 🔴 | Dashboard dist 路径 — 非 repo root 启动 FATAL |
| 5 | 🔴 | `bash` + `make` 硬编码 — Alpine 无 bash |
| 6 | 🔴 | Redis DefaultAddr 与环境变量/flag 优先级不明确 |
| 7 | 🔴 | `splitRedisAddr` 回退值与 `DefaultAddr` 重复定义 |
| 8 | 🔴 | Dashboard 端口 `:8080` 不可配置 |
| 9 | 🔴 | Dashboard CORS `Allow-Origin: *` |
| 10 | 🟡 | 18+ Redis key 常量散落 (boot/run/shutdown) |
| 11 | 🟡 | 19+ 超时/间隔常量散落（`200ms` 出现 7 次） |
| 12 | 🟡 | 服务列表 `{"op-metal","heap-metal","vm","dashboard"}` 4 处独立副本 |
| 13 | 🟡 | `/func/main` 在 run.go 中出现 6 次 |
| 14 | 🟡 | Loader 输出解析依赖魔法字符串（`"OK"`, `"ENTRY /func/main →"`） |
| 15 | 🟡 | `/dev/null` 硬编码（应用 `os.DevNull`） |
| 16 | 🟢 | 版本号 `"0.2.0"` 硬编码、文件权限 `0644`/`0755` 散落 |

### 3.5 dashboard (Go + React/TypeScript/CSS)

| # | 严重度 | 问题 |
|---|--------|------|
| 1 | 🔴 | 超时 `60` 前后端各一份（run.go + CodeEditor.tsx） |
| 2 | 🔴 | 入口函数 `"main"` 前后端各一份（run.go + CodeEditor.tsx + OutputPanel.tsx） |
| 3 | 🔴 | `"/api/status/stream"` client.ts + App.tsx 重复 |
| 4 | 🔴 | `"/sys/heartbeat/term:dashboard"` read.go + terminal.go 重复 |
| 5 | 🔴 | 端口 `:8080` main.go + vite.config.ts 耦合 |
| 6 | 🟡 | xterm 主题 20 个颜色值 — 与 CSS 变量重复 |
| 7 | 🟡 | 终端 `"dashboard"` 实例名 + shell `/bin/bash` 硬编码 |
| 8 | 🟡 | 13 处魔数: 缓冲 `4096`/`256`/`8`、重连 `2000`/`3000`、轮询 `500`/`3000` |
| 9 | 🟡 | localStorage key 字符串散落 |
| 10 | 🟢 | 面板尺寸 `260`/`320`、字体 `13`、滚动 `5000` |

---

## 四、跨文件重复常量（前后端不一致风险）

| 值 | 出现位置 | 风险 |
|----|----------|------|
| `60` (默认超时秒) | `run.go` + `CodeEditor.tsx` | 一端改另一端行为异常 |
| `"main"` (默认入口) | `run.go` + `CodeEditor.tsx` + `OutputPanel.tsx` | 同上 |
| `"/api/status/stream"` | `client.ts` + `App.tsx` | 重复定义 |
| `"/sys/heartbeat/term:dashboard"` | `read.go` + `terminal.go` | 改键名一端 404 |
| `"127.0.0.1"` | `terminal.go` + `read.go` | 非本地不可达 |
| `:8080` | `main.go` + `vite.config.ts` | 端口变更改多处 |
| `16379` vs `6379` | VM ↔ loader ↔ executor | Redis 连接失败 |

---

## 五、修复优先级总表

### P0 — 阻塞性（立即修复）

| # | 问题 | 组件 | 影响 |
|---|------|------|------|
| 1 | `byte_size = element_count * 4` BUG | heap-cpu/metal | 非 f32 tensor 元信息错误 |
| 2 | `Shape::operator==` 始终返回 true | deepx-core | 逻辑错误 |
| 3 | Dockerfile 全部路径过时 | exop-cpu | 无法构建镜像 |
| 4 | CMake 绝对路径 `/opt/homebrew/` → `find_package` | 全部 C++ | Linux 不可编译 |
| 5 | `KEYS` 命令滥用 (5 处) → 确定性构造/SCAN | VM | 生产性能灾难 |
| 6 | `cmd:heap-metal:0` 硬编码实例 → 动态路由 | VM | 无法多实例 |

### P1 — 高优先（架构级改进）

| # | 问题 | 组件 |
|---|------|------|
| 7 | Redis Key 全项目常量化 → `config/keys.go` | 全部 |
| 8 | 消除 parser↔ir 200 行重复 → 提取 `grammar/` 包 | VM |
| 9 | 消除 heap-cpu↔heap-metal 300 行重复 → 公共库 | executor |
| 10 | 统一超时配置 → 集中 `config.go` + 环境变量覆盖 | 全部 |
| 11 | 统一 Redis 端口 (6379 vs 16379) → 环境变量 `REDIS_ADDR` | 全部 |
| 12 | deepxctl 二进制路径 → `DEEPX_OUT` 环境变量 | deepxctl |
| 13 | 前后端共享常量 (超时/入口函数) → 统一定义 | dashboard + deepxctl |
| 14 | Dashboard dist → `embed.FS` 嵌入二进制 | dashboard |

### P2 — 中优先（代码卫生）

| # | 问题 | 组件 |
|---|------|------|
| 15 | 状态字符串常量化 (`"init"/"running"/"done"/"error"`) | VM + executor |
| 16 | Backend list 统一 (`[]string{"op-metal",...}`) | VM |
| 17 | 魔法字符串 dispatch → enum/表驱动 | exop-cpu/metal |
| 18 | Loader 输出格式 → JSON 解析 | deepxctl + loader |
| 19 | xterm 主题从 CSS 变量动态读取 | dashboard |
| 20 | WebSocket 连接池 TTL 清理 | VM (term_ws.go) |
| 21 | `using namespace std;` 移除 | deepx-core |
| 22 | include guard 命名统一 | deepx-core |

### P3 — 低优先（锦上添花）

| # | 问题 | 组件 |
|---|------|------|
| 23 | 版本号 → ldflags 注入 | deepxctl |
| 24 | 文件权限 `0644`/`0755` 常量化 | deepxctl |
| 25 | `/dev/null` → `os.DevNull` | deepxctl |
| 26 | `toYaml()` → `toJson()` 重命名 | deepx-core |
| 27 | `#include "string"` → `<string>` | deepx-core |
| 28 | 魔数面板尺寸/字体常量化 | dashboard |
| 29 | localStorage key 常量化 | dashboard |

---

## 六、修复预估

| 阶段 | 内容 | 预计人天 |
|------|------|----------|
| P0 | 6 项阻塞性修复 | 3d |
| P1 | 8 项架构改进 | 5d |
| P2 | 8 项代码卫生 | 3d |
| P3 | 7 项锦上添花 | 1d |
| **合计** | | **12 人天** |
