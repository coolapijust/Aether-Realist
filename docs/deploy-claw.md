# ClawCloud 部署实操指南

针对 ClawCloud 环境，我们推荐两种部署方式。由于 WebTransport 强依赖 **UDP 443** 端口，请根据您的实际产品类型选择合适方案。

## 方案 A：Claw VPS + Docker Compose (推荐 - 生产级)
**适用场景**：已购买 ClawCloud 服务器（VPS），拥有全系统权限。

### 1. 准备环境
- 确保已安装 Docker 和 Docker Compose。
- 将本地 `deploy/` 目录同步到服务器（可以使用 `scp` 或 `git clone`）。

### 2. 关键操作步骤
1. **配置域名**：将您的域名解析到 Claw VPS 的 IP。
2. **修改 Caddyfile**：
   - 进入 `deploy/` 目录，编辑 `Caddyfile`。
   - 将 `your-domain.com` 替换为您的真实域名。
3. **设置 PSK**：
   - 编辑 `docker-compose.yml`。
   - 修改 `aether-gateway` 服务下的 `PSK=your_super_secret_token`。
4. **防火墙放行**：
   - **核心步骤**：在 Claw 控制面板或服务器内（ufw/iptables）务必放行 **443/UDP** 和 **443/TCP**。
5. **一键启动**：
   ```bash
   cd deploy
   docker-compose up -d
   ```

---

## 方案 B：ClawCloud Run (PaaS - 轻量化)
**适用场景**：使用 Claw 的容器托管服务（类似 Cloud Run）。

### 1. 关键配置表 (Launchpad)
在创建 APP 时，请严格按照以下参数填写：

| 配置项 | 推荐值 | 说明 |
| :--- | :--- | :--- |
| **容器镜像** | `ghcr.io/coolapijust/aether-rea:latest` | 使用最新的预构建镜像 |
| **环境变量** | `PSK=你的密钥` | 必需，用于连接认证 |
| **监听端口** | `443` | 内部程序会自动从 `$PORT` 读取或默认为 443 |
| **出口端口** | `443` | 外部访问端口 |
| **网络协议** | **UDP & TCP** | **极其重要**：如果不支持 UDP，WebTransport 将无法使用 |

### 2. 操作注意事项
1. **UDP 检查**：由于 WebTransport 基于 HTTP/3，如果 Claw 的容器服务屏蔽了 UDP 443，客户端将降级到普通 HTTPS 或连接失败。
2. **证书验证**：
   - PaaS 模式下通常自带 TLS 终结（SSL 证书）。
   - 如果遇到证书问题，请在客户端开启 `tls_insecure_skip_verify` 或在镜像启动参数中挂载证书。
3. **健康检查**：
   - 建议将健康检查路径设为 `/`（这是我们的伪装站点路径）。

## 验证部署
部署完成后，请使用以下命令测试：
```bash
# 测试 TCP 连通性
curl -I https://你的域名/

# 测试 WebTransport 特征 (预期返回 401 主动防御或正常的 WT 握手响应)
curl -i -H "Upgrade: webtransport" https://你的域名/v1/api/sync
```
