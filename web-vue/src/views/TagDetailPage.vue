<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import { fetchArticles, fetchTags } from '@/api/content'
import type { Article, Tag } from '@/types/content'
import { articleDetailPath, formatDate, plainSummary, tagPath, tagsOf } from '@/utils/content'

const route = useRoute()
const loading = ref(false)
const errorMessage = ref('')
const articles = ref<Article[]>([])
const tags = ref<Tag[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = 9

const tagName = computed(() => {
  const raw = route.params.name
  return typeof raw === 'string' ? decodeURIComponent(raw) : ''
})
const fromPath = computed(() => {
  const raw = route.query.from
  return typeof raw === 'string' && raw.startsWith('/') ? raw : '/'
})
const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize)))
const pageSummary = computed(() => {
  if (loading.value) return `正在整理 #${tagName.value} 标签下的文章…`
  if (errorMessage.value) return errorMessage.value
  if (total.value === 0) return `#${tagName.value} 下面暂时还没有公开文章。`
  const start = (page.value - 1) * pageSize + 1
  const end = Math.min(page.value * pageSize, total.value)
  return `#${tagName.value} 当前展示第 ${start} - ${end} 篇，共 ${total.value} 篇相关文章。`
})
const relatedTags = computed(() => tags.value.filter((tag) => tag.name !== tagName.value).slice(0, 12))

async function loadTags() {
  try {
    const response = await fetchTags()
    tags.value = response.data
  } catch (error) {
    console.error(error)
    tags.value = []
  }
}

async function loadArticles() {
  if (!tagName.value) {
    errorMessage.value = '标签名称为空，无法建立阅读视角。'
    articles.value = []
    total.value = 0
    return
  }

  loading.value = true
  errorMessage.value = ''
  try {
    const response = await fetchArticles({
      page: page.value,
      page_size: pageSize,
      status: 'published',
      tag: tagName.value
    })
    articles.value = response.data
    total.value = response.total
    document.title = `#${tagName.value} | Alexcld`
  } catch (error) {
    console.error(error)
    errorMessage.value = '标签聚合接口暂时失联了，请稍后再试。'
    articles.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

function articleDetailTo(articleId: number) {
  return `${articleDetailPath(articleId)}?from=${encodeURIComponent(route.fullPath)}`
}

function tagDetailTo(name: string) {
  return `${tagPath(name)}?from=${encodeURIComponent(fromPath.value)}`
}

function changePage(nextPage: number) {
  if (nextPage < 1 || nextPage > totalPages.value || nextPage === page.value) return
  page.value = nextPage
}

watch(() => route.params.name, async () => {
  page.value = 1
  await loadArticles()
})

watch(page, async () => {
  await loadArticles()
})

onMounted(async () => {
  await loadTags()
  await loadArticles()
})
</script>

<template>
  <main class="detail-shell">
    <section class="detail-hero tag-hero">
      <div class="detail-hero-copy">
        <RouterLink class="back-link" :to="fromPath">← 返回上一视角</RouterLink>
        <div class="eyebrow">Tag Dispatch · Topic Lens</div>
        <h1># {{ tagName || '未命名标签' }}</h1>
        <p class="detail-summary">
          把零散文章按标签重新编排成一条专题阅读路径：先看这个主题下有什么，再决定下一篇往哪钻。
        </p>
        <div class="detail-meta-list">
          <span class="meta-chip">🏷 共 {{ total }} 篇相关文章</span>
          <span class="meta-chip">📚 支持分页聚合</span>
          <span class="meta-chip">🔁 返回上下文已保留</span>
        </div>
      </div>
      <div class="detail-hero-side">
        <article class="detail-side-card hero-side-card">
          <div class="section-kicker">Tag Snapshot</div>
          <h2>标签侧写</h2>
          <p>{{ pageSummary }}</p>
          <div class="tag-spotlight-list compact">
            <RouterLink v-for="tag in relatedTags" :key="tag.id" class="tag-spotlight-chip" :to="tagDetailTo(tag.name)">
              <span># {{ tag.name }}</span>
            </RouterLink>
            <div v-if="!relatedTags.length" class="compact-empty">等标签接口回神后，这里会出现更多可跳转主题。</div>
          </div>
        </article>
      </div>
    </section>

    <section class="detail-pagination-panel">
      <div class="section-head">
        <div class="section-copy">
          <div class="section-kicker">Topic Feed</div>
          <h2>标签下的文章列表</h2>
          <p>这不是纯展示型标签云，而是可继续深入阅读的内容入口。</p>
        </div>
        <div class="page-stats">页码 {{ page }} / {{ totalPages }}</div>
      </div>

      <div v-if="loading" class="empty-state detail-state">
        <strong>正在编排标签文章…</strong>
        <p>{{ pageSummary }}</p>
      </div>
      <div v-else-if="errorMessage" class="empty-state detail-state">
        <strong>标签视角暂时打不开</strong>
        <p>{{ errorMessage }}</p>
      </div>
      <div v-else-if="articles.length === 0" class="empty-state detail-state">
        <strong>这个标签还没攒出公开内容</strong>
        <p>{{ pageSummary }}</p>
      </div>
      <div v-else class="article-list tag-detail-list">
        <article v-for="article in articles" :key="article.id" class="article-card">
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