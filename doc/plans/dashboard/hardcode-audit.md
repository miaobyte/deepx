# dashboard 硬编码审计

> 扫描范围：`tool/dashboard/` 下所有 `.go`、`.ts`、`.tsx`、`.css`（排除 `node_modules`、`dist`）

---

## 后端 (Go)

### 1. `main.go`

| # | 位置 | 硬编码值 | 类型 | 建议 |
|---|------|----------|------|------|
| 1.1 | `flag.String("addr", ":8080", ...)` | `:8080` | 端口 | 已通过 flag 覆盖，默认值可接受 |
| 1.2 | `flag.String("loader", "/tmp/deepx/vm/loader", ...)` | `/tmp/deepx/vm/loader` | 路径 | ⚠️ 与 Makefile 定义的 `/tmp/deepx/vm/loader` 一致，但 hardcode 在源码中。deepxctl boot 传递 `-loader` 参数覆盖，但独立启动 dashboard 时会用此默认值 |

### 2. `internal/server/server.go`

| # | 位置 | 硬编码值 | 类型 | 建议 |
|---|------|----------|------|------|
| 2.1 | `findDistDir()` 候选路径 | `"frontend/dist"` | 路径 | 开发/部署双路径探测已覆盖，可接受 |
| 2.2 | `findDistDir()` 候选路径 | `"tool/dashboard/frontend/dist"` | 路径 | 同上 |

### 3. `internal/redis/read.go`

| # | 位置 | 硬编码值 | 类型 | 建议 |
|---|------|----------|------|------|
| 3.1 | `Connect()` | `PoolSize: 8, MinIdleConns: 2` | 连接池 | 可接受，小规模部署 |
| 3.2 | `GetSystemStatus()` — 心跳 key 匹配 | `"/sys/heartbeat/term:dashboard"` | KV 路径 | ⚠️ 终端心跳 key 硬编码，term 实例 ID 不固定时可能错 |
| 3.3 | `GetSystemStatus()` — 算子扫描 | `"/op/*/list"` | KV 路径 | 协议约定，可接受 |
| 3.4 | `GetSystemStatus()` — vthread 扫描 | `"/vthread/*"` | KV 路径 | 协议约定，可接受 |
| 3.5 | `SrcFuncNames()` | `"/src/func/*"` | KV 路径 | 协议约定，可接受 |

### 4. `internal/redis/scan.go`

| # | 位置 | 硬编码值 | 类型 | 建议 |
|---|------|----------|------|------|
| 4.1 | `ScanOpPlatInstances()` | `"/sys/op-plat/*"` | KV 路径 | 协议约定，可接受 |
| 4.2 | `ScanHeapPlatInstances()` | `"/sys/heap-plat/*"` | KV 路径 | 协议约定，可接受 |
| 4.3 | `ScanVMInstance()` | `"/sys/vm/*"` | KV 路径 | 协议约定，可接受 |
| 4.4 | `ScanTermInstance()` | `"/sys/term/*"` | KV 路径 | 协议约定，可接受 |
| 4.5 | `HeartbeatKeyFromInstance()` — sysType 匹配 | `"op-plat"`, `"heap-plat"`, `"vm"` | 字符串 | ⚠️ 硬编码类型名，新增组件需改代码 |

### 5. `internal/redis/write.go`

| # | 位置 | 硬编码值 | 类型 | 建议 |
|---|------|----------|------|------|
| 5.1 | `RunVThread()` — timeout | `5*time.Second` | 超时 | ⚠️ 魔数，应可配置 |
| 5.2 | `RunVThread()` — 连接池 | `PoolSize: 1, MinIdleConns: 0` | 连接池 | 可接受（短期写入连接） |
| 5.3 | `RunVThread()` — vtid 计数器 key | `"/sys/vtid_counter"` | KV 路径 | 协议约定，可接受 |
| 5.4 | `RunVThread()` — vthread key 模板 | `"/vthread/%d"` | KV 路径 | 协议约定，可接受 |
| 5.5 | `RunVThread()` — 初始状态 | `{"pc":"[0,0]","status":"init"}` | JSON | 协议约定，可接受 |
| 5.6 | `RunVThread()` — ret 槽位 | `"./ret"` | 字符串 | 协议约定，可接受 |
| 5.7 | `RunVThread()` — 通知队列 | `"notify:vm"` | KV 路径 | 协议约定，可接受 |

### 6. `internal/handler/terminal.go`

| # | 位置 | 硬编码值 | 类型 | 建议 |
|---|------|----------|------|------|
| 6.1 | `DefaultTermRows` / `DefaultTermCols` | `24` / `80` | 终端尺寸 | 可接受（行业标准） |
| 6.2 | `TermWriterType` | `"websocket"` | 类型标识 | 协议约定，可接受 |
| 6.3 | `RegisterTerminal()` — 流名称 | `"stdout"`, `"stderr"`, `"stdin"` | 字符串 | 协议约定，可接受 |
| 6.4 | `RegisterTerminal()` — 终端实例 key | `"/sys/term/dashboard"` | KV 路径 | ⚠️ 实例名为 `dashboard` 硬编码 |
| 6.5 | `RegisterTerminal()` — I/O key 模板 | `"/sys/term/dashboard/"` + stream | KV 路径 | 同上 |
| 6.6 | `RegisterTerminal()` — 心跳 key | `"/sys/heartbeat/term:dashboard"` | KV 路径 | ⚠️ 与 read.go 3.2 重复 hardcode |
| 6.7 | `RegisterTerminal()` — 心跳间隔 | `3 * time.Second` | 超时 | 可接受 |
| 6.8 | `RegisterTerminal()` — 心跳 TTL | `10 * time.Second` | 超时 | 可接受 |
| 6.9 | `RegisterTerminal()` — host 默认 | `"127.0.0.1"` | IP | ⚠️ 绑定地址探测，`:8080` 时默认为 `127.0.0.1` |
| 6.10 | `upgrader` — 缓冲区 | `ReadBufferSize: 1024, WriteBufferSize: 1024` | 缓冲区 | 可接受 |
| 6.11 | `ServeTerminal()` — shell 默认 | `"/bin/bash"` | 路径 | ⚠️ 非 Linux 环境（macOS 有 `/bin/bash`，但其他系统可能不同） |
| 6.12 | `ServeTerminal()` — 环境变量 | `"TERM=xterm-256color"`, `"COLORTERM=truecolor"` | 字符串 | 可接受 |
| 6.13 | `ServeTerminal()` — PTY 缓冲区 | `4096` | 缓冲区 | ⚠️ 魔数 |
| 6.14 | `addBrowser()` — 通道缓冲区 | `256` | 缓冲区 | ⚠️ 魔数 |
| 6.15 | `ServeTermStderr()` — ANSI 红 | `\x1b[31m`, `\x1b[0m` | 终端转义 | 可接受（标准 ANSI） |

### 7. `internal/handler/run.go`

| # | 位置 | 硬编码值 | 类型 | 建议 |
|---|------|----------|------|------|
| 7.1 | `Run()` — 默认超时 | `req.Timeout = 60` | 秒 | ⚠️ 魔数，前端也有一份 60 |
| 7.2 | `execLoader()` — loader 超时 | `10 * time.Second` | 超时 | ⚠️ 魔数 |
| 7.3 | `execLoader()` — 输出解析 | `"/src/func/"`, `"OK"`, `"ENTRY /func/main →"` | 字符串 | 协议约定，可接受 |
| 7.4 | `detectEntry()` — 默认入口 | `"main"` | 字符串 | 协议约定，可接受 |

### 8. `internal/handler/status.go`

| # | 位置 | 硬编码值 | 类型 | 建议 |
|---|------|----------|------|------|
| 8.1 | `ServeSSE()` — 通道缓冲 | `8` | 缓冲区 | ⚠️ 魔数 |
| 8.2 | `Run()` — SSE 推送间隔 | `2 * time.Second` | 超时 | ⚠️ 魔数 |
| 8.3 | `ListVThreads()` — 响应 key | `"vthreads"` | JSON key | 协议约定，可接受 |
| 8.4 | `ListFunctions()` — 响应 key | `"functions"` | JSON key | 协议约定，可接受 |

### 9. `internal/handler/ops.go`

| # | 位置 | 硬编码值 | 类型 | 建议 |
|---|------|----------|------|------|
| 9.1 | `GetOps()` — Redis key 模板 | `"/op/" + backend + "/list"` | KV 路径 | 协议约定，可接受 |

---

## 前端 (React/TypeScript)

### 10. `src/api/client.ts`

| # | 位置 | 硬编码值 | 类型 | 建议 |
|---|------|----------|------|------|
| 10.1 | `BASE` | `''` | 空字符串 | 可接受（同源请求） |
| 10.2 | 所有 fetch URL | `/api/status`, `/api/vthreads`, `/api/vthread/`, `/api/functions`, `/api/ops/`, `/api/run` | 路径 | 协议约定，可接受 |
| 10.3 | `sseUrl()` | `'/api/status/stream'` | 路径 | 同上 |
| 10.4 | `termWsUrl()` | `'/api/terminal'` | 路径 | 同上 |

### 11. `src/App.tsx`

| # | 位置 | 硬编码值 | 类型 | 建议 |
|---|------|----------|------|------|
| 11.1 | 面板宽度 | `initialSize: 260` (left), `initialSize: 320` (right) | px | ⚠️ 魔数 |
| 11.2 | 面板最小/最大 | `minSize: 200, maxSize: 400` (left) / `minSize: 240, maxSize: 500` (right) | px | ⚠️ 魔数 |
| 11.3 | SSE URL | `'/api/status/stream'` | 路径 | ⚠️ 与 client.ts 10.3 重复定义 |

### 12. `src/components/CodeEditor.tsx`

| # | 位置 | 硬编码值 | 类型 | 建议 |
|---|------|----------|------|------|
| 12.1 | `DEFAULT_CODE` | 7 行 dxlang 示例代码 | 字符串 | 可接受（默认模板） |
| 12.2 | localStorage key | `'dx-editor-source'` | 字符串 | ⚠️ 应提取为常量 |
| 12.3 | 默认入口函数 | `'main'` | 字符串 | ⚠️ 与 run.go 7.4 重复 |
| 12.4 | 默认超时 | `60` | 秒 | ⚠️ 与 run.go 7.1 重复 |
| 12.5 | Tab 缩进宽度 | `'    '` (4空格) | 字符串 | 可接受 |
| 12.6 | 自动检测入口正则 | `/def\s+(\w+)/g` | 正则 | 可接受 |

### 13. `src/components/OutputPanel.tsx`

| # | 位置 | 硬编码值 | 类型 | 建议 |
|---|------|----------|------|------|
| 13.1 | localStorage key | `'dx-run-history'` | 字符串 | ⚠️ 应提取为常量 |
| 13.2 | 历史记录上限 | `50` | 条数 | ⚠️ 魔数 |
| 13.3 | vthread 轮询间隔 | `3000` | ms | ⚠️ 魔数 |
| 13.4 | 默认入口显示 | `entry: 'main'` | 字符串 | ⚠️ 与 12.3/7.4 重复 |

### 14. `src/components/Terminal.tsx`

| # | 位置 | 硬编码值 | 类型 | 建议 |
|---|------|----------|------|------|
| 14.1 | xterm 配置 — `fontSize` | `13` | px | ⚠️ 魔数 |
| 14.2 | xterm 配置 — `scrollback` | `5000` | 行 | ⚠️ 魔数 |
| 14.3 | xterm 配置 — `tabStopWidth` | `4` | 空格 | 可接受 |
| 14.4 | xterm 主题 — 全部 20 个颜色值 | `#0d1117`, `#e6edf3`, ... | 颜色 | ⚠️ CSS 变量有相同值，应引用 CSS 变量 |
| 14.5 | `fontFamily` | `'SF Mono', 'Fira Code', ...` | 字体 | ⚠️ CSS 变量 `--font-mono` 有相同值，重复定义 |
| 14.6 | WebSocket 重连间隔 | `2000` | ms | ⚠️ 魔数 |
| 14.7 | ResizeObserver 防抖 | `100` | ms | ⚠️ 魔数 |
| 14.8 | tab 切换后 fit 延迟 | `50` | ms | ⚠️ 魔数 |

### 15. `src/hooks/useVThreadPoll.ts`

| # | 位置 | 硬编码值 | 类型 | 建议 |
|---|------|----------|------|------|
| 15.1 | 成功轮询间隔 | `500` | ms | ⚠️ 魔数 |
| 15.2 | 错误重试间隔 | `2000` | ms | ⚠️ 魔数 |

### 16. `src/hooks/useWebSocket.ts`

| # | 位置 | 硬编码值 | 类型 | 建议 |
|---|------|----------|------|------|
| 16.1 | SSE 重连间隔 | `3000` | ms | ⚠️ 魔数 |

### 17. `src/App.css`

| # | 位置 | 硬编码值 | 类型 | 建议 |
|---|------|----------|------|------|
| 17.1 | `--bg-primary` … `--accent-purple` | 全部 12 个 CSS 变量值 | 颜色/字体 | 可接受（主题定义） |
| 17.2 | `scrollbar` 宽度 | `6px` | px | 可接受 |

### 18. `vite.config.ts`

| # | 位置 | 硬编码值 | 类型 | 建议 |
|---|------|----------|------|------|
| 18.1 | dev server port | `5173` | 端口 | 可接受（Vite 默认） |
| 18.2 | proxy target | `http://localhost:8080` | URL | ⚠️ 与 main.go 1.1 端口耦合 |
| 18.3 | ws proxy target | `ws://localhost:8080` | URL | ⚠️ 同上 |
| 18.4 | build outDir | `'dist'` | 路径 | 可接受（Vite 默认） |

---

## 跨文件重复 hardcode

| 值 | 出现位置 | 风险 |
|----|----------|------|
| `60` (默认超时秒数) | `run.go:7.1`, `CodeEditor.tsx:12.4` | 前后端不一致时行为异常 |
| `"main"` (默认入口函数) | `run.go:7.4`, `CodeEditor.tsx:12.3`, `OutputPanel.tsx:13.4` | 同上 |
| `"/api/status/stream"` | `client.ts:10.3`, `App.tsx:11.3` | 重复定义 |
| `"/sys/heartbeat/term:dashboard"` | `read.go:3.2`, `terminal.go:6.6` | 一端改键名另一端 404 |
| `"127.0.0.1"` | `terminal.go:6.9`, `read.go:DefaultAddr` | 非本地部署时不可达 |
| `:8080` 端口 | `main.go:1.1`, `vite.config.ts:18.2/18.3` | 端口变更需改多处 |

---

## 风险等级汇总

| 等级 | 数量 | 典型问题 |
|------|------|----------|
| 🔴 高风险 | 6 | 跨文件重复常量（前后端超时/入口/端口），单点修改导致不一致 |
| 🟡 中风险 | 12 | 魔数（超时/缓冲/间隔），不可配置，调优需改源码 |
| 🟢 低风险 | ~30 | KV 路径、协议约定，符合架构规范，可接受 |

## 建议优先修复

1. **提取共享常量文件** — 超时 `60`、入口函数 `"main"`、心跳间隔等，前后端统一引用
2. **xterm 主题从 CSS 变量读取** — `Terminal.tsx` 的 `theme` 对象改由 `getComputedStyle` 动态获取，消除 20 个颜色重复
3. **心跳 key 模板化** — `HeartbeatKeyFromInstance` 不支持 `term` 类型，导致 `/sys/heartbeat/term:dashboard` 被散落 hardcode
4. **魔数提取为包级常量** — `4096`(PTY buf)、`256`(channel buf)、`8`(SSE buf)、`2000/3000`(重连间隔)
