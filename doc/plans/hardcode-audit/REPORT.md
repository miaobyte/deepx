# deepx executor 硬编码审计报告

> 审计目标: `heap-cpu` / `heap-metal` / `exop-cpu` / `exop-metal`
> 审计日期: 2026-05-01
> 审计范围: 所有 `.cpp`, `.hpp`, `.h`, `.mm`, `CMakeLists.txt`, `build.sh`, `Dockerfile`, `dockerbuild.sh`

---

## 一、总览

| 目录 | 源文件数 | 硬编码问题数 |
|------|---------|-------------|
| heap-cpu | 4 | 12 |
| heap-metal | 5 (含 main.mm) | 12 |
| exop-cpu | 23 | 16 |
| exop-metal | 18 | 15 |

共发现 **~55 处硬编码**，分为 7 大类。

---

## 二、分类详报

### 1. Redis 队列名 / Key 名硬编码 (13 处)

所有 main.cpp 的队列名、实例 Key、心跳 Key 均以 `static const` 硬编码在源码中，
**缺少命令行参数或配置文件支持**。不同进程仅靠后缀 `:0` 区分，无法多实例部署。

#### heap-cpu (`src/main.cpp:19-22`)
```cpp
static const char *HEAP_QUEUE    = "cmd:heap-cpu:0";
static const char *SYS_QUEUE     = "sys:cmd:heap-cpu:0";
static const char *INSTANCE_KEY  = "/sys/heap-plat/heap-cpu:0";
static const char *HEARTBEAT_KEY = "/sys/heartbeat/heap-cpu:0";
```

#### heap-metal (`src/main.cpp:20-23` + `src/main.mm:20-22`)
```cpp
static const char *HEAP_QUEUE    = "cmd:heap-metal:0";
static const char *SYS_QUEUE     = "sys:cmd:heap-metal:0";
static const char *INSTANCE_KEY  = "/sys/heap-plat/heap-metal:0";
static const char *HEARTBEAT_KEY = "/sys/heartbeat/heap-metal:0";
```

#### exop-cpu (`src/client/main.cpp:23-26`)
```cpp
static const char *OP_QUEUE       = "cmd:op-cpu:0";
static const char *SYS_QUEUE      = "sys:cmd:op-cpu:0";
static const char *INSTANCE_KEY   = "/sys/op-plat/op-cpu:0";
static const char *HEARTBEAT_KEY  = "/sys/heartbeat/op-cpu:0";
```
另有 `/op/op-cpu/list` 硬编码 7 次 (行 83-120)。

#### exop-metal (`src/client/main.cpp:24-28`)
```cpp
static const char *OP_QUEUE       = "cmd:exop-metal:0";
static const char *SYS_QUEUE      = "sys:cmd:exop-metal:0";
static const char *INSTANCE_KEY   = "/sys/op-plat/exop-metal:0";
static const char *HEARTBEAT_KEY  = "/sys/heartbeat/exop-metal:0";
static const char *OP_LIST_KEY    = "/op/exop-metal/list";
```

**建议**: 通过命令行参数 `--instance-id` 或环境变量 `DEEPX_INSTANCE_ID` 动态拼接。

---

### 2. Redis 连接地址/端口硬编码 (4 处)

所有 4 个项目的 `main()` 函数都将 Redis 默认地址硬编码为 `127.0.0.1:6379`:

```cpp
const char *redis_addr = "127.0.0.1";   // heap-cpu:236, heap-metal:245, exop-cpu:1042, exop-metal:719
int redis_port = 6379;                   // 同上
```

虽然支持 `argv[1]` / `argv[2]` 覆盖，但:
- 无 `--redis-addr` / `--redis-port` 长选项
- 无环境变量 fallback (如 `REDIS_ADDR`)
- 无配置文件支持

**建议**: 增加 `--redis-addr` / `--redis-port` CLI 参数 + `DEEPX_REDIS_ADDR` 环境变量。

---

### 3. 连接超时 / 心跳间隔 / 重试间隔硬编码 (8 处)

| 常量 | 值 | 位置 |
|------|-----|------|
| `BLOCK_TIMEOUT_SEC` | 5 | heap-cpu, heap-metal, exop-cpu, exop-metal (各1处) |
| `HEARTBEAT_INTERVAL_SEC` | 2 | 同上 4 处 |
| Redis connect timeout | `{2, 0}` (2秒) | 同上 4 处 |
| 重连 sleep | `sleep(1)` (1秒) | 同上 4 处 |

**建议**: 统一为可配置参数，通过 CLI 或环境变量设置。

---

### 4. device / node / instance 标识硬编码 (8 处)

#### 4.1 device 字段
所有 heap-* 注册信息中将 `device` 硬编码为 `"gpu0"`:
- `heap-cpu/main.cpp:68` — ⚠️ **但这是 CPU 进程，不应写 `gpu0`**
- `heap-metal/main.cpp:77` — Metal 进程，`gpu0` 勉强合理但不通用
- `exop-cpu/main.cpp:74` — `"cpu0"` ✅ 合理
- `exop-metal/main.cpp:93` — `"gpu0"` 勉强合理

#### 4.2 node 标识
heap-cpu 和 heap-metal 的 `handle_newtensor()` 将 `address.node` 硬编码为 `"n1"`:
- `heap-cpu/main.cpp:161`
- `heap-metal/main.cpp:170`
- `heap-metal/main.mm:158`

#### 4.3 address type
- 全部硬编码为 `"shm"` (5处)

#### 4.4 program 名称
- `heap-cpu/main.cpp:67` — `reg["program"] = "heap-cpu"` ❌ 硬编码
- `heap-metal/main.cpp:76` — 使用 `hostname-pid` 拼接 ✅ 较好

**建议**: device/node 应通过参数注入，program name 应统一使用 hostname-pid 模式。

---

### 5. 魔法数字 (Magic Numbers) 多处

| 位置 | 硬编码值 | 说明 |
|------|---------|------|
| `heap-cpu/main.cpp:157` | `cmd.element_count * 4` | ⚠️ **潜在 BUG**: byte_size 固定为 4字节(f32)，忽略实际 dtype。lifecycle.cpp 中有正确的 dtype→bytes 映射，但 main.cpp 在写 Redis meta 时未使用 |
| `heap-metal/main.cpp:166` | `cmd.element_count * 4` | ⚠️ 同上 BUG |
| `heap-metal/main.mm:154` | `cmd.element_count * 4` | ⚠️ 同上 BUG |
| `heap-cpu/lifecycle.cpp:49` | `int elem_bytes = 4` | 默认 f32 — 合理但可优化为查表 |
| `exop-cpu/main.cpp:148` | `fallback 16384` | 页大小 fallback 硬编码 |
| `exop-metal/main.cpp:167` | `fallback 16384` | 同上 |
| `exop-cpu/parallel.hpp:103` | `const int minblock = 256` | 并行粒度硬编码 |

---

### 6. CMake / 构建脚本中的绝对路径 (20+ 处)

#### 6.1 `/opt/homebrew/...` 路径 (macOS Homebrew 硬编码)

所有 CMakeLists.txt 都硬编码了 Homebrew 路径:
```cmake
include_directories(/opt/homebrew/opt/hiredis/include)
link_directories(/opt/homebrew/opt/hiredis/lib)
include_directories(/opt/homebrew/opt/nlohmann-json/include)
include_directories(/opt/homebrew/opt/openblas/include)
include_directories(/opt/homebrew/opt/highway/include)
```
**影响**: 无法在 Linux 或非 Homebrew 环境构建。

#### 6.2 `/usr/lib/x86_64-linux-gnu/...` 路径 (Linux 硬编码)
```cmake
# exop-cpu/CMakeLists.txt:28
list(APPEND CMAKE_PREFIX_PATH "/usr/lib/x86_64-linux-gnu/openblas-pthread/cmake")
```
**影响**: 非 x86_64 Linux (如 arm64) 找不到此路径。

#### 6.3 `/tmp/deepx/...` 构建输出路径
| 项目 | build.sh 行号 | 路径 |
|------|-------------|------|
| heap-cpu | :8 | `/tmp/deepx/heap-cpu/${PLAT}` |
| exop-cpu | :10 | `/tmp/deepx/exop-cpu/${PLAT}` |
| heap-metal | :3 | `/tmp/deepx/heap-metal` |
| exop-metal | :3 | `/tmp/deepx/exop-metal` |

**影响**: `/tmp` 在系统重启后可能被清空。

#### 6.4 运行时 registry 路径 (`/tmp/deepx_heap_registry.txt`)
- `heap-cpu/main.cpp:242`
- `heap-metal/main.cpp:251`
- `heap-metal/main.mm:238`

**建议**: 使用 `$XDG_RUNTIME_DIR` 或 `getenv("HOME")` + 相对路径。

---

### 7. 其他硬编码

#### 7.1 string 类型 dispatch (opcode / dtype / status)

所有 opcode 路由均使用 `if-else` 字符串比较:
- **exop-cpu/main.cpp**: 大量 `if (opcode == "add")` / `if (dtype == "f32")` 式字符串比对 (~200行)
- **exop-metal/main.cpp**: 同样，每个 dtype×opcode 组合手工编写 (~400行)

状态值全为裸字符串:
- `"ok"`, `"error"`, `"running"`, `"stopped"`, `"ready"`, `"deleted"`
- 错误码: `"HEAP_ERROR"`, `"OP_ERROR"`

#### 7.2 `"done:"` key 前缀 (5处)
```cpp
std::string key = "done:" + vtid;
```
全部 4 个项目的 notify_done 都使用此硬编码前缀。

#### 7.3 SHM 命名前缀 (2处)
`heap-cpu/lifecycle.cpp:37` + `heap-metal/lifecycle.cpp:37`:
```cpp
oss << "/deepx_t_" << std::hex << now << "_" << id;
```

#### 7.4 exop-cpu Dockerfile (`exop-cpu/Dockerfile`) — ⚠️ 完全过时
| 问题 | 行 | 说明 |
|------|-----|------|
| `FROM ...ubuntu:18.04` | :1 | Ubuntu 18.04 已 EOL |
| `ADD cpp-common cpp-common` | :20 | 路径 `cpp-common` 在当前仓库不存在 |
| `ADD op-cpu op-cpu` | :21 | 路径 `op-cpu` 在当前仓库不存在 |
| `WORKDIR /home/op-cpu` | :22 | 目录不存在 |
| `CMD ["./build/bin/deepx-executor-cpu"]` | :27 | 产物名与当前 CMake project `deepx-exop-cpu` 不一致 |
| `dockerbuild.sh:4` | — | `-f op-cpu/Dockerfile` — Dockerfile 路径名错误 |

#### 7.5 `exop-cpu/main.cpp:1002` — changeshape ops stub
```cpp
else if (opcode == "reshape" || opcode == "transpose" || opcode == "concat" ||
         opcode == "broadcastTo" || opcode == "repeat") {
    error = "changeshape ops not available (refactoring in progress)";
}
```
changeshape 算子未在此可执行文件中接入，始终返回错误。

---

### 8. heap-cpu 与 heap-metal 代码重复问题

`heap-cpu/src/` 与 `heap-metal/src/` 的 `.cpp/.h/.hpp` 文件完全一致:

| 文件 | 状态 |
|------|------|
| `lifecycle.cpp` | 逐字节相同 |
| `lifecycle.h` | 逐字节相同 |
| `registry_file.h` | 逐字节相同 |
| `main.cpp` | 仅日志前缀 `[heap-cpu]` vs `[heap]` 不同 |

heap-metal 另有 `main.mm` 为第三个副本。

---

## 三、按目录汇总

### heap-cpu (12 项)
1. 4 个 Redis Key/Queue 常量硬编码
2. Redis 默认地址 `127.0.0.1:6379`
3. `BLOCK_TIMEOUT_SEC=5`, `HEARTBEAT_INTERVAL_SEC=2`, connect timeout `{2,0}`, retry `sleep(1)`
4. `device="gpu0"` ⚠️ (CPU 进程写 gpu0)
5. `node="n1"` 硬编码
6. `meta["byte_size"] = cmd.element_count * 4` ⚠️ BUG
7. CMake Homebrew 绝对路径
8. `/tmp/deepx/heap-cpu/` + `/tmp/deepx_heap_registry.txt`
9. `/deepx_t_` SHM 前缀
10. `"done:"` key 前缀
11. `"ok"/"error"/"running"/"HEAP_ERROR"` 等裸字符串
12. 与 heap-metal 代码完全重复

### heap-metal (12 项)
与 heap-cpu 完全相同的 11 项 + `main.mm` 额外副本。差异仅:
- Queue Key 中 `heap-cpu` → `heap-metal`
- 日志前缀 `[heap-cpu]` → `[heap]`
- `register_instance` 使用 hostname 拼接 program name ✅

### exop-cpu (16 项)
1. 4 个 Redis Key/Queue + `/op/op-cpu/list` (7次引用)
2. Redis 默认地址 `127.0.0.1:6379`
3. 4 项超时/心跳/重试硬编码
4. `PROG_NAME = "op-cpu"` 硬编码
5. CMake: Homebrew 路径 + `/usr/lib/x86_64-linux-gnu/...`
6. CMake: SIMD flags `-mavx2 -msse4.2` / `-march=armv8.5-a`
7. `/tmp/deepx/exop-cpu/` 构建路径
8. 16384 page size fallback
9. `"done:"` key 前缀
10. ~200 行 if-else 字符串 opcode/dtype dispatch
11. changeshape ops stub ⚠️
12. Dockerfile 完全过时 ⚠️
13. `dockerbuild.sh` 引用不存在的路径
14. `"ok"/"error"/"running"/"OP_ERROR"` 裸字符串
15. `minblock = 256`

### exop-metal (15 项)
1. 5 个 Redis Key/Queue + `OP_LIST_KEY`
2. Redis 默认地址 `127.0.0.1:6379`
3. 4 项超时/心跳/重试硬编码
4. CMake: Homebrew 绝对路径
5. CMake: `CMAKE_OSX_DEPLOYMENT_TARGET "12.0"` + `if(NOT APPLE) FATAL_ERROR`
6. `/tmp/deepx/exop-metal/` 构建路径
7. `device="gpu0"` 硬编码
8. 16384 page size fallback
9. `"done:"` key 前缀
10. ~400 行 dtype×opcode 手工 dispatch
11. `"ok"/"error"/"running"/"OP_ERROR"` 裸字符串
12. 引用 io-metal plane 的硬编码 queue 名

---

## 四、建议优先级

### 🔴 高优先级 (功能性 BUG)
1. **`byte_size = element_count * 4`** — `heap-cpu/main.cpp:157` 和 `heap-metal/main.cpp:166`(及 `.mm:154`) 对非 f32 类型的 tensor 写入错误的 byte_size 到 Redis meta
2. **Dockerfile 路径全部过时** — `exop-cpu/Dockerfile` 无法构建
3. **changeshape ops stub** — `exop-cpu/main.cpp:1002` 所有 changeshape 算子返回 "not available"

### 🟡 中优先级 (可维护性 / 可移植性)
4. CMake 中所有 `/opt/homebrew/...` 绝对路径 → 用 `find_package` 替代
5. `/usr/lib/x86_64-linux-gnu/...` Linux 路径 → 条件编译或 `find_package`
6. Redis Key/Queue 无参数化 → 添加 `--instance-id` / `--redis-addr` 参数
7. heap-cpu 与 heap-metal 代码完全重复 → 抽取公共库

### 🟢 低优先级 (代码卫生)
8. `"done:"` / `"/deepx_t_"` 等魔法字符串前缀
9. `BLOCK_TIMEOUT_SEC` / `HEARTBEAT_INTERVAL_SEC` 等可配置化
10. opcode/dtype/status 字符串 → enum 化
