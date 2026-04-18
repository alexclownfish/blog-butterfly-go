<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import { marked } from 'marked'
import { fetchArticleDetail, fetchArticles } from '@/api/content'
import type { Article } from '@/types/content'
import { articleDetailPath, formatDate, plainSummary, tagPath, tagsOf } from '@/utils/content'

const route = useRoute()
const loading = ref(false)
const errorMessage = ref('')
const article = ref<Article | null>(null)
const relatedArticles = ref<Article[]>([])

const articleId = computed(() => Number(route.params.id))
const fromPath = computed(() => {
  const raw = route.query.from
  return typeof raw === 'string' && raw.startsWith('/') ? raw : '/'
})
const renderedContent = computed(() => {
  if (!article.value?.content) return '<p>这篇文章正文还在路上。</p>'
  return marked.parse(article.value.content, { breaks: true })
})
const categoryName = computed(() => article.value?.category?.name || '未分类')
const summaryText = computed(() => article.value ? plainSummary(article.value) : '')
const pageHeroMeta = computed(() => article.value ? [
  `🗓 发布于 ${formatDate(article.value.created_at)}`,
  `👀 ${article.value.views} 次阅读`,
  `📁 ${categoryName.value}`
] : [])

async function loadArticle() {
  if (!Number.isFinite(articleId.value) || articleId.value <= 0) {
    errorMessage.value = '文章编号无效，无法加载详情。'
    article.value = null
    return
  }

  loading.value = true
  errorMessage.value = ''
  try {
    const detail = await fetchArticleDetail(articleId.value)
    article.value = detail
    document.title = `${detail.title} | Alexcld`
    await loadRelated(detail)
  } catch (error) {
    console.error(error)
    errorMessage.value = '文章详情接口暂时失联了，请稍后再来。'
    article.value = null
    relatedArticles.value = []
  } finally {
    loading.value = false
  }
}

async function loadRelated(current: Article) {
  try {
    const response = await fetchArticles({
      page: 1,
      page_size: 4,
      status: 'published',
      category_id: current.category_id || undefined
    })
    relatedArticles.value = response.data.filter((item) => item.id !== current.id).slice(0, 3)
  } catch (error) {
    console.error(error)
    relatedArticles.value = []
  }
}

function articleDetailTo(articleIdValue: number) {
  return `${articleDetailPath(articleIdValue)}?from=${encodeURIComponent(fromPath.value)}`
}

function tagDetailTo(tagName: string) {
  return `${tagPath(tagName)}?from=${encodeURIComponent(route.fullPath)}`
}

watch(() => route.fullPath, async () => {
  await loadArticle()
})

onMounted(async () => {
  await loadArticle()
})
</script>

<template>
  <main class="detail-shell">
    <section class="detail-hero" v-if="article">
      <div class="detail-hero-copy">
        <RouterLink class="back-link" :to="fromPath">← 返回列表</RouterLink>
        <div class="eyebrow">Article Dispatch · Detailed Reading</div>
        <h1>{{ article.title }}</h1>
        <p class="detail-summary">{{ summaryText }}</p>
        <div class="detail-meta-list">
          <span v-for="item in pageHeroMeta" :key="item" class="meta-chip">{{ item }}</span>
        </div>
        <div class="article-tags" v-if="tagsOf(article).length">
          <RouterLink v-for="tag in tagsOf(article)" :key="tag" class="tag-chip" :to="tagDetailTo(tag)"># {{ tag }}</RouterLink>
        </div>
      </div>
      <div class="detail-hero-side">
        <img v-if="article.cover_image" class="detail-cover" :src="article.cover_image" :alt="article.title" />
        <div v-else class="detail-cover detail-cover-fallback"></div>
      </div>
    </section>

    <div v-if="loading" class="empty-state detail-state">
      <strong>正文正在从内容仓库漂流过来…</strong>
      <p>请等一下下，这篇文章马上靠岸。</p>
    </div>

    <div v-else-if="errorMessage" class="empty-state detail-state">
      <strong>文章详情暂时不可读</strong>
      <p>{{ errorMessage }}</p>
    </div>

    <template v-else-if="article">
      <section class="detail-layout">
        <article class="detail-article-panel">
          <div class="detail-prose" v-html="renderedContent"></div>
        </article>

        <aside class="detail-sidebar">
          <article class="detail-side-card">
            <div class="section-kicker">Article Snapshot</div>
            <h2>阅读摘要</h2>
            <p>{{ summaryText }}</p>
            <div class="metric-list compact-metrics">
              <div class="metric-item compact">
                <small>文章编号</small>
                <strong>#{{ article.id }}</strong>
              </div>
              <div class="metric-item compact">
                <small>分类</small>
                <strong>{{ categoryName }}</strong>
              </div>
              <div class="metric-item compact">
                <small>最后更新</small>
                <strong>{{ formatDate(article.updated_at) }}</strong>
              </div>
            </div>
          </article>

          <article class="detail-side-card">
            <div class="section-kicker">Continue Reading</div>
            <h2>同类文章</h2>
            <div class="compact-article-list">
              <RouterLink v-for="item in relatedArticles" :key="item.id" class="compact-article" :to="articleDetailTo(item.id)">
                <div>
                  <strong>{{ item.title }}</strong>
                  <p>{{ plainSummary(item) }}</p>
                </div>
                <span>{{ formatDate(item.created_at) }}</span>
              </RouterLink>
              <div v-if="!relatedArticles.length" class="compact-empty">接口恢复后，这里会出现更多同类内容推荐。</div>
            </div>
          </article>
        </aside>
      </section>

      <section class="detail-pagination-panel">
        <div class="section-head">
          <div class="section-copy">
            <div class="section-kicker">Navigation</div>
            <h2>读完别迷路</h2>
            <p>保留返回列表的上下文，让你看完详情还能回到刚才的筛选位置。</p>
          </div>
          <RouterLink class="btn btn-secondary" :to="fromPath">返回列表</RouterLink>
        </div>
      </section>
    </template>
  </main>
</template>
