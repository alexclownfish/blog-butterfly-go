<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { fetchArticles, fetchCategories } from '@/api/content'
import type { Article, Category } from '@/types/content'
import { articleDetailPath, formatDate, plainSummary, withFromQuery } from '@/utils/content'

const loading = ref(false)
const errorMessage = ref('')
const categories = ref<Category[]>([])
const articles = ref<Article[]>([])

const categoryStats = computed(() => categories.value.map((category) => {
  const related = articles.value.filter((article) => article.category_id === category.id)
  const views = related.reduce((sum, article) => sum + (article.views || 0), 0)

  return {
    ...category,
    articleCount: related.length,
    views,
    latest: related[0] || null
  }
}).sort((a, b) => b.articleCount - a.articleCount || b.views - a.views))

const totalArticles = computed(() => articles.value.length)
const totalViews = computed(() => articles.value.reduce((sum, article) => sum + (article.views || 0), 0))
const topCategories = computed(() => categoryStats.value.slice(0, 3))
const categorySignals = computed(() => [
  {
    label: '分类总数',
    value: `${categories.value.length} 个`,
    note: '全部来自真实分类接口，不是手搓演示数据。'
  },
  {
    label: '已加载文章',
    value: `${totalArticles.value} 篇`,
    note: '用于建立分类密度和内容分布的第一视角。'
  },
  {
    label: '阅读热度',
    value: totalViews.value.toLocaleString('zh-CN'),
    note: '按当前已加载文章浏览量聚合。'
  }
])

async function loadPage() {
  loading.value = true
  errorMessage.value = ''
  try {
    const [categoryResponse, articleResponse] = await Promise.all([
      fetchCategories(),
      fetchArticles({
        page: 1,
        page_size: 100,
        status: 'published'
      })
    ])

    categories.value = categoryResponse.data
    articles.value = articleResponse.data
    document.title = '分类 | Alexcld'
  } catch (error) {
    console.error(error)
    errorMessage.value = '分类页暂时没拿到内容编排数据，请稍后再来巡逻。'
    categories.value = []
    articles.value = []
  } finally {
    loading.value = false
  }
}

function categoryArticles(categoryId: number) {
  return articles.value
    .filter((article) => article.category_id === categoryId)
    .slice(0, 3)
}

function articleLink(articleId: number) {
  return withFromQuery(articleDetailPath(articleId), '/categories/')
}

onMounted(async () => {
  await loadPage()
})
</script>

<template>
  <main class="main-grid nav-page-shell">
    <section class="hero nav-hero">
      <div class="hero-copy">
        <div class="eyebrow">Category Dispatch · Structure Lens</div>
        <h1>把零散文章，整理成一眼能读懂的内容地图。</h1>
        <p class="hero-lead">
          分类页不是“标签云换个皮”。它应该帮人迅速看出：这个站都在写什么、哪一类最厚、从哪里开始读更顺手。
        </p>
        <div class="hero-subnote">
          <span class="micro-pill">真实分类接口</span>
          <span class="micro-pill">按内容密度排序</span>
          <span class="micro-pill">保留站内阅读跳转</span>
        </div>
      </div>
      <div class="hero-side">
        <article class="stat-card">
          <span class="stat-card-label">● Category Radar</span>
          <div class="stat-value">{{ String(categories.length).padStart(2, '0') }}</div>
          <p class="search-hint">
            {{ loading ? '正在整理分类版图…' : errorMessage || `当前识别到 ${categories.length} 个分类，覆盖 ${totalArticles} 篇文章。` }}
          </p>
          <div class="stat-card-note">
            <span>真实接口编排</span>
            <span>延续首页视觉语言</span>
          </div>
        </article>
        <div class="mini-grid">
          <article v-for="item in topCategories" :key="item.id" class="mini-card">
            <small>{{ item.name }}</small>
            <strong>{{ item.articleCount }} 篇</strong>
          </article>
        </div>
      </div>
    </section>

    <section class="signal-grid nav-signals">
      <article class="signal-panel">
        <div class="section-kicker">Category Signals</div>
        <h2>分类气压计</h2>
        <div class="metric-list">
          <div v-for="metric in categorySignals" :key="metric.label" class="metric-item">
            <small>{{ metric.label }}</small>
            <strong>{{ metric.value }}</strong>
            <p>{{ metric.note }}</p>
          </div>
        </div>
      </article>

      <article class="signal-panel signal-panel-wide">
        <div class="section-kicker">Top Lanes</div>
        <h2>优先浏览这几条内容车道</h2>
        <div class="editorial-grid">
          <article v-for="item in topCategories" :key="item.id" class="editorial-card">
            <div class="editorial-mark">✦</div>
            <h3>{{ item.name }}</h3>
            <p>
              已收录 {{ item.articleCount }} 篇文章，累计 {{ item.views.toLocaleString('zh-CN') }} 次阅读。
              {{ item.latest ? `最近更新：${formatDate(item.latest.updated_at)}` : '暂未捕获最近更新。' }}
            </p>
          </article>
        </div>
      </article>
    </section>

    <section class="detail-pagination-panel">
      <div class="section-head">
        <div class="section-copy">
          <div class="section-kicker">Category Atlas</div>
          <h2>所有分类</h2>
          <p>每个分类都带一段真实内容预览，不让页面只剩空壳按钮。</p>
        </div>
        <div class="page-stats">{{ loading ? '同步中…' : `${categoryStats.length} 个分类` }}</div>
      </div>

      <div v-if="loading" class="empty-state detail-state">
        <strong>正在整理分类地图…</strong>
        <p>内容小精灵正在给每个分类贴门牌号。</p>
      </div>
      <div v-else-if="errorMessage" class="empty-state detail-state">
        <strong>分类页暂时失联</strong>
        <p>{{ errorMessage }}</p>
      </div>
      <div v-else-if="!categoryStats.length" class="empty-state detail-state">
        <strong>暂时还没有分类数据</strong>
        <p>等内容接口回神，这里会变成完整的内容地图。</p>
      </div>
      <div v-else class="category-atlas">
        <article v-for="item in categoryStats" :key="item.id" class="category-atlas-card">
          <div class="section-head atlas-head">
            <div>
              <div class="section-kicker">Category / {{ item.id }}</div>
              <h3 class="atlas-title">{{ item.name }}</h3>
            </div>
            <div class="meta-inline atlas-meta">
              <span class="meta-chip">📚 {{ item.articleCount }} 篇</span>
              <span class="meta-chip">👀 {{ item.views.toLocaleString('zh-CN') }}</span>
            </div>
          </div>

          <div class="compact-article-list">
            <RouterLink
              v-for="article in categoryArticles(item.id)"
              :key="article.id"
              class="compact-article"
              :to="articleLink(article.id)"
            >
              <div>
                <strong>{{ article.title }}</strong>
                <p>{{ plainSummary(article) }}</p>
              </div>
              <span>{{ formatDate(article.created_at) }}</span>
            </RouterLink>
            <div v-if="!categoryArticles(item.id).length" class="compact-empty">这个分类还没抓到公开文章。</div>
          </div>
        </article>
      </div>
    </section>
  </main>
</template>
