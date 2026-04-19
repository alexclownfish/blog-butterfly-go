<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { fetchArticles } from '@/api/content'
import type { Article } from '@/types/content'
import { articleDetailPath, formatDate, plainSummary, withFromQuery } from '@/utils/content'

const loading = ref(false)
const errorMessage = ref('')
const articles = ref<Article[]>([])

const archiveGroups = computed(() => {
  const groups = new Map<string, Article[]>()

  for (const article of articles.value) {
    const date = new Date(article.created_at)
    const key = Number.isNaN(date.getTime())
      ? '日期待同步'
      : `${date.getFullYear()}年${String(date.getMonth() + 1).padStart(2, '0')}月`

    if (!groups.has(key)) groups.set(key, [])
    groups.get(key)!.push(article)
  }

  return Array.from(groups.entries()).map(([label, items]) => ({
    label,
    items: items.sort((a, b) => +new Date(b.created_at) - +new Date(a.created_at))
  }))
})

const archiveSignals = computed(() => {
  const monthCount = archiveGroups.value.length
  const latest = articles.value[0]
  const oldest = articles.value[articles.value.length - 1]
  return [
    {
      label: '归档月份',
      value: `${monthCount} 个`,
      note: '按真实发布时间聚合归档，不靠静态占位。'
    },
    {
      label: '最新记录',
      value: latest ? formatDate(latest.created_at) : '--',
      note: '可快速确认站点最近是否持续更新。'
    },
    {
      label: '最早记录',
      value: oldest ? formatDate(oldest.created_at) : '--',
      note: '帮助判断内容积累的时间跨度。'
    }
  ]
})

async function loadPage() {
  loading.value = true
  errorMessage.value = ''
  try {
    const response = await fetchArticles({
      page: 1,
      page_size: 100,
      status: 'published'
    })
    articles.value = [...response.data].sort((a, b) => +new Date(b.created_at) - +new Date(a.created_at))
    document.title = '归档 | Alexcld'
  } catch (error) {
    console.error(error)
    errorMessage.value = '归档页还没把时间线捋顺，请稍后再来查岗。'
    articles.value = []
  } finally {
    loading.value = false
  }
}

function articleLink(articleId: number) {
  return withFromQuery(articleDetailPath(articleId), '/archives/')
}

onMounted(async () => {
  await loadPage()
})
</script>

<template>
  <main class="main-grid nav-page-shell">
    <section class="hero nav-hero">
      <div class="hero-copy">
        <div class="eyebrow">Archive Dispatch · Timeline View</div>
        <h1>把内容按时间排好队，方便你回看站点的生长轨迹。</h1>
        <p class="hero-lead">
          归档页应该像时间轴，不只是把文章再列一遍。它要让人能迅速看到：最近在写什么、多久更一次、哪几个月特别高产。
        </p>
        <div class="hero-subnote">
          <span class="micro-pill">按月份聚合</span>
          <span class="micro-pill">保留详情返回路径</span>
          <span class="micro-pill">真实发布时间排序</span>
        </div>
      </div>
      <div class="hero-side">
        <article class="stat-card">
          <span class="stat-card-label">● Timeline Radar</span>
          <div class="stat-value">{{ String(archiveGroups.length).padStart(2, '0') }}</div>
          <p class="search-hint">
            {{ loading ? '时间线正在对齐月份刻度…' : errorMessage || `当前归档覆盖 ${articles.length} 篇文章。` }}
          </p>
          <div class="stat-card-note">
            <span>月份聚合完成</span>
            <span>适合回看更新节奏</span>
          </div>
        </article>
      </div>
    </section>

    <section class="signal-grid nav-signals">
      <article class="signal-panel">
        <div class="section-kicker">Archive Signals</div>
        <h2>时间线读数</h2>
        <div class="metric-list">
          <div v-for="metric in archiveSignals" :key="metric.label" class="metric-item">
            <small>{{ metric.label }}</small>
            <strong>{{ metric.value }}</strong>
            <p>{{ metric.note }}</p>
          </div>
        </div>
      </article>

      <article class="signal-panel signal-panel-wide">
        <div class="section-kicker">Archive Reading Note</div>
        <h2>怎么用归档页更顺手</h2>
        <div class="journey-grid">
          <article class="journey-card">
            <span class="journey-step">01</span>
            <h3>先看最近月份</h3>
            <p>快速判断最近的关注重点是不是你想追的方向。</p>
          </article>
          <article class="journey-card">
            <span class="journey-step">02</span>
            <h3>再翻高产月份</h3>
            <p>高频输出期通常藏着某次集中折腾或持续迭代的主线。</p>
          </article>
          <article class="journey-card">
            <span class="journey-step">03</span>
            <h3>进入详情后再返回</h3>
            <p>保留归档上下文，不会读完一篇就迷路。</p>
          </article>
        </div>
      </article>
    </section>

    <section class="detail-pagination-panel">
      <div class="section-head">
        <div class="section-copy">
          <div class="section-kicker">Archive Timeline</div>
          <h2>归档时间线</h2>
          <p>每个月份下直接展示文章节点，省去反复跳转的认知摩擦。</p>
        </div>
        <div class="page-stats">{{ loading ? '同步中…' : `${archiveGroups.length} 个时间分组` }}</div>
      </div>

      <div v-if="loading" class="empty-state detail-state">
        <strong>归档时间线加载中…</strong>
        <p>正在给文章排队，不插队不打架。</p>
      </div>
      <div v-else-if="errorMessage" class="empty-state detail-state">
        <strong>归档页暂时打不开</strong>
        <p>{{ errorMessage }}</p>
      </div>
      <div v-else-if="!archiveGroups.length" class="empty-state detail-state">
        <strong>还没有可归档文章</strong>
        <p>等公开文章到位，这里就会长出完整时间线。</p>
      </div>
      <div v-else class="archive-timeline">
        <article v-for="group in archiveGroups" :key="group.label" class="archive-group-card">
          <div class="archive-group-head">
            <div>
              <div class="section-kicker">Archive Month</div>
              <h3 class="atlas-title">{{ group.label }}</h3>
            </div>
            <span class="meta-chip">{{ group.items.length }} 篇</span>
          </div>

          <div class="archive-list">
            <RouterLink v-for="article in group.items" :key="article.id" class="archive-item" :to="articleLink(article.id)">
              <div class="archive-dot"></div>
              <div class="archive-copy">
                <strong>{{ article.title }}</strong>
                <p>{{ plainSummary(article) }}</p>
              </div>
              <div class="archive-meta">
                <span>{{ formatDate(article.created_at) }}</span>
                <span>{{ article.category?.name || '未分类' }}</span>
              </div>
            </RouterLink>
          </div>
        </article>
      </div>
    </section>
  </main>
</template>
