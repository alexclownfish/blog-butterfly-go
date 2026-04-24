# csdn-sync-provider

一个给 `blog-butterfly-go` backend 使用的 **独立 CSDN sync provider 服务骨架**。

它当前不是“真实 CSDN 登录逆向实现”，而是一个可独立运行、可 Docker 化、可被 backend 用 `CSDN_SYNC_BASE_URL` 对接的 provider skeleton，先把服务边界、接口契约、Compose 接线和联调路径铺好。

## 当前能力

- `GET /health`：健康检查
- `POST /login/start`：返回 provider_session、二维码 URL、提示文案
- `GET /login/status?session=...`：返回登录状态
- `GET /articles?session=...`：返回占位文章列表
- `GET /articles/:id?session=...`：返回占位文章正文

## 当前定位

这是一个 **独立服务骨架**，主要用于：

1. 让 backend 的 `RealCSDNSyncProvider` 有可访问的远端 provider 服务；
2. 在不把复杂逆向逻辑塞进 backend 的前提下，先完成服务拆分；
3. 给后续真实扫码、cookie/session 管理、文章抓取逻辑预留干净接口。

一句人话总结：

> 现在它是“会说话的假人”，但接口是真的，部署链路也是真的。

## 启动方式

### 本地直接运行

```bash
cd csdn-sync-provider
go run .
```

默认监听：`8091`

### 本地测试 / 构建

```bash
cd csdn-sync-provider
go test ./...
go build ./...
```

## Docker Compose 部署

目录内已补好独立 Compose 部署文件：

- `docker-compose.yml`
- `install-docker.sh`
- `script/install-docker-compose.sh`
- `.env.example`

### 本地仓库内一键启动

```bash
cd csdn-sync-provider
bash install-docker.sh
```

等价命令：

```bash
bash script/install-docker-compose.sh
```

默认行为：

- 自动检查 `docker` 与 `docker compose`
- 若 `.env` 不存在，则从 `.env.example` 自动复制
- 启动前执行 `docker compose config` 预检查
- 执行 `docker compose up -d --build`
- 暴露宿主机端口 `8091`

启动后访问：

- Health: `http://127.0.0.1:8091/health`

### 远程自举安装（curl | bash 风格）

现在 `script/install-docker-compose.sh` 已支持 **自举拉仓库**。

当脚本检测到：

- 当前不是在完整仓库目录中执行
- 或者是 `curl | bash` / pipe 方式执行

它会自动：

1. 安装基础依赖（缺少时）：`git`、`curl`、`ca-certificates`
2. 把仓库拉到 `INSTALL_DIR`
3. 切换到 `REPO_REF`
4. 进入 `PROVIDER_SUBDIR`（默认 `csdn-sync-provider`）继续执行 Compose 部署

示例：

```bash
curl -fsSL https://raw.githubusercontent.com/alexclownfish/blog-butterfly-go/main/csdn-sync-provider/script/install-docker-compose.sh | sudo bash
```

也可自定义：

```bash
curl -fsSL https://raw.githubusercontent.com/alexclownfish/blog-butterfly-go/main/csdn-sync-provider/script/install-docker-compose.sh \
  | sudo REPO_REF=main INSTALL_DIR=/opt/blog-butterfly-go bash
```

### 常用参数

```bash
# 启动但不重新 build
bash script/install-docker-compose.sh --no-build

# 停止并删除容器
bash script/install-docker-compose.sh --down

# 启动后直接追日志
bash script/install-docker-compose.sh --logs

# 强制刷新仓库（自举/远程部署场景推荐）
bash script/install-docker-compose.sh --pull

# 禁止自动安装 Docker
bash script/install-docker-compose.sh --skip-install-docker
```

### 支持的环境变量覆盖

```env
REPO_URL=https://github.com/alexclownfish/blog-butterfly-go.git
REPO_REF=main
INSTALL_DIR=/opt/blog-butterfly-go
PROVIDER_SUBDIR=csdn-sync-provider
AUTO_INSTALL_DOCKER=1
```

说明：

- `REPO_URL`：自举时拉取的仓库地址
- `REPO_REF`：要部署的分支/引用
- `INSTALL_DIR`：仓库落地目录
- `PROVIDER_SUBDIR`：provider 在仓库里的子目录
- `AUTO_INSTALL_DOCKER=1`：缺少 Docker 时尝试自动安装

### Compose 配置说明

`docker-compose.yml` 当前只部署一个服务：

- 服务名：`csdn-sync-provider`
- 容器名默认：`csdn-sync-provider`
- 镜像名默认：`csdn-sync-provider:local`
- 宿主机端口默认：`8091`

可通过 `.env` 覆盖：

- `CSDN_SYNC_PROVIDER_PORT`
- `CSDN_SYNC_PROVIDER_IMAGE`
- `CSDN_SYNC_PROVIDER_CONTAINER_NAME`
- `IMAGE_REGISTRY_PREFIX`

其中 `IMAGE_REGISTRY_PREFIX` 默认是：

```env
IMAGE_REGISTRY_PREFIX=docker.m.daocloud.io/library/
```

这是为了在部分网络环境下更稳地拉取基础镜像，避免 Docker Hub 抽风表演杂技。

## systemd 部署

目录内已补好 systemd unit 模板：

- `systemd/csdn-sync-provider.service`

这是 **二进制直跑** 方案，不依赖 Docker，适合已有主机服务体系的场景。

### 推荐目录约定

```text
/opt/csdn-sync-provider/
├── csdn-sync-provider
├── .env
└── systemd/csdn-sync-provider.service
```

### 部署步骤

#### 1）构建二进制

```bash
cd csdn-sync-provider
go build -o csdn-sync-provider .
```

#### 2）准备目录与环境文件

```bash
sudo mkdir -p /opt/csdn-sync-provider
sudo cp csdn-sync-provider /opt/csdn-sync-provider/
sudo cp .env.example /opt/csdn-sync-provider/.env
```

#### 3）创建运行用户

```bash
sudo useradd --system --no-create-home --shell /usr/sbin/nologin csdn-sync-provider || true
sudo chown -R csdn-sync-provider:csdn-sync-provider /opt/csdn-sync-provider
```

#### 4）安装 unit 文件

```bash
sudo cp systemd/csdn-sync-provider.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now csdn-sync-provider
```

#### 5）查看运行状态

```bash
sudo systemctl status csdn-sync-provider
sudo journalctl -u csdn-sync-provider -f
```

### systemd 配置说明

unit 默认约定：

- 运行用户：`csdn-sync-provider`
- 工作目录：`/opt/csdn-sync-provider`
- 环境文件：`/opt/csdn-sync-provider/.env`
- 启动命令：`/opt/csdn-sync-provider/csdn-sync-provider`

如果你要换目录或用户，改 `systemd/csdn-sync-provider.service` 里的：

- `User=`
- `Group=`
- `WorkingDirectory=`
- `EnvironmentFile=`
- `ExecStart=`

## 环境变量

可参考 `.env.example`：

### 服务监听 / 部署相关

- `PORT`
- `CSDN_SYNC_PROVIDER_PORT`
- `CSDN_SYNC_PROVIDER_IMAGE`
- `CSDN_SYNC_PROVIDER_CONTAINER_NAME`
- `IMAGE_REGISTRY_PREFIX`

### provider skeleton 业务占位配置

- `CSDN_PROVIDER_NAME`
- `CSDN_PROVIDER_MODE`
- `CSDN_PROVIDER_SESSION_ID`
- `CSDN_PROVIDER_QR_CODE_URL`
- `CSDN_PROVIDER_LOGIN_MESSAGE`
- `CSDN_PROVIDER_DEFAULT_STATUS`
- `CSDN_PROVIDER_STATUS_MESSAGE`
- `CSDN_PROVIDER_ARTICLE_ID`
- `CSDN_PROVIDER_ARTICLE_TITLE`
- `CSDN_PROVIDER_ARTICLE_SUMMARY`
- `CSDN_PROVIDER_ARTICLE_URL`
- `CSDN_PROVIDER_ARTICLE_PUBLISHED_AT`
- `CSDN_PROVIDER_ARTICLE_TAGS`
- `CSDN_PROVIDER_ARTICLE_CONTENT`
- `CSDN_PROVIDER_ARTICLE_COVER`

## API 契约示例

### `POST /login/start`

```json
{
  "provider": "csdn",
  "provider_mode": "skeleton",
  "provider_session": "provider-skeleton-session",
  "qr_code_url": "https://example.com/csdn-provider-skeleton-qr.png",
  "message": "provider skeleton ready: 请接入真实 CSDN 登录逻辑"
}
```

### `GET /login/status?session=provider-skeleton-session`

```json
{
  "provider": "csdn",
  "provider_mode": "skeleton",
  "provider_session": "provider-skeleton-session",
  "status": "pending",
  "message": "waiting for real provider integration",
  "qr_code_url": "https://example.com/csdn-provider-skeleton-qr.png"
}
```

### `GET /articles?session=provider-skeleton-session`

```json
{
  "articles": [
    {
      "id": "provider-skeleton-article",
      "title": "CSDN Provider Skeleton Article",
      "summary": "独立 provider 服务骨架占位文章，用于 backend 真接口联调。",
      "source_url": "https://blog.csdn.net/demo/article/details/provider-skeleton",
      "published_at": "2026-04-24T03:00:00Z"
    }
  ]
}
```

## 下一步真实化建议

后续如果要把 skeleton 变成真实 provider，建议按这个方向演进：

1. 在 `POST /login/start` 接入真实二维码生成 / 登录事务创建；
2. 在 `GET /login/status` 接入轮询状态机；
3. 在服务内部维护 session/cookie/token store；
4. 在 `GET /articles` 对接真实文章列表接口；
5. 在 `GET /articles/:id` 拉正文并做 HTML/Markdown 标准化；
6. 视情况增加日志脱敏、重试、限流、过期清理。

## 验证现状

当前已补：

- HTTP handler 自动化测试
- `go test ./...` 可通过
- `go build ./...` 可通过
- `bash -n install-docker.sh script/install-docker-compose.sh` 可通过

注意：当前执行环境里 **没有安装 Docker**，所以这次只能完成静态验证，没法在这里直接跑 `docker compose up` 做运行态验收。

如果后续你要我继续，我下一刀就可以把这个 skeleton 升级成：

- 有内存 session store 的 provider 服务；
- 或直接接入真实 CSDN 登录链路的 provider v1。
