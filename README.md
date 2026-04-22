# blog-butterfly-go

`blog-butterfly-go` 是一个围绕个人博客内容发布场景构建的全栈项目，当前现役主链路由三部分组成：

- `web-vue`：博客前台站点，负责公开内容展示
- `admin-vue`：博客管理后台，负责内容创作与素材管理
- `backend`：Go API，负责鉴权、内容数据、图片与统计接口

部署层由 `k8s/` 与根目录 `deploy.sh` 串起来，当前实际运行形态是：

**Vue 前台 + Vue 后台 + Go API + K3s/Kubernetes**

`frontend-已归档/` 仅作为历史实现保留，不属于当前主线产品。

---

## 1. 项目整体关系

这三个现役模块不是平行摆设，而是一个完整的数据流闭环：

```text
创作者 / 管理员
        │
        ▼
 admin-vue  ──────▶  backend  ──────▶  MySQL / 七牛云
    │                  │
    │                  └─────▶ 提供公开文章/分类/标签接口
    │
    └── 登录、改密、文章编辑、分类管理、素材管理

普通访客
   │
   ▼
 web-vue  ─────────▶  backend  ──────▶  读取文章/分类/标签数据
```

可以简单理解为：

- `admin-vue` 是内容生产端
- `backend` 是统一数据与鉴权中台
- `web-vue` 是内容消费端

也就是说，**后台写内容，后端存和管，前台读出来展示**。

---

## 2. 当前现役模块说明

### 2.1 `admin-vue`：内容生产后台

`admin-vue` 是当前管理端主线，基于 **Vue 3 + TypeScript + Vite + Element Plus**。

当前已实现：

- 登录
- 首次/默认密码登录后的强制改密
- 工作台 Dashboard
- 文章管理
  - 列表
  - 搜索
  - 状态筛选
  - 分类筛选
  - 分页
  - 新建 / 编辑 / 删除
- Markdown 文章编辑器
  - 编辑 / 分栏预览 / 仅预览
  - 常用 Markdown 快捷插入
  - 封面图选择
  - 图片插入正文
  - `Ctrl/Cmd + S` 保存
- 本地草稿自动保存 / 恢复 / 清理
- 分类管理
  - 新建 / 编辑 / 删除
- 素材管理
  - 上传
  - 预览
  - 复制 URL
  - 删除 / 批量删除
  - 文章编辑器内直接选图

后台主要路由主线：

- `/dashboard`
- `/articles`
- `/categories`
- `/images`
- `/login`
- `/change-password`

详细说明可看：

- [`admin-vue/README.md`](./admin-vue/README.md)

### 2.2 `web-vue`：公开博客前台

`web-vue` 是当前博客前台主线，基于 **Vue 3 + Vite**。

从路由与 API 调用可确认，当前已覆盖这些页面与能力：

- 首页文章列表：`/`
- 文章详情：`/posts/:id.html`
- 标签详情：`/tags/:name`
- 分类页：`/categories/`
- 归档页：`/archives/`
- 关于页：`/about/`

它主要消费后端公开接口：

- `GET /api/articles`
- `GET /api/articles/:id`
- `GET /api/categories`
- `GET /api/tags`

所以它和后台的关系非常明确：

- `admin-vue` 负责把内容写进去
- `web-vue` 负责把内容展示出来
- 二者通过同一个 `backend` 共享内容数据

### 2.3 `backend`：统一 API 与业务中心

`backend` 是 Go 服务端，当前入口为：

- `backend/main.go`

实际启动行为：

1. 初始化数据库连接
2. 挂载静态目录 `/uploads`
3. 注册 `/api` 路由
4. 监听 `:8080`

从当前路由代码可确认，后端已经提供：

#### 公开接口

- `GET /api/health`
- `GET /api/articles`
- `GET /api/articles/:id`
- `GET /api/categories`
- `GET /api/tags`
- `POST /api/login`

#### 认证后接口

- `POST /api/change-password`
- `POST /api/articles`
- `PUT /api/articles/:id`
- `DELETE /api/articles/:id`
- `POST /api/categories`
- `PUT /api/categories/:id`
- `DELETE /api/categories/:id`
- `GET /api/dashboard/stats`
- `POST /api/upload`
- `GET /api/images`
- `DELETE /api/images/:key`

认证方式为：

```http
Authorization: Bearer <token>
```

后端同时承担这几个核心职责：

- 登录鉴权
- 强制改密策略
- 文章/分类/标签数据管理
- Dashboard 统计数据
- 图片上传、列举、删除
- 向前台和后台同时提供接口能力

---

## 3. 三个模块如何协同工作

### 内容发布链路

```text
管理员在 admin-vue 登录
  → 创建/编辑文章
  → 选择分类、标签、封面、正文图片
  → 提交到 backend
  → backend 写入 MySQL，并处理图片资源
  → web-vue 通过公开接口读取并展示给访客
```

### 素材管理链路

```text
admin-vue 上传图片
  → backend 调用图床/存储能力
  → 返回图片 URL
  → admin-vue 可用于封面或插入 Markdown 正文
```

### 鉴权链路

```text
admin-vue 登录获取 token
  → 后续请求携带 Authorization: Bearer <token>
  → backend 校验身份
  → 决定是否允许访问受保护接口
```

一句话总结就是：

> `admin-vue` 管内容，`web-vue` 展内容，`backend` 负责两边都能顺畅说人话。

---

## 4. 技术栈

### 前端

- Vue 3
- TypeScript
- Vite
- Vue Router
- Axios
- Element Plus（后台）
- Pinia（后台）
- Marked（Markdown 渲染）
- Vitest（后台测试）

### 后端

- Go 1.21
- Gin
- Gorm
- MySQL
- JWT
- 七牛云 SDK

### 部署

- Docker
- Nginx
- K3s / Kubernetes
- NodePort 对外暴露

---

## 5. 仓库结构

```text
blog-butterfly-go/
├── backend/               # Go API 服务
├── web-vue/               # Vue 前台站点
├── admin-vue/             # Vue 管理后台
├── k8s/                   # Kubernetes 部署清单
├── docs/                  # PRD / 设计文档 / 实施计划
├── frontend-已归档/       # 历史静态前端，仅供参考
└── deploy.sh              # 构建并部署现役三件套
```

### 现役目录

- `backend/`
- `web-vue/`
- `admin-vue/`
- `k8s/`

### 历史归档目录

- `frontend-已归档/`

阅读仓库时请优先围绕现役目录理解系统，不要把归档前端和当前主线混为一谈。

---

## 6. 本地开发

### 6.1 启动后端

```bash
cd backend
go mod tidy
go run .
```

后端默认监听：

- `:8080`

### 6.2 启动前台

```bash
cd web-vue
npm install
npm run dev
```

### 6.3 启动后台

```bash
cd admin-vue
npm install
npm run dev
```

后台还支持运行测试：

```bash
cd admin-vue
npm test
```

---

## 7. 部署方式

仓库根目录提供 `deploy.sh`：

```bash
./deploy.sh
```

脚本当前会按顺序执行：

1. 构建 `backend` Docker 镜像
2. 构建 `web-vue` Docker 镜像
3. 构建 `admin-vue` Docker 镜像
4. 在 `k8s/` 目录应用 Kubernetes 资源

也就是说，这个脚本部署的就是当前现役三件套，而不是归档前端。

### 当前暴露地址

根据 `deploy.sh` 与 `k8s/*.yaml`，当前对外地址为：

- 前台：`http://172.28.74.191:31086`
- 后台：`http://172.28.74.191:31085`
- API：`http://172.28.74.191:31083/api`

### Kubernetes 资源

- 命名空间：`blog-butterfly-go`
- 后端 Service：NodePort `31083`
- 后台 Service：NodePort `31085`
- 前台 Service：NodePort `31086`
- `k8s/mysql-alias.yaml` 使用 `ExternalName` 将 `mysql` 指向 `mysql.default.svc.cluster.local`

---

## 8. 配置与运行时说明

### 数据库

当前 `backend/config/database.go` 中数据库 DSN 仍为硬编码方式，并通过 Kubernetes 内 `mysql` 服务名连接数据库。

### JWT

JWT 密钥读取环境变量 `JWT_SECRET`；如果未提供，会回退到内置默认值。生产环境应显式配置。

### 默认管理员

后端启动时会检查默认管理员，并支持通过以下环境变量覆盖：

- `DEFAULT_ADMIN_USERNAME`
- `DEFAULT_ADMIN_PASSWORD`

首次登录后会要求修改密码。

### 图床 / 图片存储

图片上传、列举、删除能力依赖七牛云配置；当前后端 Dockerfile 会将 `config.ini` 一并打包进镜像。

> 更健康的后续方向是把数据库、JWT、图床等敏感配置逐步迁移到环境变量或 Kubernetes Secret，而不是继续靠硬编码和打包配置文件硬扛。

---

## 9. 已有文档

仓库内已有产品与设计文档：

- `docs/PRD.md`
- `docs/DESIGN.md`
- `docs/PRD-admin-content-workspace-p2.md`
- `docs/plans/admin-content-workspace-p2-implementation-plan.md`
- `admin-vue/README.md`

如果你要继续补产品说明、接口说明或部署文档，建议优先在这些文档基础上增量更新，不要重新发明轮子。

---

## 10. 当前已知约束 / 技术债

基于当前代码与部署文件，可确认这些现实约束：

- 数据库连接仍存在硬编码
- JWT 存在默认回退密钥
- 七牛配置仍通过 `config.ini` 提供
- 部署方式以本机构建 Docker 镜像 + K3s NodePort 暴露为主
- `frontend-已归档/` 仍保留历史资源，阅读代码时要避免误判为当前线上主线
- 根 README 与子项目 README 需要持续同步，避免出现“代码已经打工，文档还在实习”的情况

---

## 11. 推荐开发主线

当前仓库更适合沿着下面这条主线继续演进：

1. 以前台 `web-vue` + 后台 `admin-vue` + Go API 为唯一现役主线
2. 继续增强后台内容创作效率
   - Markdown 编辑体验
   - 图床联动效率
   - 自动保存草稿
3. 继续提升前台内容展示体验与内容组织能力
4. 逐步把配置与部署方式正规化
   - 环境变量
   - Secret 管理
   - 可重复部署
5. 将 `frontend-已归档/` 仅保留为视觉与历史行为参考

---

## 12. 备注

如果你第一次打开这个仓库，最值得记住的不是目录名字，而是这句话：

> **`admin-vue` 写内容，`backend` 管内容，`web-vue` 展内容。**

这就是当前 `blog-butterfly-go` 的主链路全貌。