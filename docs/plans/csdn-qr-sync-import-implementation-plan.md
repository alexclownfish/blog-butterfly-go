# CSDN 扫码登录同步导入功能实施方案

> **For Hermes:** Use subagent-driven-development skill to implement this plan task-by-task.

**Goal:** 在现有 `admin-vue` 后台中新增一个“CSDN 同步导入中心”，支持管理员扫码登录 CSDN、读取本人文章列表、预览单篇文章并按分类/状态选择性导入到博客后台，同时保留现有 URL 直贴导入能力。

**Architecture:** 保持当前 `/api/articles/import/csdn/*` 的 URL 解析导入链路不动，新增一套 `CSDN Sync` 会话层。后端负责生成二维码、轮询登录状态、保存短期 CSDN 登录态、拉取文章列表/详情并转换成博客文章导入载荷；前端新增独立页面 `CsdnSyncView` 承载登录、列表、预览、导入操作，不把复杂流程继续堆进现有 `CsdnImportDialog`。所有新增能力都围绕“增强导入成功率”设计，不承诺一定拿到 markdown 原稿；若 CSDN 仅返回 HTML/富文本，则统一在服务层转换为当前系统可接受的 Markdown/正文格式。

**Tech Stack:** Vue 3 + TypeScript + Element Plus、Vue Router、Axios、Go + Gin、GORM、现有 JWT 管理后台认证、服务端短期会话存储（首期内存 + TTL，可演进为数据库表/Redis）、可选 HTTP client + HTML/JSON 解析。

---

## 0. 当前代码事实（基于现有仓库）

### 已存在的前端能力
- `admin-vue/src/views/articles/ArticleListView.vue`
  - 当前文章管理页已集成 `CsdnImportDialog`，提供“粘贴 URL → 解析预览 → 导入”的轻量流程。
- `admin-vue/src/components/article/CsdnImportDialog.vue`
  - 已有 URL 预览与导入 UI，不适合继续承载扫码登录、登录态轮询、文章列表、批量选择等复杂交互。
- `admin-vue/src/router/index.ts`
  - 当前已注册 `/dashboard`、`/articles`、`/categories`、`/images` 等后台路由，可按相同模式新增 `articles/csdn-sync` 页面。
- `admin-vue/src/components/layout/AdminSidebar.vue`
  - 当前左侧导航是硬编码 `RouterLink`，需要手动增加新入口。
- `admin-vue/src/api/articles.ts`
  - 已封装 `previewImportCsdnApi()` 与 `importCsdnArticleApi()`，适合作为“快捷 URL 导入”保留。

### 已存在的后端能力
- `backend/router/router.go`
  - 当前只有：
    - `POST /api/articles/import/csdn/preview`
    - `POST /api/articles/import/csdn`
- `backend/controllers/csdn.go`
  - 当前控制器只处理“给 URL，返回预览/直接导入”的场景。
- `backend/services/csdn.go`
  - 当前是“实时抓公开文章 HTML → 解析结构 → 返回 `CSDNArticle`”。
  - 这是公开页抓取路线，不具备登录态管理、个人文章列表、私有接口代理能力。
- `backend/models/article.go`
  - 现有 `Article` 结构足够承接最终导入结果，首期不必修改文章表。

### 明确约束
1. **首期目标不是替换 URL 导入，而是增加扫码登录增强模式。**
2. **首期不保证拿到 markdown 原稿。** 若登录后接口仍返回 HTML，则继续在服务层做转换。
3. **不要让 admin-vue 前端直接处理 CSDN cookie。** 登录态必须由后端接管。
4. **当前后台是单管理员风格，但设计上仍要按“后台用户隔离自己的 CSDN 会话”来写。**
5. **CSDN 私有登录/文章接口可能随时变动，必须把适配逻辑集中在服务层。**

---

# Phase A — 需求收敛与数据模型预留

### Task A1: 明确首期范围与非目标

**Objective:** 把“要做什么 / 不做什么”写清楚，防止功能膨胀成大杂烩。

**Files:**
- Create: `docs/plans/csdn-qr-sync-import-implementation-plan.md`（本文档，已完成）
- Optional update later: `docs/PRD-admin-content-workspace-p2.md`（如需在更高层文档提及）

**Step 1: 固定首期范围**
- 必做：
  - 扫码登录 CSDN
  - 轮询登录状态
  - 获取文章列表
  - 查看单篇预览
  - 选择分类/状态导入
  - 退出/清理登录态
- 不做：
  - 批量导入多篇
  - 定时自动同步
  - 自动去重/增量同步
  - 100% markdown 源文稿还原承诺

**Step 2: 固定 UI 定位**
- `CsdnImportDialog.vue` 继续负责“快速 URL 导入”。
- 新建独立页面 `CsdnSyncView.vue` 负责“扫码登录同步导入”。

**Verification:**
- 文档中明确分开“快捷 URL 导入”和“同步我的 CSDN”。

**Commit message:**
```bash
git commit -m "docs: add csdn qr sync import implementation plan"
```

---

### Task A2: 预留后端会话模型

**Objective:** 为 CSDN 登录会话与远端文章元数据定义稳定结构，避免控制器里到处塞 `map[string]any`。

**Files:**
- Create: `backend/services/csdn_sync_types.go`
- Optional later: `backend/models/csdn_session.go`（若第二阶段落库）
- Test: `backend/services/csdn_sync_types_test.go`

**Step 1: 定义登录会话结构**

```go
package services

import "time"

type CSDNSessionStatus string

const (
    CSDNSessionStatusPending   CSDNSessionStatus = "pending"
    CSDNSessionStatusScanned   CSDNSessionStatus = "scanned"
    CSDNSessionStatusConfirmed CSDNSessionStatus = "confirmed"
    CSDNSessionStatusExpired   CSDNSessionStatus = "expired"
    CSDNSessionStatusFailed    CSDNSessionStatus = "failed"
)

type CSDNLoginSession struct {
    SessionID     string            `json:"session_id"`
    AdminUserID   uint              `json:"-"`
    Status        CSDNSessionStatus `json:"status"`
    QRCodeURL     string            `json:"qr_code_url,omitempty"`
    QRCodeRaw     string            `json:"qr_code_raw,omitempty"`
    PollToken     string            `json:"-"`
    Cookies       []*http.Cookie    `json:"-"`
    Nickname      string            `json:"nickname,omitempty"`
    Avatar        string            `json:"avatar,omitempty"`
    ExpiresAt     time.Time         `json:"expires_at"`
    LastError     string            `json:"last_error,omitempty"`
    CreatedAt     time.Time         `json:"created_at"`
    UpdatedAt     time.Time         `json:"updated_at"`
}
```

**Step 2: 定义远端文章结构**

```go
type CSDNRemoteArticle struct {
    ID           string    `json:"id"`
    Title        string    `json:"title"`
    Summary      string    `json:"summary,omitempty"`
    CoverImage   string    `json:"cover_image,omitempty"`
    Tags         string    `json:"tags,omitempty"`
    Status       string    `json:"status,omitempty"`
    URL          string    `json:"url"`
    CreatedAt    time.Time `json:"created_at,omitempty"`
    UpdatedAt    time.Time `json:"updated_at,omitempty"`
}

type CSDNRemoteArticleDetail struct {
    Article CSDNRemoteArticle `json:"article"`
    Content string            `json:"content"`
    Format  string            `json:"format"` // markdown | html | richtext
}
```

**Step 3: 定义首期内存仓储接口**

```go
type CSDNSessionStore interface {
    Save(session *CSDNLoginSession) error
    Get(sessionID string) (*CSDNLoginSession, error)
    Delete(sessionID string) error
    ListByAdminUser(adminUserID uint) ([]*CSDNLoginSession, error)
    CleanupExpired(now time.Time) error
}
```

**Verification:**
- `go test ./backend/services -run CSDN` 能编译通过。
- 所有控制器/服务层都不再依赖匿名 map 结构传递登录态。

**Commit message:**
```bash
git commit -m "feat(csdn): add sync session and remote article types"
```

---

# Phase B — 后端登录会话与 CSDN 服务层骨架

### Task B1: 封装 CSDN Sync 服务入口

**Objective:** 把“生成二维码 / 查询状态 / 获取文章列表 / 获取详情”统一塞到一个服务里，不让控制器知道 CSDN 私有协议细节。

**Files:**
- Create: `backend/services/csdn_sync_service.go`
- Modify: `backend/services/csdn.go`（必要时抽公共转换函数）
- Test: `backend/services/csdn_sync_service_test.go`

**Step 1: 定义服务接口**

```go
type CSDNSyncService interface {
    CreateLoginSession(adminUserID uint) (*CSDNLoginSession, error)
    GetLoginSession(adminUserID uint, sessionID string) (*CSDNLoginSession, error)
    RefreshLoginSession(adminUserID uint, sessionID string) (*CSDNLoginSession, error)
    Logout(adminUserID uint, sessionID string) error
    ListArticles(adminUserID uint, sessionID string, page int, keyword string) ([]CSDNRemoteArticle, int, error)
    GetArticleDetail(adminUserID uint, sessionID string, articleID string) (*CSDNRemoteArticleDetail, error)
}
```

**Step 2: 实现首期 stub 版本**
- 第一步先让服务层可编译、可被控制器调用：
  - `CreateLoginSession()` 先生成 UUID session_id
  - 先预留二维码字段
  - `RefreshLoginSession()` 先读仓储后返回
- 不要一上来把所有 CSDN 登录逆向细节写死在控制器。

**Step 3: 抽公共内容转换函数**
- 把现有 `services/csdn.go` 中“HTML/正文 → 系统 content”的逻辑抽成可复用 helper，例如：

```go
func NormalizeCSDNContent(raw string, format string) string
```

这样以后：
- 公开 URL 抓取结果可以用
- 登录后拿到 HTML/richtext 也可以复用

**Verification:**
- 控制器可通过接口调用服务；服务返回结构稳定。
- `go test ./backend/services -run CSDNSync` 通过。

**Commit message:**
```bash
git commit -m "feat(csdn): add sync service interface and skeleton"
```

---

### Task B2: 增加内存会话仓储与 TTL 清理

**Objective:** 首期先用内存会话，避免为了试点功能先上数据库迁移。

**Files:**
- Create: `backend/services/csdn_session_store_memory.go`
- Test: `backend/services/csdn_session_store_memory_test.go`

**Step 1: 实现线程安全内存仓储**

```go
type MemoryCSDNSessionStore struct {
    mu       sync.RWMutex
    sessions map[string]*CSDNLoginSession
}
```

**Step 2: 写 Save/Get/Delete/ListByAdminUser/CleanupExpired**
- `Save()` 时深拷贝，避免外部直接改指针。
- `Get()` 时若过期，返回 `expired` 或直接删除后报错。
- `ListByAdminUser()` 只返回归属当前后台用户的会话。

**Step 3: 增加后台清理策略**
- 简化版：每次 `Create/Get/List` 时顺手 `CleanupExpired(time.Now())`
- TTL 建议：15 分钟未完成登录自动过期；已登录会话 2 小时过期。

**Verification:**
- 并发测试下不会 data race。
- 过期会话无法继续拿文章列表。

**Commit message:**
```bash
git commit -m "feat(csdn): add in-memory login session store with ttl"
```

---

### Task B3: 对接 CSDN 登录流程适配器

**Objective:** 把真实 CSDN 私有接口适配封在单独 adapter 中，未来结构变化时只改这一层。

**Files:**
- Create: `backend/services/csdn_adapter.go`
- Create: `backend/services/csdn_adapter_test.go`

**Step 1: 定义 adapter 接口**

```go
type CSDNAdapter interface {
    CreateQRCode(ctx context.Context) (qrCodeURL string, pollToken string, expiresAt time.Time, err error)
    CheckQRCodeStatus(ctx context.Context, pollToken string) (status CSDNSessionStatus, cookies []*http.Cookie, profile *CSDNProfile, err error)
    ListUserArticles(ctx context.Context, cookies []*http.Cookie, page int, keyword string) ([]CSDNRemoteArticle, int, error)
    GetUserArticleDetail(ctx context.Context, cookies []*http.Cookie, articleID string) (*CSDNRemoteArticleDetail, error)
}
```

**Step 2: 先做 mockable 实现**
- 先不要把真实 CSDN 接口写死到测试里。
- 用 `httptest.Server` 模拟：
  - 二维码创建返回
  - 状态轮询返回
  - 文章列表返回
  - 文章详情返回

**Step 3: 再接真实 CSDN**
- 真实实现注意：
  - UA / Referer / Accept Header
  - 登录成功后的 cookie 组合
  - 接口失败重试
  - 限流和超时控制

**Verification:**
- 所有控制器测试只 mock adapter，不依赖外网。
- 真实实现失败时可返回明确 `LastError`，而不是一坨 500。

**Commit message:**
```bash
git commit -m "feat(csdn): add adapter layer for qr login and article sync"
```

---

# Phase C — 后端 API 设计与控制器落地

### Task C1: 新增控制器请求/响应结构

**Objective:** 为扫码登录导入路线提供独立 API，不污染现有 URL 导入控制器。

**Files:**
- Modify: `backend/controllers/csdn.go`
- Test: `backend/controllers/csdn_sync_test.go`

**Step 1: 新增请求/响应 DTO**

```go
type csdnCreateSessionResponse struct {
    Data *services.CSDNLoginSession `json:"data"`
}

type csdnSessionQuery struct {
    SessionID string `uri:"session_id" binding:"required"`
}

type csdnArticlesQuery struct {
    SessionID string `form:"session_id" binding:"required"`
    Page      int    `form:"page"`
    Keyword   string `form:"keyword"`
}

type csdnImportByIDRequest struct {
    SessionID  string `json:"session_id" binding:"required"`
    ArticleID  string `json:"article_id" binding:"required"`
    CategoryID uint   `json:"category_id" binding:"required"`
    Status     string `json:"status"`
}
```

**Step 2: 新增控制器方法**
- `CreateCSDNLoginSession`
- `GetCSDNLoginSessionStatus`
- `LogoutCSDNSession`
- `ListCSDNArticles`
- `GetCSDNArticleDetail`
- `ImportCSDNArticleByID`

**Step 3: 统一错误语义**
- 400：参数错误
- 401：后台未登录
- 403：会话不属于当前后台用户
- 404：session/article 不存在
- 409：二维码已过期或尚未完成登录
- 502：CSDN 远端接口异常

**Verification:**
- 每个控制器都有单测覆盖 happy path + 参数错误 + 权限隔离。

**Commit message:**
```bash
git commit -m "feat(csdn): add qr sync controllers and request models"
```

---

### Task C2: 路由注册新 API

**Objective:** 将新接口挂到现有 Gin 路由中，并沿用后台 JWT 认证。

**Files:**
- Modify: `backend/router/router.go`
- Test: `backend/controllers/csdn_sync_test.go`

**Step 1: 在 auth 组新增路由**

```go
auth.POST("/csdn/auth/qrcode", controllers.CreateCSDNLoginSession)
auth.GET("/csdn/auth/status/:session_id", controllers.GetCSDNLoginSessionStatus)
auth.POST("/csdn/auth/logout", controllers.LogoutCSDNSession)
auth.GET("/csdn/articles", controllers.ListCSDNArticles)
auth.GET("/csdn/articles/:article_id", controllers.GetCSDNArticleDetail)
auth.POST("/csdn/articles/import", controllers.ImportCSDNArticleByID)
```

**Step 2: 保留旧接口**
- 不删除：
  - `/articles/import/csdn/preview`
  - `/articles/import/csdn`

这样形成：
- 快捷 URL 导入
- 登录增强同步导入

**Verification:**
- 路由表包含两套能力，互不覆盖。

**Commit message:**
```bash
git commit -m "feat(router): register csdn qr sync routes"
```

---

### Task C3: 用文章 ID 导入到本地博客系统

**Objective:** 当用户在同步页点“导入”时，不再依赖公开 URL，而是直接从当前登录会话拉详情并入库。

**Files:**
- Modify: `backend/controllers/csdn.go`
- Modify: `backend/services/csdn_sync_service.go`
- Test: `backend/controllers/csdn_sync_test.go`

**Step 1: 增加“按 article_id 导入”逻辑**

```go
func ImportCSDNArticleByID(c *gin.Context) {
    // 1. Bind JSON
    // 2. 校验文章状态 draft/published
    // 3. 调 service.GetArticleDetail(adminUserID, sessionID, articleID)
    // 4. 转换成 models.Article
    // 5. DB.Create
    // 6. Preload Category 后返回
}
```

**Step 2: 统一导入映射规则**
- 标题：`detail.Article.Title`
- 摘要：`detail.Article.Summary`
- 封面：`detail.Article.CoverImage`
- 标签：`detail.Article.Tags`
- 正文：`NormalizeCSDNContent(detail.Content, detail.Format)`
- 分类/状态：由本地后台选择

**Step 3: 首期不做去重，但预留扩展点**
- 可在返回里附加 `source_url`
- 未来如果要防重复，可基于 `source_url` 或远端文章 ID 做唯一性检查

**Verification:**
- 导入成功后 `GET /api/articles/:id` 可读取到完整数据。
- article_id 错误时返回明确错误。

**Commit message:**
```bash
git commit -m "feat(csdn): support importing synced article by remote id"
```

---

# Phase D — admin-vue API 层与类型封装

### Task D1: 增加前端类型定义

**Objective:** 给新页面一个完整的类型地基，避免到处写 `any`。

**Files:**
- Modify: `admin-vue/src/types/article.ts`
- Or create: `admin-vue/src/types/csdn.ts`
- Test: `admin-vue/src/api/__tests__/csdn.spec.ts`（如项目测试结构允许）

**Step 1: 推荐新建独立类型文件**

```ts
export type CsdnSessionStatus = 'pending' | 'scanned' | 'confirmed' | 'expired' | 'failed'

export interface CsdnLoginSession {
  session_id: string
  status: CsdnSessionStatus
  qr_code_url?: string
  qr_code_raw?: string
  nickname?: string
  avatar?: string
  expires_at?: string
  created_at?: string
  updated_at?: string
  last_error?: string
}

export interface CsdnRemoteArticle {
  id: string
  title: string
  summary?: string
  cover_image?: string
  tags?: string
  status?: string
  url: string
  created_at?: string
  updated_at?: string
}

export interface CsdnRemoteArticleDetail {
  article: CsdnRemoteArticle
  content: string
  format: 'markdown' | 'html' | 'richtext'
}
```

**Step 2: 不污染现有 `Article` 本地类型**
- 远端 CSDN 文章 ≠ 本地博客文章
- 分开建模，减少语义混淆

**Verification:**
- `tsc` 无类型冲突
- 页面中可精准区分“远端文章”和“本地文章”

**Commit message:**
```bash
git commit -m "feat(admin-vue): add csdn sync type models"
```

---

### Task D2: 新增前端 API 封装

**Objective:** 为 `CsdnSyncView` 提供清晰 API，避免组件里手写 axios 路径字符串。

**Files:**
- Create: `admin-vue/src/api/csdn.ts`
- Modify: `admin-vue/src/api/articles.ts`（只保留旧 URL 导入，不要硬塞新接口进去）
- Test: `admin-vue/src/api/client.test.ts` 或新增 `admin-vue/src/api/csdn.test.ts`

**Step 1: 增加以下 API 方法**

```ts
import client from './client'
import type { CsdnLoginSession, CsdnRemoteArticle, CsdnRemoteArticleDetail } from '@/types/csdn'
import type { Article, ArticleStatus } from '@/types/article'

export async function createCsdnLoginSessionApi(): Promise<CsdnLoginSession> {}
export async function fetchCsdnLoginSessionStatusApi(sessionId: string): Promise<CsdnLoginSession> {}
export async function logoutCsdnSessionApi(sessionId: string): Promise<void> {}
export async function fetchCsdnArticlesApi(params: { session_id: string; page?: number; keyword?: string }): Promise<{ list: CsdnRemoteArticle[]; total: number }> {}
export async function fetchCsdnArticleDetailApi(articleId: string, sessionId: string): Promise<CsdnRemoteArticleDetail> {}
export async function importCsdnSyncedArticleApi(payload: {
  session_id: string
  article_id: string
  category_id: number
  status: ArticleStatus
}): Promise<Article> {}
```

**Step 2: API 返回值统一做 data 解包**
- 保持和现有 `articles.ts` 一样的风格
- 不让页面直接关心 `{ data: { ... } }`

**Step 3: 错误信息尽量标准化**
- 未扫码：`二维码已过期，请重新生成`
- 未确认：`请先完成 CSDN 扫码登录`
- 远端失败：`获取 CSDN 文章列表失败`

**Verification:**
- 单测验证所有 API 正确命中路径和参数。

**Commit message:**
```bash
git commit -m "feat(admin-vue): add csdn sync api client"
```

---

# Phase E — admin-vue 页面与交互实现

### Task E1: 注册新路由与侧边栏入口

**Objective:** 让“CSDN 同步导入中心”成为正式后台页面，而不是隐藏能力。

**Files:**
- Modify: `admin-vue/src/router/index.ts`
- Modify: `admin-vue/src/components/layout/AdminSidebar.vue`
- Create: `admin-vue/src/views/articles/CsdnSyncView.vue`
- Test: `admin-vue/src/router/__tests__/index.spec.ts`（如果已有）

**Step 1: 新增路由**

```ts
{
  path: 'articles/csdn-sync',
  name: 'articles-csdn-sync',
  component: () => import('@/views/articles/CsdnSyncView.vue'),
  meta: {
    title: 'CSDN 同步导入'
  }
}
```

**Step 2: 侧边栏新增入口**
建议放在“文章管理”后面：

```vue
<RouterLink to="/articles/csdn-sync" class="sidebar-link">
  <span>🔄</span>
  <span>CSDN 同步导入</span>
</RouterLink>
```

**Step 3: 保持当前文章页入口**
- `ArticleListView.vue` 里的“导入 CSDN”按钮继续保留，定位为快捷导入。

**Verification:**
- 登录后能从侧边栏进入新页面。
- 页面标题正确显示在浏览器 tab。

**Commit message:**
```bash
git commit -m "feat(admin-vue): add csdn sync route and sidebar entry"
```

---

### Task E2: 实现页面整体布局

**Objective:** 先把信息架子搭出来，再填业务逻辑。

**Files:**
- Create: `admin-vue/src/views/articles/CsdnSyncView.vue`
- Optional style split later: `admin-vue/src/views/articles/csdn-sync.css`（如果项目后续要拆）
- Test: `admin-vue/src/views/articles/__tests__/CsdnSyncView.spec.ts`

**Step 1: 页面分三栏/三段**
建议布局：
1. 顶部说明区
   - 区分“快捷 URL 导入”和“同步我的 CSDN”
2. 登录状态卡片
   - 未登录：二维码 + 刷新状态按钮
   - 已登录：头像/昵称/会话状态/退出按钮
3. 内容区
   - 左侧：文章列表 + 搜索
   - 右侧：文章预览 + 导入配置

**Step 2: 初始模板示例**

```vue
<section class="page-section csdn-sync-page">
  <div class="panel-card">
    <div class="section-head">
      <div>
        <div class="card-eyebrow">🔄 CSDN Sync</div>
        <h2>CSDN 同步导入</h2>
        <p>扫码登录你的 CSDN 账号后，拉取文章列表并按需导入到博客后台。</p>
      </div>
      <el-button @click="goBackToArticles">返回文章管理</el-button>
    </div>

    <!-- 登录状态卡 -->
    <!-- 列表 + 预览区 -->
  </div>
</section>
```

**Step 3: 页面状态变量**

```ts
const session = ref<CsdnLoginSession | null>(null)
const sessionLoading = ref(false)
const polling = ref(false)
const articleLoading = ref(false)
const detailLoading = ref(false)
const importLoading = ref(false)
const articles = ref<CsdnRemoteArticle[]>([])
const selectedArticleId = ref('')
const articleDetail = ref<CsdnRemoteArticleDetail | null>(null)
const keyword = ref('')
const page = ref(1)
const total = ref(0)
const importForm = reactive({
  category_id: null as number | null,
  status: 'draft' as ArticleStatus
})
```

**Verification:**
- 页面空状态渲染正常。
- 未接真实接口时可先通过 mock 数据驱动基本 UI。

**Commit message:**
```bash
git commit -m "feat(admin-vue): scaffold csdn sync import view"
```

---

### Task E3: 实现扫码登录与状态轮询

**Objective:** 让用户能真正看到二维码，并在扫码成功后自动进入文章列表阶段。

**Files:**
- Modify: `admin-vue/src/views/articles/CsdnSyncView.vue`
- Create optional helper: `admin-vue/src/composables/usePolling.ts`
- Test: `admin-vue/src/views/articles/__tests__/CsdnSyncView.spec.ts`

**Step 1: 点击生成二维码**

```ts
async function createSession() {
  sessionLoading.value = true
  try {
    session.value = await createCsdnLoginSessionApi()
    startPolling()
  } finally {
    sessionLoading.value = false
  }
}
```

**Step 2: 轮询登录状态**
- 每 2~3 秒请求一次状态
- 到以下状态停止轮询：
  - `confirmed`
  - `expired`
  - `failed`

```ts
let timer: number | null = null

function startPolling() {
  stopPolling()
  timer = window.setInterval(async () => {
    if (!session.value?.session_id) return
    const next = await fetchCsdnLoginSessionStatusApi(session.value.session_id)
    session.value = next
    if (['confirmed', 'expired', 'failed'].includes(next.status)) {
      stopPolling()
      if (next.status === 'confirmed') {
        await loadArticles()
      }
    }
  }, 3000)
}
```

**Step 3: 页面卸载时清理轮询**
- `onBeforeUnmount(stopPolling)`

**Step 4: 状态文案明确**
- `pending`: 等待扫码
- `scanned`: 已扫码，请在手机上确认
- `confirmed`: 登录成功
- `expired`: 二维码已过期
- `failed`: 登录失败

**Verification:**
- 轮询不会重复开启多个定时器。
- 切走页面后不会继续轮询。

**Commit message:**
```bash
git commit -m "feat(admin-vue): add csdn qr login polling flow"
```

---

### Task E4: 实现文章列表与搜索

**Objective:** 登录成功后展示 CSDN 文章列表，支持翻页和关键字过滤。

**Files:**
- Modify: `admin-vue/src/views/articles/CsdnSyncView.vue`
- Test: `admin-vue/src/views/articles/__tests__/CsdnSyncView.spec.ts`

**Step 1: 增加搜索栏与刷新按钮**
- 搜索关键字
- 刷新列表
- 当前账号昵称/状态提示

**Step 2: 列表项建议字段**
- 标题
- 摘要前 1~2 行
- 更新时间
- 远端状态（已发布/草稿，如拿得到）
- “预览”按钮

**Step 3: 点击列表项加载详情**

```ts
async function selectArticle(articleId: string) {
  if (!session.value?.session_id) return
  selectedArticleId.value = articleId
  detailLoading.value = true
  try {
    articleDetail.value = await fetchCsdnArticleDetailApi(articleId, session.value.session_id)
  } finally {
    detailLoading.value = false
  }
}
```

**Verification:**
- 登录成功后自动加载第一页文章。
- 搜索和翻页会带上当前 session_id。

**Commit message:**
```bash
git commit -m "feat(admin-vue): add csdn synced article list and search"
```

---

### Task E5: 实现单篇预览与导入配置区

**Objective:** 让用户在导入前看清楚标题/摘要/封面/正文，并选择本地分类与状态。

**Files:**
- Modify: `admin-vue/src/views/articles/CsdnSyncView.vue`
- Reuse: `admin-vue/src/api/categories.ts`
- Test: `admin-vue/src/views/articles/__tests__/CsdnSyncView.spec.ts`

**Step 1: 页面挂载时加载分类**
- 与 `ArticleListView.vue` 一样加载分类列表
- 默认选中第一项分类

**Step 2: 右侧预览区展示**
- 标题
- 来源 URL
- tags
- cover_image
- content 前若干行 / Markdown 预览区
- 如果 `format === 'html'`，显示提示：
  - “该文章由 HTML/富文本转换而来，导入后建议人工检查排版”

**Step 3: 导入按钮逻辑**

```ts
async function handleImport() {
  if (!session.value?.session_id || !selectedArticleId.value || !importForm.category_id) {
    ElMessage.error('请先完成登录、选中文章并选择分类')
    return
  }
  importLoading.value = true
  try {
    const article = await importCsdnSyncedArticleApi({
      session_id: session.value.session_id,
      article_id: selectedArticleId.value,
      category_id: importForm.category_id,
      status: importForm.status
    })
    ElMessage.success(`导入成功：${article.title}`)
  } finally {
    importLoading.value = false
  }
}
```

**Step 4: 导入后用户反馈**
- 成功后给两个快捷动作：
  - “去文章管理查看”
  - “继续导入下一篇”

**Verification:**
- 未选分类时禁止导入。
- 导入成功后提示准确，且不会清空整个登录态。

**Commit message:**
```bash
git commit -m "feat(admin-vue): add csdn article preview and import panel"
```

---

### Task E6: 实现退出登录与异常状态兜底

**Objective:** 避免用户卡在“二维码过期/会话失效/远端失败”的半瘫状态。

**Files:**
- Modify: `admin-vue/src/views/articles/CsdnSyncView.vue`
- Test: `admin-vue/src/views/articles/__tests__/CsdnSyncView.spec.ts`

**Step 1: 增加退出登录按钮**
- 调 `logoutCsdnSessionApi(session_id)`
- 成功后清空：
  - session
  - articles
  - articleDetail
  - selectedArticleId

**Step 2: 异常兜底文案**
- 二维码过期：显示“重新生成二维码”按钮
- 远端文章加载失败：保留列表，不清空 session
- token/session 无效：提醒重新登录 CSDN

**Step 3: 切换页面后的恢复策略**
- 首期可不做持久恢复；刷新页面后重新扫码即可
- 如果想增强体验，可把最近 `session_id` 放在 `sessionStorage`，但必须在后端再次校验归属与有效期

**Verification:**
- 点击退出后不再能访问旧 session 的文章接口。
- 二维码过期后可重新生成新会话。

**Commit message:**
```bash
git commit -m "feat(admin-vue): add csdn logout and expired-session recovery"
```

---

# Phase F — 测试、联调与验收

### Task F1: 后端单元测试与控制器测试

**Objective:** 保证未来 CSDN 接口一变，至少能快速定位是 adapter 挂了还是控制器挂了。

**Files:**
- Create/Modify:
  - `backend/services/csdn_sync_service_test.go`
  - `backend/services/csdn_session_store_memory_test.go`
  - `backend/services/csdn_adapter_test.go`
  - `backend/controllers/csdn_sync_test.go`

**Step 1: 测服务层**
- session 创建
- session 过期
- 未登录禁止拉文章
- article detail 转换

**Step 2: 测控制器**
- 创建二维码
- 查状态
- 列表分页
- 导入成功
- 非法 session / 非法 article_id

**Step 3: 跑命令**

```bash
cd /root/blog-butterfly-go/backend
go test ./... 
```

**Expected:**
- 所有新增测试通过

**Commit message:**
```bash
git commit -m "test(csdn): cover qr sync service and controller flows"
```

---

### Task F2: 前端组件测试

**Objective:** 确保页面状态机不会因为轮询/切换状态而抽风。

**Files:**
- Create: `admin-vue/src/views/articles/__tests__/CsdnSyncView.spec.ts`
- Create: `admin-vue/src/api/csdn.test.ts`（可选）

**Step 1: 覆盖以下场景**
- 未登录显示“生成二维码”
- 生成二维码后显示二维码卡片
- 状态从 `pending` → `confirmed` 后自动加载文章列表
- 选择文章后显示预览
- 点击导入调用正确 API
- 退出登录后清空状态

**Step 2: 跑测试**

```bash
cd /root/blog-butterfly-go/admin-vue
npm run test -- CsdnSyncView
```

**Expected:**
- 页面测试通过

**Commit message:**
```bash
git commit -m "test(admin-vue): cover csdn sync view flows"
```

---

### Task F3: 真实接口联调清单

**Objective:** 在可达后端环境中做真实验收，而不是只看本地 mock。

**Files:**
- No code required first
- Optional doc: `docs/admin-qa-round-YYYY-MM-DD.md`

**Step 1: 登录验收**
- 打开 `/articles/csdn-sync`
- 生成二维码
- 手机扫码确认
- 页面状态变为已登录

**Step 2: 列表验收**
- 看到本人文章列表
- 搜索能筛选
- 翻页正常

**Step 3: 导入验收**
- 选择一篇文章
- 预览内容
- 选择分类 + 草稿状态
- 导入成功
- 在文章管理页看到新文章
- 打开详情确认正文/封面/标签

**Step 4: 异常验收**
- 二维码过期后重新生成
- 退出登录后无法继续拉列表
- 故意断网/模拟远端失败时能看到友好错误

**Verification:**
- 至少用 1 篇公开文章 + 1 篇需要登录态的本人文章验收。

**Commit:**
- 如仅联调，无需代码提交。

---

# Phase G — 第二阶段演进（本期不做，但要留钩子）

### Task G1: 会话持久化到数据库/Redis
- 当需要多实例部署或页面刷新恢复时，再引入：
  - `models.CSDNSession`
  - DB migration 或 Redis TTL

### Task G2: 批量导入
- 增加复选框 + 批量导入任务队列
- 注意失败回滚和逐篇错误报告

### Task G3: 导入去重
- 可按 `source_url` / `remote_article_id` 做幂等检查

### Task G4: Markdown 原稿专项研究
- 仅在确认 CSDN 登录后存在稳定 markdown 源接口时再做
- 若拿不到，就继续优化 HTML → Markdown 转换质量

---

# 推荐的接口清单（汇总）

## 旧接口：保留
- `POST /api/articles/import/csdn/preview`
- `POST /api/articles/import/csdn`

## 新接口：新增
- `POST /api/csdn/auth/qrcode`
- `GET /api/csdn/auth/status/:session_id`
- `POST /api/csdn/auth/logout`
- `GET /api/csdn/articles?session_id=...&page=1&keyword=...`
- `GET /api/csdn/articles/:article_id?session_id=...`
- `POST /api/csdn/articles/import`

推荐导入请求体：

```json
{
  "session_id": "csdn_session_xxx",
  "article_id": "123456789",
  "category_id": 2,
  "status": "draft"
}
```

---

# 推荐的前端信息架构（汇总）

## 页面入口
- 侧边栏：`CSDN 同步导入`
- 文章管理页按钮：继续保留 `导入 CSDN`（快捷 URL 导入）

## 页面分区
1. 顶部说明区
2. 登录状态/二维码区
3. 左侧文章列表区
4. 右侧文章预览与导入配置区

---

# 风险与应对

## 风险 1：CSDN 私有登录接口变动
**应对：** 所有逆向适配逻辑收敛到 `csdn_adapter.go`，不要散落到 controller / view。

## 风险 2：登录后仍拿不到 markdown 原稿
**应对：** 产品文案明确为“同步导入/增强导入”，不是“还原 markdown 源文稿”。

## 风险 3：会话安全问题
**应对：**
- session 与后台用户 ID 绑定
- cookie 只保存在后端
- 设置 TTL
- 提供退出登录清理能力

## 风险 4：远端接口超时/风控
**应对：**
- 增加超时、重试、友好错误提示
- 页面上允许手动重试，而不是整页报废

---

# 最终建议

这套能力应当被定义为：

> **现有 CSDN URL 导入功能的增强版，而不是替代品。**

产品上建议同时保留两条路径：
- **快捷导入**：粘贴公开 URL，快速导入
- **同步导入**：扫码登录本人 CSDN，读取文章列表后选择导入

这样即使未来：
- 公开 HTML 解析偶发失败
- 某篇文章需要登录
- 某些文章不在公开页易抓取

你仍然有更稳的一条备胎链路，不会让整个导入功能原地表演仰卧起坐。

---

# 执行顺序建议

建议按以下顺序推进，不要乱序上头：

1. Phase B1 + B2：服务骨架 + 内存会话仓储
2. Phase C1 + C2：后端 API 打通
3. Phase D1 + D2：前端类型与 API 封装
4. Phase E1 + E2：新页面静态骨架
5. Phase E3 + E4 + E5：登录、列表、预览、导入主流程
6. Phase F1 + F2：补测试
7. Phase F3：真实环境联调

---

**Plan complete and saved. Ready to execute incrementally against the current `admin-vue + backend` codebase.**
