# 部署与环境配置

## 1. Aether Gateway (Docker - 推荐)

Aether-Realist 提供了独立的 Docker 镜像，支持 HTTP/3 WebTransport。

### 启动命令

> [!TIP]
> **自动 TLS**：如果未提供 `-cert` 和 `-key` 参数，服务端将自动生成内存中的 **自签名证书**。这非常适合测试环境或处于反向代理（如 Caddy）后端的情况。

```bash
# 方式 A：使用自签名证书（快速测试）
docker run -d \
  --name aether-gateway \
  -p 4433:4433/udp \
  -p 4433:4433/tcp \
  ghcr.io/coolapijust/aether-rea:latest \
  -psk "your-strong-password"

# 方式 B：使用自有证书（生产推荐）
docker run -d \
  --name aether-gateway \
  -v /path/to/certs:/certs \
  -p 4433:4433/udp \
  -p 4433:4433/tcp \
  ghcr.io/coolapijust/aether-rea:latest \
  -cert /certs/fullchain.pem \
  -key /certs/privkey.pem \
  -psk "your-strong-password"
```

### 必需参数
- `-cert`: TLS 证书文件路径。
- `-key`: TLS 私钥文件路径。
- `-psk`: 预共享密钥，客户端连接时需一致。

---

## 2. 反向代理与 TLS 自动化 (Caddy)

如果您希望拥有自动化的公网证书（由 Let's Encrypt 提供），推荐使用 Caddy。

### 方案 A：Caddy 获取证书 + Gateway 直接挂载 (推荐)

让 Caddy 负责证书更新，Gateway 直接读取 Caddy 生成的文件。

**Caddyfile:**
```caddy
your-domain.com {
    # 仅用于获取证书，不执行转发
    tls your-email@example.com
}
```

**Docker Compose 示例:**
```yaml
services:
  gateway:
    image: ghcr.io/coolapijust/aether-rea:latest
    ports:
      - "443:4433/udp"
    volumes:
      - caddy_data:/certs:ro
    command: -cert /certs/caddy/certificates/acme-v02.api.letsencrypt.org-directory/your-domain.com/your-domain.com.crt -key /certs/caddy/certificates/acme-v02.api.letsencrypt.org-directory/your-domain.com/your-domain.com.key -psk "your-psk"

volumes:
  caddy_data:
    external: true
```

### 方案 B：Caddy 反向代理 (HTTP/3 转发)

> [!WARNING]
> WebTransport 转发对反向代理的要求较高，请确保 Caddy 版本 >= 2.7。

**Caddyfile:**
```caddy
your-domain.com {
    reverse_proxy https://gateway:4433 {
        header_up Host {host}
        header_up X-Real-IP {remote_host}
        transport http {
            versions h3
            tls_insecure_skip_verify # 允许 Gateway 使用自签名证书
        }
    }
}
```

---

## 3. Cloudflare Worker 配置 (Legacy)

### Wrangler 配置

在项目根目录添加 `wrangler.toml`，并启用 `nodejs_compat`：

```toml
name = "aether-realist"
main = "src/worker.js"
compatibility_date = "2024-02-06"
nodejs_compat = true

[vars]
SECRET_PATH = "/v1/api/sync"
```

- `SECRET_PATH` 用于控制 WebTransport 入口路径。
- 根路径 `/` 将返回静态页面以满足伪装需求。

### Secret 管理

`PSK` 必须通过 Wrangler secret 设置：

```bash
wrangler secret put PSK
```

部署时请确保环境变量已经生效。

---

## 4. 云平台部署 (ClawCloud / Cloud Run / PaaS)

`aether-gateway` 已针对云原生环境进行了优化，支持通过环境变量自动配置。

### 核心特性
1. **端口自动适配**：服务端会自动识别并监听 `$PORT` 环境变量指定的端口。
2. **自动 TLS**：在非持久化存储环境（如无证书挂载）下，会自动生成自签名证书以满足 WebTransport 的加密要求。

### 部署步骤 (ClawCloud / Cloud Run)

1. **选择镜像**：使用 `ghcr.io/coolapijust/aether-rea:latest`。
2. **配置环境变量**：
   - `PSK`：设置您的预共享密钥（必填）。
3. **端口设置**：
   - 确保外部端口映射正确（通常为 443 或 3000）。
   - 内部端口：程序会自动绑定到平台提供的 `$PORT`。
4. **协议选择**：
   - **重要**：必须启用 **UDP** 支持，因为 WebTransport 运行在 HTTP/3 UDP 之上。

### 客户端连接建议
由于自签名证书可能导致浏览器或部分客户端报错，建议在客户端启动参数中添加相关跳过验证的逻辑（如果支持），或者在生产环境通过方案 A/B 挂载正式域名证书。
