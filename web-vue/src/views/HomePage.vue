<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { RouterLink } from 'vue-router'
import { fetchArticles, fetchCategories } from '@/api/content'
import type { Article, Category } from '@/types/content'
import { articleDetailPath, formatDate, plainSummary, tagPath, tagsOf, withFromQuery } from '@/utils/content'

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
const latestArticles = computed(() => [...articles.value].slice(0, 4))
const currentListPath = computed(() => {
  const url = new URL(window.location.href)
  return `${url.pathname}${url.search}`
})
const insightMetrics = computed(() => {
  const totalViews = articles.value.reduce((sum, article) => sum + (article.views || 0), 0)
  const tagCount = new Set(articles.value.flatMap((article) => tagsOf(article))).size
  return [
    { label: '当前列表浏览量', value: totalViews.toLocaleString('zh-CN') || '0', note: '来自当前分页已加载文章' },
    { label: '当前分类视角', value: filteredCategory.value, note: '筛选状态与 URL 同步' },
    { label: '标签丰富度', value: `${tagCount || 0} 个`, note: '便于快速建立阅读路径' }
  ]
})
const editorialHighlights = computed(() => [
  {
    title: '值班速记与事故复盘',
    description: '把线上问题拆成可复现、可追溯、可交接的记录，不让经验只活在脑子里。'
  },
  {
    title: '部署与监控流水线',
    description: '围绕 K3s、Grafana、Docker、自动化脚本构建更顺手的基础设施肌肉记忆。'
  },
  {
    title: '开发中的细节审美',
    description: '不只把功能做出来，也重视使用过程是否顺滑、界面是否有呼吸感。'
  }
])
const readingFlow = computed(() => [
  { step: '01', title: '先看置顶', detail: '优先读当前最值得看的内容，快速进入站点主题。' },
  { step: '02', title: '再用筛选', detail: '按关键词、分类缩小范围，少翻无关页面。' },
  { step: '03', title: '进入详情页', detail: '保留返回路径和上下文，读完还能回到刚才的位置。' }
])
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
const topArticleCount = computed(() => articles.value.filter((item) => item.is_top).length)

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
}

function articleDetailTo(articleId: number) {
  return withFromQuery(articleDetailPath(articleId), currentListPath.value)
}

function tagDetailTo(tagName: string) {
  return withFromQuery(tagPath(tagName), currentListPath.value)
}

watch([page, activeCategoryId], async () => {
  await loadArticles()
})

watch(searchKeyword, async () => {
  await loadArticles()
})

onMounted(async () => {
  hydrateFromUrl()
  await loadCategories()
  await loadArticles()
})
</script>

<template>
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
        <a class="btn btn-secondary" href="#signals-section">看第二屏</a>
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
      <div class="mini-card hero-live-card">
        <small>实时状态</small>
        <strong>{{ heroLivePill }}</strong>
        <p>已识别 {{ topArticleCount }} 篇置顶文章参与首页编排。</p>
      </div>
    </div>
  </section>

  <main class="main-grid">
    <section id="signals-section" class="signal-grid">
      <article class="signal-panel signal-panel-wide">
        <div class="section-kicker">Second Screen · Editorial Signals</div>
        <h2>第二屏：把站点内容组织成更清晰的阅读信号</h2>
        <p>
          首页不只是“文章列表”。第二屏应该帮助访客快速判断：这里到底写什么、适合从哪读起、哪些内容值得持续追踪。
        </p>
        <div class="editorial-grid">
          <article v-for="item in editorialHighlights" :key="item.title" class="editorial-card">
            <div class="editorial-mark">✦</div>
            <h3>{{ item.title }}</h3>
            <p>{{ item.description }}</p>
          </article>
        </div>
      </article>

      <article class="signal-panel">
        <div class="section-kicker">Insight Metrics</div>
        <h2>站点气压计</h2>
        <div class="metric-list">
          <div v-for="metric in insightMetrics" :key="metric.label" class="metric-item">
            <small>{{ metric.label }}</small>
            <strong>{{ metric.value }}</strong>
            <p>{{ metric.note }}</p>
          </div>
        </div>
      </article>

      <article class="signal-panel signal-panel-wide">
        <div class="section-kicker">Reading Flow</div>
        <h2>给新访客的阅读路径</h2>
        <div class="journey-grid">
          <article v-for="item in readingFlow" :key="item.step" class="journey-card">
            <span class="journey-step">{{ item.step }}</span>
            <h3>{{ item.title }}</h3>
            <p>{{ item.detail }}</p>
          </article>
        </div>
      </article>

      <article class="signal-panel">
        <div class="section-kicker">Fresh Queue</div>
        <h2>最近上架</h2>
        <div class="compact-article-list">
          <RouterLink v-for="article in latestArticles" :key="article.id" class="compact-article" :to="articleDetailTo(article.id)">
            <div>
              <strong>{{ article.title }}</strong>
              <p>{{ plainSummary(article) }}</p>
            </div>
            <span>{{ formatDate(article.created_at) }}</span>
          </RouterLink>
          <div v-if="!latestArticles.length" class="compact-empty">等待内容接口恢复后，这里会显示最新文章队列。</div>
        </div>
      </article>
    </section>

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
              <RouterLink v-for="tag in tagsOf(article)" :key="tag" class="tag-chip" :to="tagDetailTo(tag)"># {{ tag }}</RouterLink>
            </div>
            <div class="article-footer">
              <div class="meta-inline">
                <span class="meta-chip">🗓 {{ formatDate(article.created_at) }}</span>
                <span class="meta-chip">👀 {{ article.views }}</span>
              </div>
              <RouterLink class="article-link" :to="articleDetailTo(article.id)">阅读全文 →</RouterLink>
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
              <RouterLink v-for="tag in tagsOf(article)" :key="tag" class="tag-chip" :to="tagDetailTo(tag)"># {{ tag }}</RouterLink>
            </div>
            <div class="article-footer">
              <div class="meta-inline">
                <span class="meta-chip">🗓 {{ formatDate(article.created_at) }}</span>
                <span class="meta-chip">👀 {{ article.views }}</span>
              </div>
              <RouterLink class="article-link" :to="articleDetailTo(article.id)">阅读全文 →</RouterLink>
            </div>
          </div>
        </article>
      </div>

      <div class="pagination-shell" v-if="totalPages > 1">
        <div class="pagination">
          <button class="page-btn" type="button" :disabled="page === 1" @click="changePage(page - 1)">‹</button>
          <button v-for="pageNumber in totalPages" :key="pageNumber" class="page-btn" :class="{ active: pageNumber === page }" type="button" @click="changePage(pageNumber)">
            {{ pageNumber }}
          </button>
          <button class="page-btn" type="button" :disabled="page === totalPages" @click="changePage(page + 1)">›</button>
        </div>
      </div>
    </section>
  </main>
</template>
