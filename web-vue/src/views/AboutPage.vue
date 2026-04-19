<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { fetchArticles, fetchCategories, fetchTags } from '@/api/content'
import { tagPath, withFromQuery } from '@/utils/content'

const loading = ref(false)
const errorMessage = ref('')
const articleCount = ref(0)
const categoryCount = ref(0)
const tagCount = ref(0)

const profileSignals = computed(() => [
  {
    label: '公开文章',
    value: `${articleCount.value} 篇`,
    note: '来自真实文章接口，用来侧写站点当前内容体量。'
  },
  {
    label: '分类版图',
    value: `${categoryCount.value} 个`,
    note: '反映内容组织方式，而不是单纯的数量展示。'
  },
  {
    label: '标签密度',
    value: `${tagCount.value} 个`,
    note: '说明内容是否形成了可复用、可串联的主题网络。'
  }
])

const focusAreas = [
  {
    title: '运维与基础设施',
    description: '围绕 K3s、Docker、Grafana、Prometheus、部署脚本，把线上经验磨成可回放的操作手感。'
  },
  {
    title: '开发与产品细节',
    description: '不仅追求功能可用，也关注后台交互、创作效率、界面节奏和使用中的细小阻力。'
  },
  {
    title: '问题复盘与可复用经验',
    description: '把踩坑过程写清楚，把“为什么这样修”留下来，让未来的自己少被过去背刺。'
  }
]

const shortcuts = computed(() => [
  { label: '看全部文章', to: '/' },
  { label: '看看分类地图', to: '/categories/' },
  { label: '翻时间线归档', to: '/archives/' },
  { label: '追一个标签主题', to: withFromQuery(tagPath('运维'), '/about/') }
])

async function loadPage() {
  loading.value = true
  errorMessage.value = ''
  try {
    const [articleResponse, categoryResponse, tagResponse] = await Promise.all([
      fetchArticles({ page: 1, page_size: 100, status: 'published' }),
      fetchCategories(),
      fetchTags()
    ])

    articleCount.value = articleResponse.total || articleResponse.data.length
    categoryCount.value = categoryResponse.data.length
    tagCount.value = tagResponse.data.length
    document.title = '关于 | Alexcld'
  } catch (error) {
    console.error(error)
    errorMessage.value = '关于页暂时没拿到站点画像数据，不过人设没有跑路。'
    articleCount.value = 0
    categoryCount.value = 0
    tagCount.value = 0
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  await loadPage()
})
</script>

<template>
  <main class="main-grid nav-page-shell">
    <section class="detail-hero about-hero">
      <div class="detail-hero-copy">
        <div class="eyebrow">About Dispatch · Personal Note</div>
        <h1>一个把开发、运维和审美洁癖一起打包写进博客的人。</h1>
        <p class="detail-summary">
          这里记录的不只是“做了什么”，更关心“为什么这样做、哪里踩坑了、下次怎么更快”。
          如果说首页是内容入口，那关于页更像是这座小站的使用说明书和人格注脚。
        </p>
        <div class="detail-meta-list">
          <span class="meta-chip">🛠 运维 / 开发 / 基础设施</span>
          <span class="meta-chip">✍️ 喜欢把经验写清楚</span>
          <span class="meta-chip">🎛 在意产品细节与使用手感</span>
        </div>
      </div>
      <div class="detail-hero-side">
        <article class="detail-side-card hero-side-card">
          <div class="section-kicker">Site Portrait</div>
          <h2>站点画像</h2>
          <p>{{ loading ? '正在同步站点画像数据…' : errorMessage || '以下数字来自站点当前真实接口返回。' }}</p>
          <div class="metric-list compact-metrics">
            <div v-for="metric in profileSignals" :key="metric.label" class="metric-item compact">
              <small>{{ metric.label }}</small>
              <strong>{{ metric.value }}</strong>
            </div>
          </div>
        </article>
      </div>
    </section>

    <section class="signal-grid nav-signals">
      <article class="signal-panel signal-panel-wide">
        <div class="section-kicker">Focus Areas</div>
        <h2>这座站主要在折腾什么</h2>
        <div class="editorial-grid">
          <article v-for="item in focusAreas" :key="item.title" class="editorial-card">
            <div class="editorial-mark">✦</div>
            <h3>{{ item.title }}</h3>
            <p>{{ item.description }}</p>
          </article>
        </div>
      </article>

      <article class="signal-panel">
        <div class="section-kicker">Quick Routes</div>
        <h2>从这儿继续逛</h2>
        <div class="quick-route-list">
          <RouterLink v-for="item in shortcuts" :key="item.label" class="compact-article quick-route-card" :to="item.to">
            <div>
              <strong>{{ item.label }}</strong>
              <p>站内无刷跳转，继续沿着同一套视觉语言阅读。</p>
            </div>
            <span>→</span>
          </RouterLink>
        </div>
      </article>
    </section>

    <section class="detail-pagination-panel">
      <div class="section-head">
        <div class="section-copy">
          <div class="section-kicker">About This Blog</div>
          <h2>写这座站时在意的几件事</h2>
          <p>不是把文章丢上来就完事，而是希望整个阅读与创作流程都更顺滑、更有秩序感。</p>
        </div>
      </div>

      <div class="journey-grid about-values-grid">
        <article class="journey-card">
          <span class="journey-step">01</span>
          <h3>内容要能复用</h3>
          <p>不只记结果，也保留环境、路径、坑位和关键判断，让文章能在未来继续救场。</p>
        </article>
        <article class="journey-card">
          <span class="journey-step">02</span>
          <h3>页面要有呼吸感</h3>
          <p>即使是技术博客，也值得拥有舒服的层次、柔和的节奏和不冒犯人的信息密度。</p>
        </article>
        <article class="journey-card">
          <span class="journey-step">03</span>
          <h3>后台要服务创作</h3>
          <p>重视编辑效率、自动保存、图片插入、Markdown 预览这些真正影响写作心情的小地方。</p>
        </article>
      </div>
    </section>
  </main>
</template>
