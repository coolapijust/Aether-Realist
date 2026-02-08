# Aether-Realist API 规范指南

Aether-Realist Core (`aetherd`) 提供了一套基于 HTTP (REST) 和 WebSocket 的本地控制 API，允许 GUI 客户端或自动化脚本管理连接、监控指标并实时接收系统事件。

默认情况下，API 服务器监听在 `127.0.0.1:9880`。

---

## 1. REST API (v1)

所有 API 响应均采用标准 JSON 格式。

### 核心状态管理

#### `GET /api/v1/status`
获取当前 Core 的运行状态。
- **响应示例**:
  ```json
  {
    "state": "Active",
    "config": { ... },
    "active_streams": 12,
    "proxy_enabled": true,
    "rules_count": 5,
    "last_error": ""
  }
  ```

#### `POST /api/v1/control/start`
使用当前配置启动核心服务。
- **请求体**: 空
- **响应**: `{"status": "started"}`

#### `POST /api/v1/control/stop`
停止核心服务。
- **响应**: `{"status": "stopped"}`

#### `POST /api/v1/control/rotate`
手动触发会话轮换（Reconnect）。
- **响应**: `{"status": "rotating"}`

---

### 配置与规则

#### `GET /api/v1/config`
读取当前生效的持久化配置。

#### `POST /api/v1/config`
更新并保存配置到磁盘。
- **请求体参数**:
  - `url`: Gateway 连接串 (如 `https://host:port/path`)
  - `psk`: 预共享密钥
  - `listen_addr`: SOCKS5 监听地址
  - `http_proxy_addr`: HTTP 代理监听地址
  - `allow_insecure`: 是否跳过证书验证 (Boolean)
  - `bypass_cn`: 国内直连 (Boolean)
  - `block_ads`: 广告拦截 (Boolean)

#### `POST /api/v1/control/proxy`
快捷开启/关闭系统代理。
- **请求体**: `{"enabled": true}`

---

### 监控与诊断

#### `GET /api/v1/metrics`
获取当前的流量统计快照。
- **响应字段**: `sessionUptime`, `bytesSent`, `bytesReceived`, `latencyMs` 等。

#### `GET /api/v1/streams`
列出所有活跃的 TCP 流详情。

---

## 2. WebSocket 事件流 (v1)

### 连接地址
`ws://127.0.0.1:9880/api/v1/events`

### 订阅机制
建立连接后，服务端会主动推送 Core 产生的所有事件。

### 事件类型示例

#### 核心状态变更 (`core.stateChanged`)
```json
{
  "type": "core.stateChanged",
  "from": "Starting",
  "to": "Active",
  "timestamp": 1707312345
}
```

#### 指标快照 (`metrics.snapshot`)
每秒推送一次实时流量数据。
```json
{
  "type": "metrics.snapshot",
  "bytesSent": 1048576,
  "bytesReceived": 5242880,
  "activeStreams": 3,
  "timestamp": 1707312350
}
```

#### 系统日志 (`app.log`)
```json
{
  "type": "app.log",
  "level": "info",
  "message": "Connected to upstream gateway",
  "source": "core"
}
```

### 客户端指令
可以通过 WebSocket 发送心跳：
- **发送**: `{"action": "ping"}`
- **接收**: `{"type": "pong"}`
