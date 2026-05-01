# deepxctl Hardcode 审计报告

> 日期: 2026-05-01
> 范围: `tool/deepxctl/` + 相关文件 (`tool/dashboard/internal/server/server.go`)

---

## 总览

| 类别 | 数量 | 严重程度 |
|------|------|----------|
| 文件/目录路径 | 8 | 🔴 高 |
| 网络地址/端口 | 5 | 🔴 高 |
| Redis Key 常量 | 18+ | 🟡 中 |
| 超时/间隔常量 | 19+ | 🟡 中 |
| 服务名称 | 4 | 🟢 低 |
| 构建脚本路径 | 4 | 🟡 中 |
| 权限/其他魔法值 | 6 | 🟢 低 |

---

## 🔴 高严重度 — 文件/目录路径

### 1. PID 文件路径 — 硬编码 `/tmp`

**文件**: `cmd/boot.go:21`
```go
const BootPIDFile = "/tmp/deepx-boot.json"
```
**问题**: macOS `/tmp` 是符号链接到 `/private/tmp`，重启后可能被清理；多用户环境冲突；Linux 上 `systemd-tmpfiles` 可能清理。
**建议**: 使用 `os.TempDir()` 或通过环境变量/flag 可配置。

### 2. 日志目录 — 硬编码 `/tmp`

**文件**: `cmd/boot.go:91`
```go
mgr.SetLogDir("/tmp/deepx/logs")
```
**问题**: 与 PID 文件相同问题。无用户隔离，无跨平台考虑。
**建议**: 从 `Boo

# 2. 日志目录 — 硬编码 `/tmp`

**文件**: `cmd/boot.go:91`
```go
mgr.SetLogDir("/tmp/deepx/logs")
```
**问题**: 与 PID 文件相同问题。无用户隔离，多实例日志互相覆盖。
**建议**: 通过环境变量 `DEEPXCTL_LOG_DIR` 配置，默认 `os.TempDir()/deepx/logs`。

### 3. 二进制产物路径 — 全部硬编码 `/tmp/deepx/`

**文件**: `internal/builder/builder.go:25-29`
```go
var (
    OpMetal   = "/tmp/deepx/exop-metal/deepx-exop-metal"
    HeapMetal = "/tmp/deepx/heap-metal/deepx-heap-metal"
    VM        = "/tmp/deepx/vm/vm"
    Loader    = "/tmp/deepx/vm/loader"
    Dashboard = "/tmp/deepx/dashboard/dash-server"
)
```
**问题**: 与 Makefile 中 `DEEPX_OUT := /tmp/deepx` 同步但无一致性校验。如果 Makefile 修改了输出目录，deepxctl 会找不到二进制
**建议**: 统一到常量配置，或通过 flag `--output-dir` / 环境变量 `DEEPX_OUT` 覆盖。

### 4. Dashboard 前端资源路径 — 硬编码相对路径

**文件**: `dashboard/internal/server/server.go:146-153`
```go
for _, d := range []string{
    filepath.Join("frontend", "dist"),
    filepath.Join("tool", "dashboard", "frontend", "dist"),
} {
```
**问题**: 这些路径只在从 repo root 运行时有效。从其他目录启动 dashboard 会找不到 `index.html` 进而 FATAL。
**建议**: 在构建时将 `dist/` 嵌入二进制（`embed.FS`），或通过 flag `--dist-dir` 指定。

### 5. 临时 term 目录 — 使用默认 TempDir

**文件**: `cmd/run.go:420`
```go
dir, err := os.MkdirTemp("", "deepxctl-run-*")
```
**问题**: 虽然使用 `os.TempDir()` 作为基础，但未提供自定义覆盖。
**建议**: 轻微问题，可通过 `DEEPXCTL_TMPDIR` 环境变量支持自定义。

### 6. `/dev/null` — 硬编码设备路径 (for stdin term writer)

**文件**: `cmd/run.go:443`
```go
pipe.HSet(ctx, "/sys/term/deepxctlrun/stdin", "type", "file", "detail", "/dev/null")
```
**问题**: `/dev/null` 在 Unix 上可用，但在 Windows 上无效。且第 98 行已使用 `os.DevNull`，此处不一致。
**建议**: 用 `os.DevNull` 替代。

### 7. 硬编码 shell / 构建命令

**文件**: `internal/builder/builder.go:96,119`
```go
cmd := exec.Command("bash", c.script)
cmd := exec.Command("make", "build-dashboard")
```
**问题**:
- `bash` 在 Alpine 等最小化发行版中可能不存在（只有 `sh`）。脚本 `.sh` 包含 shebang，应直接执行。
- `make` + `"build-dashboard"` 目标名称硬编码，若 Makefile 重命名则失败。
**建议**: 用 `exec.Command(c.script)` 依赖 shebang；make 目标名称可配置。

---

## 🔴 高严重度 — 网络地址/端口

### 8. Redis 默认地址

**文件**: `internal/redis/redis.go:29`
```go
const DefaultAddr = "127.0.0.1:16379"
```
**问题**: 硬编码端口 `16379`，与其他 Redis 实例冲突风险。虽然有 flag `--redis` 可覆盖，但无环境变量回退。
**建议**: 支持 `DEEPXCTL_REDIS_ADDR` 环境变量，优先级: flag > env > default。

### 9. Dashboard 监听地址 — 硬编码 `:8080`

**文件**: `cmd/boot.go:134`
```go
"-addr", ":8080",
```
**问题**: 端口 8080 是常见开发端口，容易冲突。Dashboard 完全无法配置监听地址。
**建议**: 添加 `--dashboard-addr` flag + `DEEPXCTL_DASHBOARD_ADDR` 环境变量。

### 10. splitRedisAddr 回退值与 DefaultAddr 重复

**文件**: `cmd/common.go:39-40`
```go
host = "127.0.0.1"
port = "16379"
```
**问题**: 回退值与 `redis.DefaultAddr = "127.0.0.1:16379"` 重复定义。修改 DefaultAddr 时容易遗漏此处，导致不一致行为。
**建议**: 直接从 `redis.DefaultAddr` 拆分，或使用 `net.SplitHostPort`。

### 11. Redis 连接池参数硬编码

**文件**: `internal/redis/redis.go:38-39`
```go
PoolSize:    4,
MinIdleConns: 1,
```
**问题**: 硬编码不适合所有场景。高并发时 PoolSize=4 可能成为瓶颈。
**建议**: 可配置或使用 go-redis 默认值 (`PoolSize: 10 * runtime.GOMAXPROCS(0)`)。

### 12. Dashboard CORS 策略 — `Allow-Origin: *`

**文件**: `dashboard/internal/server/server.go:178`
```go
w.Header().Set("Access-Control-Allow-Origin", "*")
```
**问题**: 生产环境中存在安全风险。
**建议**: 开发环境使用 `*`，生产环境通过 flag 限制。

---

## 🟡 中严重度 — Redis Key 常量 (18+ 处)

### 问题描述
整个代码库中有 18+ 处硬编码的 Redis key 字符串，分布在 `boot.go`、`run.go`、`shutdown.go` 中。这些 key 是 deepx 系统的协议约定，但如果系统 key 命名规范发生变化，需要修改多处，容易遗漏。

### 完整清单

| Redis Key | 文件:行 | 用途 |
|-----------|---------|------|
| `/sys/op-plat/exop-metal:0` | boot.go:100, run.go:141 | Op-metal 实例 |
| `/sys/heap-plat/heap-metal:0` | boot.go:113, run.go:142 | Heap-metal 实例 |
| `/sys/vm/0` | boot.go:126, run.go:143 | VM 实例 |
| `/sys/term/dashboard` | boot.go:143, run.go:144 | Dashboard 终端 |
| `/sys/term/deepxctlrun/stdout` | run.go:441,475 | 终端 stdout writer |
| `/sys/term/deepxctlrun/stderr` | run.go:442,476 | 终端 stderr writer |
| `/sys/term/deepxctlrun/stdin` | run.go:443,477 | 终端 stdin writer |
| `/func/main` | run.go:308 (多处) | VM 入口函数 |
| `/src/func/` | run.go:275 | 函数注册前缀（解析用） |
| `sys:cmd:op-metal:0` | shutdown.go:92 | Op-metal 命令队列 |
| `sys:cmd:heap-metal:0` | shutdown.go:92 | Heap-metal 命令队列 |
| `sys:cmd:vm:0` | shutdown.go:123 | VM 命令队列 |
| `/sys/heartbeat/op-metal:0` | shutdown.go:92,150 | Op-metal 心跳 |
| `/sys/heartbeat/heap-metal:0` | shutdown.go:93,151 | Heap-metal 心跳 |
| `/sys/heartbeat/vm:0` | shutdown.go:123,152 | VM 心跳 |

### 服务列表重复
服务列表 `{"op-metal", "heap-metal", "vm", "dashboard"}` 在以下位置各定义了一次独立副本:
- `boot.go:24-28` — `BootState` 结构体字段
- `shutdown.go:179-187` — Phase 4 force check
- `shutdown.go:196-204` — Phase 4 force kill
- `shutdown.go:273-278` — `forceKill` 函数

### 同一 key 字符串多次出现
`/func/main` 在 `run.go` 中出现 **6 次**（定义处 + 每次引用处）。

**建议**: 集中定义为常量包 `internal/keys/`，所有服务列表用 `[]ServiceInfo` 统一定义。

---

## 🟡 中严重度 — 超时与间隔常量 (19+ 处)

### 超时值

| 值 | 文件:行 | 用途 |
|----|---------|------|
| `5 * time.Second` | boot.go:96,102,109,114,121,127 | StopAll 超时 (启动失败时) |
| `30 * time.Second` | boot.go:100,113,126 | 实例注册等待超时 |
| `10 * time.Second` | boot.go:143 | Dashboard 注册等待超时 |
| `10 * time.Second` | shutdown.go:111,137 | 心跳停止等待超时 |
| `3 * time.Second` | redis.go:33,52 | Redis PING / FLUSHDB 超时 |
| `60` (int seconds) | run.go:77 | 默认执行超时 |
| `5 * time.Minute` | run.go:212 | 回退执行超时(当 timeout=0) |
| `5 * time.Second` | run.go:147 | 服务验证等待超时 |
| `5 * time.Second` | shutdown.go:287 | SIGTERM 等待超时 |
| `500 * time.Millisecond` | shutdown.go:173 | 关闭宽限期 |

### 轮询间隔

| 值 | 文件:行 | 用途 |
|----|---------|------|
| `200 * time.Millisecond` | redis.go:75,80,87 | WaitForInstance 轮询间隔 |
| `200 * time.Millisecond` | run.go:316,324,336 | /func/main 轮询间隔 |
| `200 * time.Millisecond` | shutdown.go:306 | waitPID 轮询间隔 |
| `200 * time.Millisecond` | shutdown.go:293 | forceKill SIGKILL 延迟 |
| `100 * time.Millisecond` | run.go:344 | vthread 状态轮询间隔 |
| `100 * time.Millisecond` | shutdown.go:208 | force kill 操作间延迟 |
| `300 * time.Millisecond` | shutdown.go:264 | waitHeartbeats 轮询间隔 |

**问题**: 这些值分散在代码中，调优需要理解上下文。`200ms` 出现了 7 次，但含义各异。
**建议**: 按功能分组为具名常量，如:
```go
const (
    pollFast  = 100 * time.Millisecond
    pollNormal = 200 * time.Millisecond
    pollSlow  = 300 * time.Millisecond
    startupWait   = 30 * time.Second
    shutdownWait  = 10 * time.Second
    stopWait      = 5 * time.Second
)
```

---

## 🟡 中严重度 — 构建脚本路径

**文件**: `internal/builder/builder.go:40-45`
```go
func DefaultScripts(repoRoot string) Scripts {
    return Scripts{
        OpMetal:   filepath.Join(repoRoot, "executor/exop-metal/build.sh"),
        HeapMetal: filepath.Join(repoRoot, "executor/heap-metal/build.sh"),
        VM:        filepath.Join(repoRoot, "executor/vm/build.sh"),
    }
}
```
**问题**: 如果 executor 目录结构变化，需修改代码。
**建议**: 通过配置文件或环境变量 `DEEPXCTL_BUILD_SCRIPT_OP` 等可覆盖。

---

## 🟢 低严重度 — 其他魔法值

### 版本号

**文件**: `main.go:13`
```go
var version = "0.2.0"
```
**建议**: 构建时通过 `-ldflags "-X main.version=$VERSION"` 注入。

### 文件权限

| 值 | 文件:行 |
|----|---------|
| `0644` | boot.go:176, process/manager.go:97, run.go:429,433 |
| `0755` | process/manager.go:93 |

**建议**: 使用 `fs.FileMode` 具名常量统一管理。

### Loader 输出解析魔法字符串

**文件**: `cmd/run.go:275-283`
```go
"/src/func/"           // 函数路径前缀
"OK"                   // 成功标识
"ENTRY /func/main →"   // 入口点标识
```
**问题**: 与 loader 二进制强耦合。loader 输出格式变更会导致静默失败。
**建议**: loader 输出 JSON 格式的解析结果。

### Term writer 协议字段

**文件**: `cmd/run.go:441-443`
```go
"type"    // → "file"
"detail"  // → 文件路径
```
**问题**: 这是 VM 与 deepxctl 之间的协议约定，硬编码字符串散布。
**建议**: 定义为协议常量。

### `index.html` 重复

**文件**: `dashboard/internal/server/server.go`
- 行 141: `filepath.Join(dir, "index.html")`
- 行 149-150: `filepath.Join(d, "index.html")`
- 行 153: `filepath.Join(d, "index.html")`

**建议**: 定义为常量 `const indexPath = "index.html"`，或使用 `embed.FS`。

---

## 修复建议优先级

### P0 (高优先 — 影响可移植性和可维护性)
1. **统一 output 路径** (问题 3): 让 deepxctl 和 Makefile 通过 `DEEPX_OUT` 环境变量共享输出路径
2. **提取所有 Redis Key 常量**: 建立 `internal/keys/` 包集中定义
3. **Dashboard 端口可配置** (问题 9): 添加 `--dashboard-addr` flag
4. **统一 DefaultAddr 引用** (问题 10): `splitRedisAddr` 使用 `net.SplitHostPort(redis.DefaultAddr)`

### P1 (中优先 — 提升代码质量和健壮性)
5. **所有超时值**: 按功能分组为具名常量
6. **嵌入前端产物** (问题 4): 使用 `go:embed`
7. **消除服务列表重复**: 定义 `var AllServices = []ServiceInfo{...}`
8. **Loader 输出格式** (问题 16): 改为 JSON 解析

### P2 (低优先 — 锦上添花)
9. **版本号** (问题 14): 从 VCS tag / ldflags 获取
10. **构建脚本路径**: 环境变量覆盖
11. **文件权限**: 统一定义常量
12. **`/dev/null`** (问题 6): 统一使用 `os.DevNull`

---

## 受影响文件汇总

| 文件 | 硬编码数量 | 最高严重度 |
|------|-----------|-----------|
| `cmd/run.go` | 28 | 🟡 中 |
| `cmd/shutdown.go` | 20 | 🟡 中 |
| `cmd/boot.go` | 16 | 🔴 高 |
| `internal/builder/builder.go` | 12 | 🔴 高 |
| `internal/redis/redis.go` | 6 | 🔴 高 |
| `cmd/common.go` | 3 | 🔴 高 |
| `dashboard/internal/server/server.go` | 8 | 🟡 中 |
| `main.go` | 1 | 🟢 低 |
| `internal/process/manager.go` | 3 | 🟢 低 |
