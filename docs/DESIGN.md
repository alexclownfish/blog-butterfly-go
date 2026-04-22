# blog-butterfly-go 设计文档（DESIGN）

> 版本：v1.0  
> 状态：基于当前代码实现整理  
> 项目路径：`/root/blog-butterfly-go`

---

## 1. 设计目标

本文档描述 `blog-butterfly-go` 当前代码库的实现架构、核心模块、主要接口、部署方式与关键设计约束，用于帮助后续开发在现有真实实现之上继续演进，而不是按想象补图纸。

---

## 2. 系统总览

项目当前由三层组成：

1. **web-vue**：前台站点，Vue 3 + Vite 构建后的静态页面
2. **admin-vue**：后台管理 UI，Vue 3 + Vite 管理端
3. **Backend API**：Go 服务，位于 `backend`

### 2.1 逻辑关系

```text
浏览器
  ├─ 访问 web-vue 前台页面（Nginx）
  └─ 访问 admin-vue 后台页面（Nginx）
          │
          └─ 调用 http://172.28.74.191:31083/api
                    │
                    ├─ Gin Router
                    ├─ Auth Middleware
                    ├─ Controllers
                    ├─ Gorm
                    ├─ MySQL
                    └─ 七牛云素材存储
```

### 2.2 当前主线说明

当前前端主线已经明确切换为：

- `web-vue/`：前台站点主线
- `admin-vue/`：后台管理主线
- `frontend-已归档/`：历史静态实现归档目录

旧 `frontend/` 内容已转入 `frontend-已归档/`，仅保留历史参考，不再作为现役实现。

---

## 3. 代码结构

```text
blog-butterfly-go/
├── backend/
│   ├── main.go
│   ├── config/database.go
│   ├── controllers/
│   │   ├── article.go
│   │   ├── auth.go
│   │   ├── category.go
│   │   └── upload.go
│   ├── middleware/auth.go
│   ├── models/article.go
│   ├── router/router.go
│   ├── utils/
│   │   ├── jwt.go
│   │   └── qiniu.go
│   └── Dockerfile
├── web-vue/
│   ├── src/
│   ├── public/
│   ├── Dockerfile
│   └── nginx.conf
├── admin-vue/
│   ├── src/
│   ├── public/
│   ├── Dockerfile
│   └── nginx.conf
├── frontend-已归档/
│   ├── admin/
│   ├── js/
│   └── Dockerfile
├── k8s/
│   ├── backend.yaml
│   ├── web-vue.yaml
│   ├── admin-vue.yaml
│   └── frontend-已归档.yaml
├── deploy.sh
└── docs/
```

---

## 4. 前端设计

## 4.1 前端技术选型

当前前端主线已经切换为两个 Vue 3 + Vite 应用：

- `web-vue`：前台站点
- `admin-vue`：后台管理端

这种设计的优点：
- 前后台职责更清晰
- 组件化与状态管理更适合持续演进
- 与现代前端构建/部署方式一致

需要注意：
- `frontend-已归档/` 仅保留历史参考，不应继续承担主线开发

---

## 4.2 API Base 解析

`web-vue` 与 `admin-vue` 均通过各自前端配置访问后端 API；旧 `frontend-已归档/js/config.js` 仅保留历史参考。

当前默认值：
- `http://172.28.74.191:31083/api`

解析顺序：
1. `window.APP_CONFIG.apiBase`
2. `window.API_BASE`
3. `document.documentElement.dataset.apiBase`
4. `localStorage['api_base']`
5. 默认值

这意味着现役前端应继续保持可配置 API 地址能力；固定公网地址仅应视为当前部署默认值。

---

## 4.3 后台页面结构

`admin-vue` 后台页面由路由、页面视图与复用组件共同组成：

1. 左侧导航栏
   - 文章管理
   - 分类管理
   - 素材管理
2. 顶部 Header
3. Dashboard strip
4. 内容区 `#content`
5. 文章编辑弹窗 `#editorModal`
6. 图片选择弹窗 `#imagePickerModal`

### 4.3.1 编辑器设计

主编辑器弹窗中包含：
- 标题
- 摘要
- 分类
- 标签
- 封面 URL
- 封面预览
- 置顶状态
- 发布状态
- Markdown 正文编辑区
- 保存状态条
- 字数与阅读时长信息

### 4.3.2 Markdown 编辑能力

`app.js` 中通过 `initMarkdownEditor()` 初始化 EasyMDE：

- 关闭拼写检查
- 开启预览相关能力
- 工具栏包含：
  - bold
  - italic
  - heading
  - quote
  - unordered-list
  - ordered-list
  - code
  - link
  - 自定义 image-library
  - preview
  - side-by-side
  - fullscreen

自定义 `image-library` 按钮直接联动图片选择器，避免手工拼接 Markdown 图片语法。

---

## 4.4 状态管理设计

当前后台前端主要使用模块级变量管理页面状态，关键变量包括：

- `editingId`
- `articleFilters`
- `articlePagination`
- `articleCategories`
- `imageCache`
- `selectedImages`
- `markdownEditor`
- `autosaveTimer`
- `editorDirty`
- `isSavingArticle`
- `imagePickerMode`
- `currentEditorDraftKey`

这是典型的“单文件集中式状态管理”方案。

优点：
- 简单直接
- 调试成本低

风险：
- 功能继续增加时，耦合会越来越重
- 容易出现局部状态相互影响

---

## 4.5 本地草稿设计

自动保存基于 `localStorage` 实现。

### Key 规则
- 新建文章：`admin:draft:new`
- 编辑文章：`admin:draft:article:<id>`

### 自动保存触发字段
- 标题
- 摘要
- 封面
- 分类
- 标签
- 是否置顶
- 状态
- 正文

### 机制特点
- 使用防抖保存（2 秒）
- 支持恢复本地草稿
- 支持放弃本地草稿
- 正式保存到服务器后自动清除本地草稿

### 设计取舍
未引入服务端 autosave API，优先用浏览器本地保存降低后端改造成本。

---

## 5. 后端设计

## 5.1 技术栈

后端当前使用：

- Go
- Gin
- Gorm
- MySQL
- JWT
- 七牛云 SDK

入口文件：`backend/main.go`

启动流程：
1. `config.InitDB()` 初始化数据库
2. `gin.Default()` 创建路由
3. 挂载 `/uploads` 静态目录
4. `router.SetupRoutes(r)` 注册 API
5. 监听 `:8080`

---

## 5.2 数据模型

定义位于 `backend/models/article.go`。

### Article
- `ID`
- `Title`
- `Content`
- `Summary`
- `CoverImage`
- `CategoryID`
- `Category`
- `Tags`
- `IsTop`
- `Status`
- `Views`
- `CreatedAt`
- `UpdatedAt`

### Category
- `ID`
- `Name`

### Tag
- `ID`
- `Name`

### User
- `ID`
- `Username`
- `Password`

---

## 5.3 数据库设计

数据库初始化在 `backend/config/database.go`：

- 驱动：MySQL
- DSN 当前硬编码为：`root:ywz0207.@tcp(mysql:3306)/blog?...`
- 启动时带重试机制
- 自动迁移：
  - `Article`
  - `Category`
  - `Tag`
  - `User`

### 风险
- DSN 硬编码在代码中，不利于环境切换和安全治理

---

## 5.4 路由设计

路由定义位于 `backend/router/router.go`。

### 公共中间件
注册了简化 CORS：
- `Access-Control-Allow-Origin: *`
- `Allow-Methods: GET, POST, PUT, DELETE`
- `Allow-Headers: Content-Type, Authorization`

### API 前缀
- `/api`

### 公开接口
- `GET /api/health`
- `GET /api/articles`
- `GET /api/articles/:id`
- `GET /api/categories`
- `GET /api/tags`
- `POST /api/login`

### 鉴权接口
通过 `middleware.AuthMiddleware()` 保护：
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

---

## 5.5 鉴权设计

### 登录
`controllers/auth.go` 的 `Login()` 流程：
1. 接收用户名密码
2. 从数据库查询用户
3. 使用 `bcrypt.CompareHashAndPassword` 校验密码
4. 成功后生成 JWT token

### JWT
位于 `backend/utils/jwt.go`：
- Claims 包含 `user_id`、`username`
- 有效期：24 小时
- 默认签名算法：HS256
- 密钥来自环境变量 `JWT_SECRET`
- 若未配置，则回退到默认值 `your-secret-key`

### 鉴权中间件
位于 `backend/middleware/auth.go`：
- 读取 `Authorization` 头
- 去除 `Bearer ` 前缀
- 调用 `ParseToken`
- 成功后将 `user_id` 写入 Gin context

### 风险
- JWT 存在默认弱密钥回退逻辑，生产环境应避免

---

## 5.6 文章接口设计

位于 `backend/controllers/article.go`。

### 文章状态
支持两种状态：
- `draft`
- `published`

### 查询能力
`GetArticles()` 支持：
- 状态筛选（默认 `published`）
- 关键词搜索（标题/正文）
- 分类筛选
- 标签模糊匹配
- 分页
- 按 `is_top desc, created_at desc` 排序
- 预加载 `Category`

### 详情能力
`GetArticle()`：
- 根据 `id` 查询文章
- 返回文章详情
- 会把 `views` 加 1

### 写入能力
- `CreateArticle()`：默认空状态写入 `draft`
- `UpdateArticle()`：允许更新全部主要字段
- `DeleteArticle()`：按 ID 删除

### 设计特点
- 请求结构简单直接
- REST 风格较清晰
- 列表接口默认只看已发布文章，对前台展示友好，对后台则需要显式传 `status`

---

## 5.7 分类与统计接口设计

位于 `backend/controllers/category.go`。

### 分类能力
- 获取分类列表
- 创建分类
- 更新分类
- 删除分类

### 标签能力
- 获取标签列表

### Dashboard 统计
`GetDashboardStats()` 返回：
- 文章总数
- 已发布文章数
- 草稿数
- 分类数
- 图片数
- 置顶文章数

其中图片数通过 `utils.ListQiniuImages()` 动态计算。

---

## 5.8 素材接口设计

位于 `backend/controllers/upload.go` 与 `backend/utils/qiniu.go`。

### 上传
- 接收表单字段 `image`
- 调用 `UploadToQiniu()`
- 成功返回 `{ "url": ... }`

### 列表
- 调用 `ListQiniuImages()`
- 返回 `{ "data": [...] }`

### 删除
- 按 key 删除七牛素材

### 七牛配置
从 `config.ini` 的 `[qiniu]` 段读取：
- `AccessKey`
- `SecretKey`
- `Bucket`
- `QiniuServer`

### 设计特点
- 配置只加载一次（`sync.Once`）
- URL 拼接规则为 `QiniuServer + key`
- 列表结果中包含：
  - `url`
  - `key`
  - `size`
  - `time`

### 风险
- `UseHTTPS: false`
- 强依赖 `config.ini`
- 删除接口把 key 直接放在 URL path 中，若 key 编码复杂需额外注意

---

## 6. 部署设计

## 6.1 Docker

### Backend Dockerfile
- 构建镜像：`golang:1.21-alpine`
- 执行：`go mod tidy && go build -o main .`
- 运行镜像：`alpine:latest`
- 暴露端口：`8080`
- 复制 `config.ini`

### web-vue / admin-vue Dockerfile
- 基于 Node 20 执行前端构建
- 构建产物通过 `nginx:alpine` 对外提供静态服务
- 暴露端口：`80`

---

## 6.2 Kubernetes
