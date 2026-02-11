# Aether-Realist 设计文档（当前实现）

## 1. 组件与边界

当前工程分为三层：

- `cmd/aether-gateway`：服务端网关
- `cmd/aetherd` + `internal/core`：本地核心代理与控制面
- `gui/`：Tauri + React 桌面端

数据面和控制面分离：

- 数据面：`User App -> SOCKS5/HTTP Proxy(aetherd) -> WebTransport -> Gateway -> Target`
- 控制面：`GUI -> HTTP API(/api/v1/*) + WebSocket(/api/v1/events) -> aetherd`

## 2. 协议设计要点

### 2.1 Zero-Sync V5

- Header 长度 30 字节
- Nonce：`SessionID(4B) + Counter(8B)`
- Metadata 使用 AES-128-GCM，Header 作为 AAD
- 防重放：时间窗口 + 计数器单调检查

### 2.2 Record 与分片

- Data Record 最大 payload：`16KB`（`MaxRecordPayload`）
- Data Record padding：当前实现为 `0`
- Metadata padding：随机范围（用于握手混淆）

## 3. 性能策略

### 3.1 Flow Control 分级

通过 `WINDOW_PROFILE` 做链路分层：

- `conservative`：低指纹、低吞吐上限
- `normal`：默认平衡配置
- `aggressive`：高 BDP 场景配置（最高 `32MB/48MB` 窗口上限）

### 3.2 UDP 缓冲

客户端和网关都会尝试设置 `32MB` UDP 读写缓冲，降低 burst 丢包风险。

### 3.3 HoL 缓解

网关在 `TCP -> WebTransport` 方向会按 `16KB` 分片后再封装，避免大记录在弱网中放大队头阻塞惩罚。

## 4. 安全策略

- 强制 TLS 1.3
- QUIC ALPN 仅 `h3`
- 非协议流量走 decoy 页面或通用响应
- 握手失败采用统一失败路径（随机延迟 + 诱饵输出 + 关闭流）

## 5. Core 运行模型

`aetherd` 内部包含：

- 状态机（Idle/Starting/Active/Rotating/Closing/Closed/Error）
- Session manager（拨号、重连、轮换）
- SOCKS5 + HTTP 代理入口
- 规则引擎（`proxy/direct/block/reject`）
- 指标采集与事件总线

系统代理能力通过 `internal/systemproxy` 实现，默认优先使用 HTTP 代理端口进行系统级接管。

## 6. 控制 API 设计

`internal/api/server.go` 提供：

- REST：状态、配置、规则、流、指标、控制命令
- WebSocket：实时事件推送与心跳（`ping/pong`）

GUI 是 API 的消费者，不参与协议细节实现。

