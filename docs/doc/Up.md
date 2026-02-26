# Router 从零到公网可用的完整交付手册（指挥官视角）

我按“部署总指挥 + 教练”的方式编排：你能直接复制粘贴执行，也能理解每一步为什么这么做。目标只有三条：
- 不影响同机其他服务
- 强制连接同一套 PostgreSQL（禁止 SQLite）
- 公网 `https://router.yeying.pub/` 可访问

---

**Architecture**

```mermaid
flowchart LR
    User[Client / Browser] -->|HTTPS 443| Nginx[Nginx: router.yeying.pub]
    Nginx -->|proxy_pass 127.0.0.1:13011| Router[Router Service]
    Router -->|SQL_DSN| PG[(PostgreSQL)]

    subgraph Host[Server Host]
        Nginx
        Router
    end
```

---

**Prereqs**

你从“全新机器 + 全新仓库”开始时，先保证基础工具齐全。

```bash
# 基础工具与依赖
apt-get update
apt-get install -y git curl build-essential
apt-get install -y postgresql-client

# Go / Node / npm 需符合版本要求
# Go 1.22+，Node 18+，npm

go version
node -v
npm -v
```

如果你需要部署公网 HTTPS：
```bash
apt-get install -y nginx certbot python3-certbot-nginx
nginx -v
```

---

**Step 0: 非侵入式调研**

不打扰其他服务的前提下，先扫清冲突与现存服务。

```bash
# 端口与服务占用
ss -lntp
systemctl list-units --type=service --no-pager
ps aux

docker ps
```

指挥官提示：
- 目标端口建议从 `13011` 起步，避免与 80/443/3011 冲突。
- 若 13011 被占用，顺延到 13012/13013，并保持 Nginx 反代一致。

---

**Step 1: 克隆仓库**

从 GitHub 拉取最新源码：

```bash
cd /root/code

git clone https://github.com/yeying-community/router.git
cd /root/code/router
```

如果目录已存在，直接进入：
```bash
cd /root/code/router
```

---

**Step 2: 写入 .env（核心）**

`.env` 是强制走 PG 的唯一锚点，必须写对、必须位于 `WorkingDirectory` 下。

```bash
cd /root/code/router

cat > .env <<'EOF_ENV'
SQL_DSN=postgres://router:***@51.75.133.235:5432/router?sslmode=disable
EOF_ENV

chmod 600 .env
```

参数解释：
- `postgres://user:pass@host:port/db?sslmode=disable`
- `sslmode=disable` 代表不启用 TLS，若你强制 TLS 则改成 `require` 或 `verify-full`
- `.env` 权限必须 600，避免敏感信息泄露

---

**Step 3: 验证 PG 可连通（必须做）**

```bash
cd /root/code/router
set -a; source .env; set +a
psql "$SQL_DSN" -Atqc "select current_user, current_database(), inet_server_port();"
```

你必须看到类似输出：
```
router|router|5432
```

这一步是硬门槛：不通过就不要启动服务。

---

**Step 4: 构建前端（必须有 web/dist）**

```bash
cd /root/code/router

npm install --prefix web
npm run build --prefix web
```

解释：
- 本仓库没有 `package-lock.json`，因此 `npm ci` 会失败
- `web/dist` 是 Go embed 所需资源，没有它编译必失败

---

**Step 5: 构建后端**

```bash
cd /root/code/router

mkdir -p build
GOFLAGS="" go build -o build/router ./cmd/router
```

---

**Step 6: systemd 服务（守护运行）**

推荐 systemd，方便统一日志与进程守护。这里采用端口 `13011`。

```bash
cat > /etc/systemd/system/router.service <<'EOF_SERVICE'
[Unit]
Description=Router Local Service
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/root/code/router
ExecStart=/root/code/router/build/router --port 13011 --log-dir ./logs
Restart=on-failure
RestartSec=3

[Install]
WantedBy=multi-user.target
EOF_SERVICE

systemctl daemon-reload
systemctl restart router
systemctl status router --no-pager
```

指挥官提示：
- `WorkingDirectory` 必须是仓库根，否则 `.env` 不加载
- `--port` 必须和 Nginx 反代端口一致

---

**Step 7: Nginx 反代（公网可访问）**

保证 `router.yeying.pub` 反代到 127.0.0.1:13011。

```bash
cat > /etc/nginx/conf.d/router.conf <<'EOF_NGX'
server {
    server_name router.yeying.pub;

    client_max_body_size 1000M;
    client_body_timeout 300s;
    client_body_buffer_size 128k;

    add_header Strict-Transport-Security "max-age=63072000; includeSubDomains; preload";
    add_header X-Content-Type-Options nosniff;
    add_header X-Frame-Options DENY;
    add_header X-XSS-Protection "1; mode=block";
    add_header Referrer-Policy "no-referrer-when-downgrade";

    ssl_dhparam /etc/letsencrypt/ssl-dhparams.pem;

    location / {
        proxy_pass http://127.0.0.1:13011;

        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        proxy_request_buffering off;
        proxy_read_timeout 300s;
        proxy_connect_timeout 300s;
        proxy_send_timeout 300s;

        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }

    listen 443 ssl;
    ssl_certificate /etc/letsencrypt/live/router.yeying.pub/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/router.yeying.pub/privkey.pem;
    include /etc/letsencrypt/options-ssl-nginx.conf;
}

server {
    if ($host = router.yeying.pub) {
        return 301 https://$host$request_uri;
    }

    server_name router.yeying.pub;
    listen 80;
    return 404;
}
EOF_NGX

nginx -t && systemctl reload nginx
```

如证书不存在，先签发：
```bash
certbot --nginx -d router.yeying.pub
```

---

**Step 8: 验证（硬标准）**

```bash
# 1) 日志必须出现 openPostgreSQL
journalctl -u router --since 'today' --no-pager | rg "openPostgreSQL|openSQLite|openMySQL" -S

# 2) 本地健康检查
curl -s http://127.0.0.1:13011/api/status

# 3) 公网可达性
curl -I https://router.yeying.pub
```

只要出现 `openSQLite`，立即停下排查 `.env` 与 `WorkingDirectory`。

---

**Troubleshoot**

常见问题与对策：
- `openSQLite` 或默认账户 `root/123456`：说明没加载 `.env`，检查 `WorkingDirectory` 与 `.env` 位置。
- `embed.go: pattern web/dist/*: no matching files found`：先 `npm run build --prefix web` 再 `go build`。
- `502 Bad Gateway`：Nginx 反代端口与 Router 监听端口不一致。
- `nginx: bind() to 0.0.0.0:443 failed`：端口被占用，检查已有 Nginx 或其他服务。

---

**Security**

安全红线：
- `.env` 不进 Git，权限保持 `600`
- 文档与截图不输出明文密码
- 端口变更必须同步修改 Router 与 Nginx

---

**SSH 端口转发**

如果你只是想在本地浏览器访问：

```bash
ssh -L 13011:127.0.0.1:13011 root@<服务器IP或域名>

curl -s http://127.0.0.1:13011/api/status
```

---

**Quick Recall**

三条记忆法：
1. `WorkingDirectory` 对了 `.env` 才会生效
2. Nginx 端口必须与 Router 端口一致
3. 日志出现 `openPostgreSQL` 才算合格

---

**UCAN Audience 修复（本次成功经验）**

```mermaid
sequenceDiagram
    participant Chat as chat.yeying.pub
    participant Router as router.yeying.pub
    Chat->>Router: UCAN token (aud=did:web:router.yeying.pub)
    Router->>Router: 校验 expectedAud
    Router-->>Chat: 通过 (no mismatch)
```

当 Router 不在默认端口 `3011` 时，`UCAN_AUD` 必须显式配置，否则 Router 会按本机端口推导出 `did:web:localhost:<PORT>`，导致 `UCAN audience mismatch`。

**一行修复（已验证有效）**
```bash
# /root/code/router/.env
UCAN_AUD=did:web:router.yeying.pub
```

规则总结（简洁但必须牢记）：
- **默认端口 3011**：可不设置 `UCAN_AUD`，Router 会自动推导为 `did:web:localhost:3011`（且常与本地一致）
- **改端口（如 13011）**：必须显式设置 `UCAN_AUD=did:web:router.yeying.pub`，确保和 Chat 端生成的 aud 一致
