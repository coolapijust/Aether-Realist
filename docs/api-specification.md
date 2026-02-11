# Aether-Realist API 规范（aetherd）

默认监听：`127.0.0.1:9880`

- REST Base: `http://127.0.0.1:9880/api/v1`
- Events: `ws://127.0.0.1:9880/api/v1/events`

## 1. REST API

### 1.1 状态

#### `GET /status`

返回核心状态与摘要信息。

示例：

```json
{
  "state": "Active",
  "config": {},
  "active_streams": 2,
  "proxy_enabled": true,
  "rules_count": 3,
  "last_error": ""
}
```

### 1.2 配置

#### `GET /config`

读取当前配置（`SessionConfig`）。

#### `POST /config`

更新配置并持久化。常用字段：

- `url`
- `psk`
- `listen_addr`
- `http_proxy_addr`
- `dial_addr`
- `max_padding`
- `allow_insecure`
- `bypass_cn`
- `block_ads`
- `window_profile` (`conservative` / `normal` / `aggressive`)
- `rotation`
- `rules`

成功返回：

```json
{"status":"updated"}
```

### 1.3 规则

#### `GET /rules`
获取规则列表。

#### `POST /rules`
整体更新规则列表。

### 1.4 运行控制

#### `POST /control/start`
启动 Core（使用当前配置）。

#### `POST /control/stop`
停止 Core。

#### `POST /control/rotate`
触发会话轮换。

#### `POST /control/proxy`

切换系统代理：

```json
{"enabled": true}
```

响应：

```json
{"status":"success","enabled":true}
```

### 1.5 观测接口

#### `GET /streams`
返回活动流列表。

#### `GET /metrics`
返回当前指标快照（等价于一次 `metrics.snapshot` 事件）。

## 2. WebSocket 事件流

连接地址：`/api/v1/events`

服务端推送事件对象（JSON），典型类型：

- `core.stateChanged`
- `session.established`
- `session.rotating`
- `session.closed`
- `stream.opened`
- `stream.closed`
- `core.error`
- `metrics.snapshot`
- `rotation.scheduled`
- `rotation.prewarm.started`
- `rotation.completed`
- `app.log`

### 2.1 客户端心跳

客户端可发送：

```json
{"action":"ping"}
```

服务端响应：

```json
{"type":"pong"}
```

## 3. 错误语义

- 非法方法：`405 Method Not Allowed`
- 请求体错误：`400 Bad Request`
- Core 操作失败：`500 Internal Server Error`

建议 GUI/自动化调用按 HTTP 状态码处理，不依赖响应文本做逻辑判断。

