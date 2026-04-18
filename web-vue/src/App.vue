<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { fetchArticles, fetchCategories } from '@/api/content'
import type { Article, Category } from '@/types/content'

const theme = ref<'dark' | 'light'>('dark')
const loading = ref(false)
const errorMessage = ref('')
const articles = ref<Article[]>([])
const categories = ref<Category[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = 9
const searchInput = ref('')
const searchKeyword = ref('')
const activeCategoryId = ref<number>(0)

const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize)))
const filteredCategory = computed(() => {
  if (activeCategoryId.value === 0) return '全部文章'
  return categories.value.find((item) => item.id === activeCategoryId.value)?.name ?? '分类视角'
})
const featuredArticles = computed(() => articles.value.filter((item) => item.is_top).slice(0, 3))
const regularArticles = computed(() => articles.value.filter((item) => !item.is_top))
const pageSummary = computed(() => {
  if (loading.value) return '正在从内容仓库抓取最新文章，请稍候。'
  if (errorMessage.value) return errorMessage.value
  if (total.value === 0) return '当前条件下没有匹配内容，换个关键词试试。'
  const start = (page.value - 1) * pageSize + 1
  const end = Math.min(page.value * pageSize, total.value)
  return `当前展示第 ${start} - ${end} 篇，共 ${total.value} 篇已发布文章。`
})
const heroLivePill = computed(() => loading.value ? '文章加载中…' : `已加载 ${articles.value.length} 篇`)
const heroCount = computed(() => total.value > 0 ? String(total.value).padStart(2, '0') : '--')
const themeLabel = computed(() => theme.value === 'dark' ? '暗色流光' : '柔光纸感')
const themeHint = computed(() => theme.value === 'dark' ? '点击切到明亮模式' : '点击切回暗色模式')
const themeIcon = computed(() => theme.value === 'dark' ? '🌙' : '☀️')
const topArticleCount = computed(() => articles.value.filter((item) => item.is_top).length)

function formatDate(dateString: string) {
  const date = new Date(dateString)
  if (Number.isNaN(date.getTime())) return '日期待同步'
  return new Intl.DateTimeFormat('zh-CN', {
    year: 'numeric',
    month: 'short',
    day: 'numeric'
  }).format(date)
}

function tagsOf(article: Article) {
  return article.tags
    .split(',')
    .map((item) => item.trim())
    .filter(Boolean)
    .slice(0, 3)
}

function plainSummary(article: Article) {
  const normalizedContent = article.content
    ?.replace(/[#>*`\-\n]/g, ' ')
    .replace(/\s+/g, ' ')
    .trim()

  const base = article.summary?.trim() || normalizedContent || '这篇文章还没来得及写摘要，但已经准备好让你继续深挖。'
  return base.slice(0, 120)
}

async function loadCategories() {
  const response = await fetchCategories()
  categories.value = response.data
}

async function loadArticles() {
  loading.value = true
  errorMessage.value = ''
  try {
    const response = await fetchArticles({
      page: page.value,
      page_size: pageSize,
      status: 'published',
      search: searchKeyword.value || undefined,
      category_id: activeCategoryId.value || undefined
    })
    articles.value = response.data
    total.value = response.total
    syncUrl()
  } catch (error) {
    console.error(error)
    errorMessage.value = '文章接口暂时闹小脾气了，请稍后再试。'
    articles.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

function syncUrl() {
  const url = new URL(window.location.href)
  if (searchKeyword.value) url.searchParams.set('search', searchKeyword.value)
  else url.searchParams.delete('search')
  if (activeCategoryId.value) url.searchParams.set('category_id', String(activeCategoryId.value))
  else url.searchParams.delete('category_id')
  if (page.value > 1) url.searchParams.set('page', String(page.value))
  else url.searchParams.delete('page')
  window.history.replaceState({}, '', url)
}

function hydrateFromUrl() {
  const url = new URL(window.location.href)
  const search = url.searchParams.get('search')
  const categoryId = Number(url.searchParams.get('category_id') || '0')
  const currentPage = Number(url.searchParams.get('page') || '1')

  if (search) {
    searchInput.value = search
    searchKeyword.value = search
  }
  if (!Number.isNaN(categoryId) && categoryId > 0) activeCategoryId.value = categoryId
  if (!Number.isNaN(currentPage) && currentPage > 1) page.value = currentPage
}

function submitSearch() {
  page.value = 1
  searchKeyword.value = searchInput.value.trim()
}

function switchCategory(categoryId: number) {
  if (activeCategoryId.value === categoryId) return
  activeCategoryId.value = categoryId
  page.value = 1
}

function changePage(nextPage: number) {
  if (nextPage < 1 || nextPage > totalPages.value || nextPage === page.value) return
  page.value = nextPage
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

function toggleTheme() {
  theme.value = theme.value === 'dark' ? 'light' : 'dark'
}

watch(theme, (value) => {
  document.body.dataset.theme = value
  localStorage.setItem('web-vue-theme', value)
}, { immediate: true })

watch([page, activeCategoryId], async () => {
  await loadArticles()
})

watch(searchKeyword, async () => {
  await loadArticles()
})

onMounted(async () => {
  const storedTheme = localStorage.getItem('web-vue-theme')
  if (storedTheme === 'dark' || storedTheme === 'light') theme.value = storedTheme
  hydrateFromUrl()
  await loadCategories()
  await loadArticles()
})
</script>

<template>
  <div class="aurora-bg" aria-hidden="true">
    <div class="aurora-layer one"></div>
    <div class="aurora-layer two"></div>
    <div class="aurora-layer three"></div>
    <div class="aurora-beam"></div>
  </div>

  <div class="site-shell">
    <div class="container">
      <header class="topbar">
        <div class="brand">
          <div class="brand-mark">A</div>
          <div class="brand-copy">
            <strong>children don't lie</strong>
            <span>导航 / 技术记录 / 运维 / 开发</span>
          </div>
        </div>
        <div class="topbar-meta">
          <button class="theme-toggle" type="button" aria-label="切换主题" @click="toggleTheme">
            <span class="theme-toggle-icon">{{ themeIcon }}</span>
            <span class="theme-toggle-text">
              <strong>{{ themeLabel }}</strong>
              <small>{{ themeHint }}</small>
            </span>
          </button>
          <a class="pill" href="/">首页</a>
          <a class="pill" href="/categories/">分类</a>
          <a class="pill" href="/archives/">归档</a>
          <a class="pill" href="/about/">博主</a>
          <span class="pill">{{ heroLivePill }}</span>
        </div>
      </header>

      <section class="hero">
        <div class="hero-copy">
          <div class="hero-intro">
            <div class="eyebrow">✦ Personal Tech Journal · Continuously Refined</div>
            <h1>把经验写成作品， 也把踩坑沉淀成下一次的底气。</h1>
            <p class="hero-lead">
              这里不是流水账式的技术堆放区，而是一份持续更新的个人技术刊物：记录运维现场、开发实战、Kubernetes 折腾、部署复盘，
              也认真保存那些今天解决、明天就能救命的关键细节。
            </p>
            <div class="hero-subnote">
              <span class="micro-pill">现场经验 / 可复用</span>
              <span class="micro-pill">运维 · 开发 · 基础设施</span>
              <span class="micro-pill">API 驱动 · 持续更新</span>
            </div>
          </div>
          <div class="hero-actions">
            <a class="btn btn-primary" href="#articles-section">开始阅读</a>
            <a class="btn btn-secondary" href="/archives/">查看归档馆</a>
          </div>
          <div class="hero-editor-note">
            <small>Editor's Note</small>
            <p>
              每一篇内容都尽量保留“为什么这样做、踩过什么坑、下次如何更快”的上下文。与其只给结论，不如把路径也写清楚——这样下一次回来看，才不会被过去的自己背刺。
            </p>
          </div>
        </div>
        <div class="hero-side">
          <article class="stat-card">
            <span class="stat-card-label">● Home Dispatch</span>
            <div class="stat-value">{{ heroCount }}</div>
            <p class="search-hint">{{ pageSummary }}</p>
            <div class="stat-card-note">
              <span>精选内容实时编排</span>
              <span>筛选状态与 URL 同步</span>
            </div>
          </article>
          <div class="mini-grid">
            <article class="mini-card">
              <small>当前阅读视角</small>
              <strong>{{ filteredCategory }}</strong>
            </article>
            <article class="mini-card">
              <small>阅读进度</small>
              <strong>Page {{ page }}</strong>
            </article>
          </div>
        </div>
      </section>

      <main class="main-grid">
        <section class="control-panel">
          <div class="control-head">
            <div>
              <div class="section-kicker">Reading Controls</div>
              <h2>筛选与检索</h2>
              <p>像翻阅一本有目录的技术刊物一样浏览内容：先定主题，再缩小范围，把你想找的经验尽快捞出来。</p>
            </div>
            <div class="page-stats">
              <span class="dot"></span>
              <span>{{ pageSummary }}</span>
            </div>
          </div>

          <div class="control-grid">
            <form class="search-box" @submit.prevent="submitSearch">
              <label class="search-label" for="search-input">按标题或关键词检索</label>
              <div class="search-input-wrap">
                <span>⌕</span>
                <input id="search-input" v-model="searchInput" type="text" placeholder="试试：K3s、Grafana、Oracle、Docker、部署、监控…" />
              </div>
              <div class="search-actions">
                <div class="search-hint">搜索条件会同步写入 URL，刷新、分享或稍后回来都不会丢。</div>
                <button class="btn btn-secondary search-submit" type="submit">更新检索</button>
              </div>
            </form>

            <div class="search-box">
              <div class="category-label">按主题浏览</div>
              <div class="category-filter">
                <button class="category-btn" :class="{ active: activeCategoryId === 0 }" type="button" @click="switchCategory(0)">全部内容</button>
                <button v-for="category in categories" :key="category.id" class="category-btn" :class="{ active: activeCategoryId === category.id }" type="button" @click="switchCategory(category.id)">
                  {{ category.name }}
                </button>
              </div>
              <div class="category-hint">点击分类即时切换阅读视角，并保留当前浏览上下文。</div>
            </div>
          </div>
        </section>

        <section v-if="featuredArticles.length" id="featured-section">
          <div class="section-head">
            <div class="section-copy">
              <div class="section-kicker">Featured Dispatch</div>
              <h2>精选 / 置顶编排</h2>
              <p>优先把置顶文章单独抬出来：一眼先看当前最值得阅读的内容，再决定是否继续往下刷最新列表。</p>
            </div>
            <div class="page-stats">共 {{ topArticleCount }} 篇置顶内容参与编排</div>
          </div>
          <div class="article-list featured-list">
            <article v-for="article in featuredArticles" :key="`featured-${article.id}`" class="article-card">
              <img v-if="article.cover_image" class="article-cover" :src="article.cover_image" :alt="article.title" loading="lazy" />
              <div v-else class="article-cover article-cover-fallback"></div>
              <div class="article-content">
                <div class="article-topline">
                  <span class="article-category">{{ article.category?.name || '未分类' }}</span>
                  <span class="article-id">TOP {{ article.id }}</span>
                </div>
                <h3 class="article-title">{{ article.title }}</h3>
                <p class="article-summary">{{ plainSummary(article) }}</p>
                <div class="article-tags" v-if="tagsOf(article).length">
                  <span v-for="tag in tagsOf(article)" :key="tag" class="tag-chip"># {{ tag }}</span>
                </div>
                <div class="article-footer">
                  <div class="meta-inline">
                    <span class="meta-chip">🗓 {{ formatDate(article.created_at) }}</span>
                    <span class="meta-chip">👀 {{ article.views }}</span>
                  </div>
                  <a class="article-link" :href="`/posts/${article.id}.html`">阅读全文 →</a>
                </div>
              </div>
            </article>
          </div>
        </section>

        <section id="articles-section">
          <div class="section-head">
            <div class="section-copy">
              <div class="section-kicker">Latest Dispatch</div>
              <h2>最新文章</h2>
              <p>从接口动态取回内容，保留列表来源参数与分页状态；你可以专注阅读，不用担心“点进去之后回不来”。</p>
            </div>
            <div class="page-stats">页码 {{ page }} / {{ totalPages }}</div>
          </div>

          <div v-if="loading" class="empty-state">
            <strong>正在搬运最新内容…</strong>
            <p>接口小快递正在路上，请稍等一下下。</p>
          </div>
          <div v-else-if="errorMessage" class="empty-state">
            <strong>接口短暂打盹了</strong>
            <p>{{ errorMessage }}</p>
          </div>
          <div v-else-if="regularArticles.length === 0" class="empty-state">
            <strong>当前条件下没有文章</strong>
            <p>换个关键词或切回“全部内容”，也许宝藏就在下一页。</p>
          </div>
          <div v-else class="article-list">
            <article v-for="article in regularArticles" :key="article.id" class="article-card">
              <img v-if="article.cover_image" class="article-cover" :src="article.cover_image" :alt="article.title" loading="lazy" />
              <div v-else class="article-cover article-cover-fallback"></div>
              <div class="article-content">
                <div class="article-topline">
                  <span class="article-category">{{ article.category?.name || '未分类' }}</span>
                  <span class="article-id">No. {{ article.id }}</span>
                </div>
                <h3 class="article-title">{{ article.title }}</h3>
                <p class="article-summary">{{ plainSummary(article) }}</p>
                <div class="article-tags" v-if="tagsOf(article).length">
                  <span v-for="tag in tagsOf(article)" :key="tag" class="tag-chip"># {{ tag }}</span>
                </div>
                <div class="article-footer">
                  <div class="meta-inline">
                    <span class="meta-chip">🗓 {{ formatDate(article.created_at) }}</span>
                    <span class="meta-chip">👀 {{ article.views }}</span>
                  </div>
                  <a class="article-link" :href="`/posts/${article.id}.html`">阅读全文 →</a>
                </div>
              </div>
            </article>
          </div>

          <div class="pagination-shell" v-if="totalPages > 1">
            <div class="pagination">
              <button class="page-btn" type="button" :disabled="page === 1" @click="changePage(page - 1)">‹</button>
              <button
                v-for="pageNumber in totalPages"
                :key="pageNumber"
                class="page-btn"
                :class="{ active: pageNumber === page }"
                type="button"
                @click="changePage(pageNumber)"
              >
                {{ pageNumber }}
              </button>
              <button class="page-btn" type="button" :disabled="page === totalPages" @click="changePage(page + 1)">›</button>
            </div>
          </div>
        </section>
      </main>

      <footer class="footer-note">Alexcld Home · Vue runtime edition · 从原站视觉语言平滑迁移</footer>
    </div>
  </div>
</template>
