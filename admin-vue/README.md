# admin-vue

`admin-vue` 是 `blog-butterfly-go` 项目的现役管理后台前端，基于 **Vue 3 + TypeScript + Vite + Element Plus** 构建，用来对接 Go 后端真实接口，完成博客内容管理与素材管理。

它不是一个演示模板，也不是纯静态假页面，而是当前后台主链路的一部分。

---

## 1. 这个项目是干嘛的

`admin-vue` 主要服务于博客后台管理场景，当前已经覆盖这些核心能力：

- 登录
- 首次/默认密码登录后的强制改密
- 工作台 Dashboard
- 文章管理
- 分类管理
- 素材管理
- 与真实后端 API 联调

当前后台导航主线为：

- `工作台 /dashboard`
- `文章管理 /articles`
- `分类管理 /categories`
- `素材管理 /images`

---

## 2. 技术栈

- Vue 3
- TypeScript
- Vite
- Vue Router
- Pinia
- Axios
- Element Plus
- Vitest
- jsdom

---

## 3. 当前实现能力

### 3.1 登录与鉴权

- 登录页：`/login`
- 修改密码页：`/change-password`
- 登录成功后将 token 持久化到本地存储
- 路由守卫会校验登录状态
- 若后端返回强制改密标记，前端会自动跳转到修改密码页
- 请求拦截器会自动携带：

```http
Authorization: Bearer <token>
```

### 3.2 工作台 Dashboard

当前 Dashboard 已作为后台首页存在，主要承担入口和阶段提示作用：

- 引导进入文章管理
- 展示当前后台建设重点
- 强调真实 API 联调策略

### 3.3 文章管理

文章页当前支持：

- 加载文章列表
- 按关键词搜索标题或正文
- 按状态筛选（`published` / `draft`）
- 按分类筛选
- 分页浏览
- 新建文章
- 编辑文章
- 删除文章

编辑器当前支持：

- 标题、摘要、分类、标签、封面、状态、置顶、正文编辑
- Markdown 编辑
- 分栏预览 / 仅预览
- 常用 Markdown 快捷插入：
  - 粗体
  - 斜体
  - 标题
  - 列表
  - 引用
  - 行内代码
  - 代码块
  - 链接
- 插入图片
- 字数统计与预计阅读时长显示
- `Ctrl/Cmd + S` 保存到服务器

### 3.4 本地草稿保护

文章编辑器已实现本地草稿能力：

- 自动保存到 `localStorage`
- 新建文章与编辑文章使用不同草稿 key
- 重新打开编辑器时可恢复本地草稿
- 支持忽略或清除本地草稿
- 成功保存到服务器后自动清理本地草稿

### 3.5 素材管理

素材页当前支持：

- 上传图片
- 拖拽上传
- 多图上传
- 刷新素材列表
- 预览图片
- 复制图片 URL
- 单张删除
- 批量删除
- 分页浏览素材

同时，文章编辑器内置图片选择器，可直接：

- 选图作为封面
- 插入图片到 Markdown 正文
- 上传后刷新素材列表

### 3.6 分类管理

分类页当前支持：

- 分类列表查询
- 新建分类
- 编辑分类
- 删除分类

---

## 4. 项目结构

```text
admin-vue/
├── src/
│   ├── api/                    # 接口封装
│   ├── components/             # 复用组件
│   │   ├── article/            # 文章编辑器等组件
│   │   └── layout/             # 后台布局组件
│   ├── router/                 # 路由配置与守卫
│   ├── stores/                 # Pinia 状态管理
│   ├── styles/                 # 全局样式与主题样式
│   ├── types/                  # TypeScript 类型定义
│   ├── utils/                  # 工具函数
│   └── views/                  # 页面视图
├── public/
├── Dockerfile
├── nginx.conf
├── package.json
└── vite.config.ts
```

重点目录说明：

- `src/router/index.ts`：后台路由、登录保护、强制改密跳转
- `src/stores/auth.ts`：登录状态、token、强制改密状态管理
- `src/api/client.ts`：Axios 实例、请求头注入、401/403 处理
- `src/views/articles/ArticleListView.vue`：文章列表页
- `src/components/article/ArticleEditorDialog.vue`：文章编辑器核心组件
- `src/views/images/ImageLibraryView.vue`：素材管理页
- `src/views/categories/CategoryListView.vue`：分类管理页

---

## 5. 启动开发

安装依赖：

```bash
npm install
```

启动开发环境：

```bash
npm run dev
```

构建生产包：

```bash
npm run build
```

本地预览构建结果：

```bash
npm run preview
```

运行测试：

```bash
npm test
```

---

## 6. API 配置

前端通过 `VITE_API_BASE` 指定后端 API 地址。

当前 Axios 客户端配置位于：

- `src/api/client.ts`

示例：

```bash
VITE_API_BASE=http://172.28.74.191:31083/api
```

如果未正确配置 API 地址，后台即使页面能打开，也会进入“UI 很努力，接口不理你”的尴尬状态。

---

## 7. 路由说明

当前主要路由：

```text
/login
/change-password
/dashboard
/articles
/categories
/images
```

路由规则：

- 未登录访问受保护页面时跳转 `/login`
- 已登录但被要求强制改密时跳转 `/change-password`
- 已登录后访问 `/login` 会被重定向到工作台或改密页

---

## 8. 构建与部署

项目采用多阶段 Docker 构建：

1. 使用 `node:20-alpine` 执行 `npm install` 和 `npm run build`
2. 使用 `nginx:alpine` 承载构建产物

对应文件：

- `Dockerfile`
- `nginx.conf`

在主仓库中，`deploy.sh` 会统一构建 `admin-vue` 镜像并配合 `k8s/admin-vue.yaml` 部署到 K3s。

当前 K8s Service 暴露端口：

- NodePort: `31085`

---

## 9. 当前定位与后续方向

`admin-vue` 当前的产品主线不是“功能越多越酷炫”，而是：

1. 稳定跑通真实后端接口
2. 优先保障文章创作链路
3. 提升 Markdown 写作效率
4. 提升图床联动效率
5. 降低内容丢失风险

因此，当前更值得继续投入的方向包括：

- 编辑器体验继续增强
- 草稿保护细节优化
- 素材选择与插图效率优化
- 真实接口联调稳定性提升
- 子模块文档补齐

---

## 10. 相关项目关系

`admin-vue` 不是孤立项目，它属于 `blog-butterfly-go` 的整体架构之一：

- `web-vue`：博客前台
- `admin-vue`：博客后台
- `backend`：Go API
- `k8s/`：部署清单

如果要理解完整上下文，建议同时查看仓库根 README：

- `../README.md`

---

## 11. 备注

如果你看到这份 README，说明之前那份 “Vue 3 + TypeScript + Vite 模板欢迎你” 已经光荣退休了。它没有做错什么，只是这项目已经开始认真上班了 😼
