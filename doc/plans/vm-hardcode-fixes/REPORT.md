# VM Hardcode 检测报告

> 扫描范围: `executor/vm/**/*.go` (19 个非测试源文件)
> 生成日期: 2026-05-01

---

## 总览

| 类别 | 数量 | 严重程度 |
|------|------|----------|
| Redis Key Prefix 硬编码 | 35+ | 🔴 高 |
| Magic Timeout 硬编码 | 12 | 🟡 中 |
| Redis KEYS 命令滥用 | 5 | 🔴 高 |
| 硬编码 Backend List | 3 (重复) | 🟡 中 |
| 硬编码 Queue/Channel 名称 | 6 | 🟡 中 |
| 代码重复 (DRY 违规) | ~200 行 | 🟡 中 |
| 硬编码字符串常量 (status) | 15+ | 🟢 低 |
| 全局可变状态（无 TTL） | 2 | 🟡 中 |
| 缺失集中配置 | 全局 | 🔴 高 |

---

## 1. Redis Key Prefix 硬编码

**问题**: 整个 VM 代码库中，Redis key 前缀 `"/vthread/"`、`"/src/func/"`、`"/sys/"` 等以字符串字面量散落在 15+ 个文件中，无常量定义。重构 key 格式时需要修改所有位置，极易遗漏。

### `/vthread/` — 出现在 7 个文件、20+ 处

| 文件 | 行号 | 代码 |
|------|------|------|
| `state/state.go` | 24,44,59,67,71,73,76 | `"/vthread/"+vtid` |
| `vm/control.go` | 26 | `"/vthread/"+vtid+"/"+condKey[2:]` |
| `codegen/codegen.go` | 53,112,118,124 | `"/vthread/%s/%s/"` |
| `ir/instruction.go` | 25 | `"/vthread/%s/%s"` |
| `sched/sched.go` | 17,23 | `"/vthread/*"` |
| `termio/native.go` | 108,115,140 | `"/vthread/"+vtid+"/"` |
| `platform/dispatch.go` | 54 | `"/vthread/"+vtid+"/"` |
| `cache/cache.go` | 21 | `"/vthread/%s/%s/"` |
| `cmd/vm/main.go` | 295,333 | `"/vthread/%d"` |

### `/src/func/` — 出现在 3 个文件

| 文件 | 行号 | 代码 |
|------|------|------|
| `ast/ast.go` | 47,51 | `"/src/func/"+f.Name` |
| `codegen/codegen.go` | 26,50 | `"/src/func/"+funcName` |
| `vm/vm.go` | 111 | `"/src/func/"+opcode` |

### `/sys/` 和 `/op/` — 出现在 5 个文件

| 文件 | 行号 | Key |
|------|------|-----|
| `cmd/vm/main.go` | 85,101 | `"/sys/vm/"`, `"/sys/heartbeat/vm:%s"` |
| `cmd/vm/main.go` | 287 | `"/sys/vtid_counter"` |
| `platform/router.go` | 18,49 | `"/op/*/list"`, `"/sys/op-plat/*"` |
| `termio/native.go` | 119 | `"/sys/term/"` |
| `cmd/vm/main.go` | 204,254 | `"/op/buildin/list"`, `"/func/main"` |

### 修复建议

在 `internal/config/` 下统一定义所有 Redis key 常量和构造器函数：

```go
package config

const (
    PrefixVThread    = "/vthread/"
    PrefixSrcFunc    = "/src/func/"
    PrefixSys        = "/sys/"
    PrefixOp         = "/op/"
    KeyFuncMain      = "/func/main"
    KeyOpBuildinList = "/op/buildin/list"
    KeyVtidCounter   = "/sys/vtid_counter"
    KeyNotifyVM      = "notify:vm"
)

func VThreadState(vtid string) string    { return PrefixVThread + vtid }
func VThreadData(vtid, key string) string { return PrefixVThread + vtid + "/" + key }
func SrcFunc(name string) string         { return PrefixSrcFunc + name }
func SysVM(vmID string) string           { return PrefixSys + "vm/" + vmID }
func SysHeartbeat(vmID string) string    { return PrefixSys + "heartbeat/vm:" + vmID }
func OpFunc(backend, name string) string { return PrefixOp + backend + "/func/" + name }
```

---

## 2. Magic Timeout 硬编码

**问题**: 超时值直接写死在代码中，无统一入口，不同环境（开发/生产）无法调整。

| 文件 | 行号 | 值 | 用途 |
|------|------|-----|------|
| `cmd/vm/main.go` | 63-65 | `10s / 3s / 3s` | Redis 连接池超时 |
| `cmd/vm/main.go` | 104 | `2s` | 心跳上报间隔 |
| `cmd/vm/main.go` | 123 | `1s` | `/func/main` 轮询间隔 |
| `cmd/vm/main.go` | 141 | `5s` | sys command BLPOP 超时 |
| `cmd/vm/main.go` | 181 | `3s` | shutdown 超时 |
| `sched/sched.go` | 61 | `5s` | Wait BLPOP 超时 |
| `platform/dispatch.go` | 188 | `30s` | Compute BLPOP 超时 |
| `platform/dispatch.go` | 211 | `5s` | Lifecycle BLPOP 超时 |
| `termio/term_ws.go` | 51 | `3s` | WebSocket ping 超时 |
| `termio/term_ws.go` | 89 | `30s` | WebSocket read 超时 |
| `main.go` (测试) | 55,77,79 | `100ms / 50ms / 10s` | 调试主程 |
| `cmd/vm/main.go` | 232 | `3s` | singleRun 等待 |

---

## 3. Redis KEYS 命令滥用 🔴

**问题**: 生产代码使用 `KEYS pattern` 扫描 Redis。该命令 O(N) 遍历全库，阻塞其他操作。

| 文件 | 行号 | 命令 | 影响 |
|------|------|------|------|
| `sched/sched.go` | 17 | `KEYS /vthread/*` | 每次 Pick 全量扫描 vthread |
| `codegen/codegen.go` | 124 | `KEYS /vthread/<vtid>/<pc>/*` | HandleReturn 扫描子栈 |
| `platform/router.go` | 18 | `KEYS /op/*/list` | Select 扫描所有 op 后端 |
| `platform/router.go` | 49 | `KEYS /sys/op-plat/*` | Select 扫描所有实例 |
| `cache/cache.go` | 23 | `KEYS /vthread/<vtid>/<pc>/*` | LoadSubStack 扫描子栈 |

### 修复建议

- `sched.go`: 移除轮询，完全依赖 `notify:vm` 队列驱动
- `codegen.go` + `cache.go`: 子栈 key 遵循固定格式 `[addr0, addr1]`，可直接构造无需 KEYS
- `router.go`: 改用 Redis Set 维护 op-plat 注册表，新增/删除实例时原子更新

---

## 4. 硬编码 Backend List（两处重复）

| 文件 | 行号 | 代码 |
|------|------|------|
| `vm/vm.go` | 115 | `[]string{"op-metal", "op-cuda", "op-cpu"}` |
| `platform/router.go` | 99 | `[]string{"op-metal", "op-cuda", "op-cpu"}` |
| `platform/router.go` | 109 | `return "op-metal"` (硬编码默认后端) |

---

## 5. 硬编码 Queue / Channel 名称

| 文件 | 行号 | 值 | 用途 |
|------|------|-----|------|
| `main.go` | 66 | `"notify:vm"` | worker 唤醒 |
| `cmd/vm/main.go` | 138 | `"sys:cmd:vm:%s"` | 系统命令队列 |
| `cmd/vm/main.go` | 327 | `"notify:vm"` | worker 唤醒 |
| `sched/sched.go` | 61 | `"notify:vm"` | BLPOP 等待 |
| `platform/dispatch.go` | 181 | `"cmd:op-%s"` | op 任务队列 |
| `platform/dispatch.go` | 207 | `"cmd:heap-metal:0"` | **heap 队列硬编码实例 0!** |

### 🔴 特别注意: `cmd:heap-metal:0`

`dispatch.go:207` 硬编码了实例 ID:
```go
rdb.RPush(ctx, "cmd:heap-metal:0", taskJSON)
```

heap 任务固定发送到 `heap-metal:0`，不支持多实例负载均衡。应与 `Compute()` 一样使用动态路由。

---

## 6. 代码重复 (DRY 违规)

### 6.1 `parser/parser.go` ↔ `ir/instruction.go` 重复

以下函数在两个文件中**完全重复**（共约 200 行）：

| 函数 | parser.go 行号 | instruction.go 行号 | 重复行数 |
|------|---------------|---------------------|----------|
| `parseInfix` | 331-363 | 266-309 | ~33 |
| `parseParamList` | 365-410 | 311-358 | ~44 |
| `parseParamListRaw` | 412-457 | 361-406 | ~45 |
| `stripQuotes` | 459-466 | 409-419 | ~8 |
| `isKeyRef` | 468-470 | 437-439 | ~3 |
| `isString` | 473-475 | 442-444 | ~3 |
| `findArrow` | 322-329 | 252-262 | ~8 |

**建议**: 提取到 `internal/grammar/` 公共包。

### 6.2 `isRelative` 重复

| 文件 | 行号 | 可见性 |
|------|------|--------|
| `termio/native.go` | 102-104 | 私有 `isRelative` |
| `platform/dispatch.go` | 46-48 | 导出 `IsRelative` |

相同逻辑，一个导出、一个未导出。

---

## 7. 状态字符串硬编码

Status 值 `"init"`, `"running"`, `"wait"`, `"done"`, `"error"` 遍布 8 个文件，应常量化。

---

## 8. 全局可变状态

`termio/term_ws.go` 中两个全局 map 无 TTL 清理:

```go
var conns   = map[string]*wsConn{}   // WebSocket 连接池 → 可能内存泄漏
var fileWriters = map[string]*os.File{} // 文件句柄池 → 可能 fd 泄漏
```

---

## 9. Redis 地址不一致

| 文件 | 默认地址 |
|------|----------|
| `main.go` | `127.0.0.1:16379` |
| `cmd/vm/main.go` (server) | `127.0.0.1:6379` |
| `cmd/vm/main.go` (single) | `127.0.0.1:6379` |
| `cmd/loader/main.go` | `127.0.0.1:16379` |

VM server 使用 6379，loader 使用 16379 — 不一致！

---

## 10. 缺失统一配置模块

**核心问题**: VM 子系统无集中配置。所有参数散落在 19 个文件中，无环境变量覆盖，无配置验证。

### 建议

创建 `executor/vm/internal/config/config.go`，聚合所有 VM 配置，通过环境变量 + 默认值加载：

```go
type Config struct {
    Redis    RedisConfig
    Timeouts TimeoutConfig
    Keys     KeyConfig
}

func Load() *Config { /* 环境变量 + 默认值 */ }
```

---

## 修复优先级

| 优先级 | 类别 | 影响 | 预计耗时 |
|--------|------|------|----------|
| **P0** | Key 前缀常量化 | 杜绝拼写错误，可维护性 | 1d |
| **P0** | 代码重复消除 (parser ↔ ir) | 减少 ~200 行，单一真相源 | 1d |
| **P1** | KEYS → 确定性构造 / SCAN | 生产性能 | 0.5d |
| **P1** | 统一超时配置 | 可运维性 | 0.5d |
| **P1** | heap queue 动态路由 | 多实例支持 | 0.5d |
| **P2** | 状态常量化 | 类型安全，IDE 补全 | 0.5d |
| **P2** | Backend list 统一 | 扩展性 | 0.5d |
| **P2** | WS 连接池 TTL | 资源泄漏 | 0.5d |
| **P3** | 集中配置模块 `config.go` | 整体可配置性 | 2d |
| **P3** | Redis 地址统一 | 消除 6379/16379 不一致 | 0.5d |

**预计总耗时: 6-7 人天**
