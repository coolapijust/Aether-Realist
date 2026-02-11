# Aether-Realist 协议定义（V5.1，当前实现）

## 1. 传输层

- 承载：WebTransport over HTTP/3
- Record 是协议最小封装单位
- 每条双向流首包必须是 `Metadata Record (0x01)`

## 2. Record 格式

统一结构：

- `LengthPrefix`：4 字节（不含自身）
- `Header`：30 字节
- `Payload`：变长
- `Padding`：变长

Header 字段（Big Endian）：

- `Version(u8)`：当前为 `0x05`
- `Type(u8)`：`Metadata/Data/Ping/Pong/Error`
- `TimestampNano(u64)`
- `PayloadLength(u32)`
- `PaddingLength(u32)`
- `SessionID(4B)`
- `Counter(u64)`

## 3. 类型定义

- `0x01` Metadata Record
- `0x02` Data Record
- `0x03` Ping Record
- `0x04` Pong Record
- `0x7F` Error Record

## 4. 加密与密钥派生

### 4.1 Metadata 加密

- 算法：`AES-128-GCM`
- Key 派生：`HKDF-SHA256(psk, salt=SessionID, info="aether-realist-v5")`
- Nonce：`SessionID(4B) || Counter(8B)`
- AAD：完整 30B Header
- Tag：16 字节

### 4.2 Data Record

当前实现不对 Data payload 做 AEAD，仅做协议封装。

- Data padding：`0`
- Metadata padding：随机（握手混淆）

## 5. 防重放

接收端校验：

1. 时间戳窗口（默认 ±30s）
2. Counter 单调递增（按流维护）

不满足即判定无效流量，进入失败处理路径。

## 6. 分片与吞吐

- 最大 Record 限制：`1MB`
- Data payload 上限：`16KB`（`MaxRecordPayload`）
- 写路径按 `16KB` 分片封装，可降低弱网 HoL 惩罚

## 7. 会话与轮换

- 每个会话有独立 `SessionID + Counter` 生成器
- Counter 到阈值（`2^32`）需 rekey（轮换会话）
- 客户端支持定时轮换与异常重建

## 8. 失败行为（当前实现）

握手失败时，服务端采用统一失败策略：

- 记录安全日志
- 随机短延迟
- 写入随机诱饵字节
- 关闭流

该行为用于降低探测方对失败原因的可观测性。

