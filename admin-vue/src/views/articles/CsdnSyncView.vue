<template>
  <section class="page-section csdn-sync-page">
    <div class="panel-card hero-card">
      <div class="section-head">
        <div>
          <div class="card-eyebrow">🪄 CSDN Sync</div>
          <h2>CSDN 同步导入中心</h2>
          <p>扫码登录 CSDN，拉取自己的文章列表，再按分类和状态导入到当前博客后台。</p>
        </div>

        <div class="section-head__actions">
          <el-button @click="startLogin" :disabled="startingLogin">开始扫码登录</el-button>
          <el-button @click="refreshSession" :disabled="!sessionId || refreshingSession">刷新登录状态</el-button>
        </div>
      </div>

      <el-alert
        v-if="sessionMessage"
        :title="sessionMessage"
        type="info"
        :closable="false"
        class="session-alert"
      />

      <div class="status-row">
        <span class="status-label">当前状态</span>
        <el-tag>{{ sessionStatusLabel }}</el-tag>
      </div>

      <div v-if="currentSession?.qr_code_data_url" class="qr-panel">
        <img :src="currentSession.qr_code_data_url" alt="CSDN 登录二维码" class="qr-image" />
        <div class="qr-copywriting">
          <strong>请使用 CSDN App 扫码</strong>
          <p>扫码后点击“刷新登录状态”，授权成功就能看到可导入文章列表。</p>
        </div>
      </div>
    </div>

    <div class="panel-card article-card">
      <div class="article-card__header">
        <div>
          <div class="card-eyebrow">📚 Remote Articles</div>
          <h3>可导入文章</h3>
        </div>
        <div class="article-actions">
          <el-select v-model="importForm.category_id" placeholder="选择分类" class="category-select">
            <el-option v-for="item in categories" :key="item.id" :label="item.name" :value="item.id" />
          </el-select>
          <el-radio-group v-model="importForm.status">
            <el-radio label="draft">导入为草稿</el-radio>
            <el-radio label="published">导入并发布</el-radio>
          </el-radio-group>
        </div>
      </div>

      <el-empty v-if="!articles.length" description="先扫码授权，成功后这里会出现你的 CSDN 文章列表。" />

      <div v-else class="article-list">
        <button
          v-for="article in articles"
          :key="article.id"
          type="button"
          class="article-item"
          :class="{ 'article-item--active': selectedArticleId === article.id }"
          @click="selectedArticleId = article.id"
        >
          <div class="article-item__title">{{ article.title }}</div>
          <div class="article-item__summary">{{ article.summary || '这篇文章比较高冷，暂时没有摘要。' }}</div>
          <div class="article-item__meta">{{ article.source_url || '未提供原文链接' }}</div>
        </button>
      </div>

      <div class="import-footer">
        <el-button
          type="primary"
          @click="importSelectedArticle"
          :disabled="!selectedArticleId || importing"
        >
          导入到当前博客
        </el-button>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { fetchCategoriesApi } from '@/api/categories'
import {
  fetchCsdnSyncSessionApi,
  importCsdnSyncArticleApi,
  startCsdnSyncLoginApi
} from '@/api/articles'
import type { ArticleStatus, CsdnSyncRemoteArticle, CsdnSyncSession } from '@/types/article'
import type { Category } from '@/types/category'

const categories = ref<Category[]>([])
const currentSession = ref<CsdnSyncSession | null>(null)
const selectedArticleId = ref('')
const startingLogin = ref(false)
const refreshingSession = ref(false)
const importing = ref(false)

const importForm = ref<{
  category_id: number | null
  status: ArticleStatus
}>({
  category_id: null,
  status: 'draft'
})

const articles = computed<CsdnSyncRemoteArticle[]>(() => currentSession.value?.articles || [])
const sessionId = computed(() => currentSession.value?.id || '')
const sessionMessage = computed(() => currentSession.value?.error_message || currentSession.value?.message || '')
const sessionStatusLabel = computed(() => {
  switch (currentSession.value?.status) {
    case 'authorized':
      return '已授权'
    case 'scanned':
      return '已扫码待确认'
    case 'expired':
      return '已过期'
    case 'failed':
      return '失败'
    case 'pending':
      return '待扫码'
    default:
      return '未开始'
  }
})

async function loadCategories() {
  try {
    categories.value = await fetchCategoriesApi()
  } catch (error: any) {
    ElMessage.error(
      error?.response?.data?.error ||
        error?.response?.data?.message ||
        error?.message ||
        '加载分类失败'
    )
  }
}

async function startLogin() {
  startingLogin.value = true
  try {
    currentSession.value = await startCsdnSyncLoginApi()
    selectedArticleId.value = ''
  } catch (error: any) {
    ElMessage.error(
      error?.response?.data?.error ||
        error?.response?.data?.message ||
        error?.message ||
        '创建 CSDN 登录会话失败'
    )
  } finally {
    startingLogin.value = false
  }
}

async function refreshSession() {
  if (!sessionId.value) {
    ElMessage.error('请先开始扫码登录')
    return
  }

  refreshingSession.value = true
  try {
    currentSession.value = await fetchCsdnSyncSessionApi(sessionId.value)
    if (!selectedArticleId.value && articles.value.length) {
      selectedArticleId.value = articles.value[0].id
    }
  } catch (error: any) {
    ElMessage.error(
      error?.response?.data?.error ||
        error?.response?.data?.message ||
        error?.message ||
        '刷新登录状态失败'
    )
  } finally {
    refreshingSession.value = false
  }
}

async function importSelectedArticle() {
  if (!sessionId.value || !selectedArticleId.value) {
    ElMessage.error('请先选择要导入的文章')
    return
  }
  if (!importForm.value.category_id) {
    ElMessage.error('请选择文章分类')
    return
  }

  importing.value = true
  try {
    const article = await importCsdnSyncArticleApi({
      session_id: sessionId.value,
      article_id: selectedArticleId.value,
      category_id: importForm.value.category_id,
      status: importForm.value.status
    })
    ElMessage.success(`导入成功：${article.title}`)
  } catch (error: any) {
    ElMessage.error(
      error?.response?.data?.error ||
        error?.response?.data?.message ||
        error?.message ||
        '导入文章失败'
    )
  } finally {
    importing.value = false
  }
}

onMounted(async () => {
  await loadCategories()
})
</script>

<style scoped>
.csdn-sync-page {
  display: grid;
  gap: 20px;
}

.hero-card,
.article-card {
  display: grid;
  gap: 20px;
}

.session-alert {
  margin-top: 4px;
}

.status-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.status-label {
  font-size: 13px;
  color: var(--el-text-color-secondary);
}

.qr-panel {
  display: flex;
  align-items: center;
  gap: 20px;
  padding: 20px;
  border-radius: 20px;
  background: rgba(15, 23, 42, 0.04);
}

.qr-image {
  width: 220px;
  height: 220px;
  object-fit: contain;
  border-radius: 18px;
  background: #fff;
  border: 1px solid rgba(148, 163, 184, 0.25);
}

.qr-copywriting {
  display: grid;
  gap: 8px;
}

.qr-copywriting p {
  margin: 0;
  color: var(--el-text-color-secondary);
}

.article-card__header {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  flex-wrap: wrap;
}

.article-actions {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.category-select {
  min-width: 180px;
}

.article-list {
  display: grid;
  gap: 12px;
}

.article-item {
  text-align: left;
  padding: 16px;
  border-radius: 18px;
  border: 1px solid rgba(148, 163, 184, 0.18);
  background: #fff;
  transition: all 0.2s ease;
}

.article-item:hover,
.article-item--active {
  border-color: rgba(99, 102, 241, 0.45);
  box-shadow: 0 12px 30px rgba(99, 102, 241, 0.12);
}

.article-item__title {
  font-size: 16px;
  font-weight: 700;
  color: var(--el-text-color-primary);
}

.article-item__summary,
.article-item__meta {
  margin-top: 8px;
  color: var(--el-text-color-secondary);
  font-size: 13px;
}

.import-footer {
  display: flex;
  justify-content: flex-end;
}

@media (max-width: 768px) {
  .qr-panel {
    flex-direction: column;
    align-items: flex-start;
  }

  .qr-image {
    width: 180px;
    height: 180px;
  }
}
</style>
