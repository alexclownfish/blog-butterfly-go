<template>
  <section class="page-section">
    <div class="panel-card">
      <div class="section-head">
        <div>
          <div class="card-eyebrow">📄 Content</div>
          <h2>文章管理</h2>
          <p>先跑通列表、筛选、分页和真实接口读取。</p>
        </div>

        <el-button type="primary" @click="handleCreate">新建文章</el-button>
      </div>

      <div class="filter-bar">
        <el-input
          v-model="filters.search"
          placeholder="搜索标题或正文关键词"
          clearable
          @keyup.enter="handleSearch"
        />

        <el-select v-model="filters.status" placeholder="状态" clearable>
          <el-option label="已发布" value="published" />
          <el-option label="草稿" value="draft" />
        </el-select>

        <el-select v-model="filters.category_id" placeholder="分类" clearable>
          <el-option
            v-for="item in categories"
            :key="item.id"
            :label="item.name"
            :value="item.id"
          />
        </el-select>

        <el-button @click="handleSearch">查询</el-button>
        <el-button @click="handleReset">重置</el-button>
      </div>

      <el-table
        :data="articles"
        v-loading="loading"
        class="article-table"
        empty-text="暂无文章数据"
      >
        <el-table-column prop="title" label="标题" min-width="260" />
        <el-table-column prop="category" label="分类" min-width="120">
          <template #default="{ row }">
            {{ row.category || '-' }}
          </template>
        </el-table-column>

        <el-table-column prop="status" label="状态" width="120">
          <template #default="{ row }">
            <el-tag :type="row.status === 'published' ? 'success' : 'info'">
              {{ row.status === 'published' ? '已发布' : '草稿' }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column prop="is_top" label="置顶" width="100">
          <template #default="{ row }">
            {{ row.is_top ? '是' : '否' }}
          </template>
        </el-table-column>

        <el-table-column prop="updated_at" label="更新时间" min-width="180">
          <template #default="{ row }">
            {{ row.updated_at || '-' }}
          </template>
        </el-table-column>

        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="handleEdit(row.id)">编辑</el-button>
            <el-button link type="danger" @click="handleDelete(row.id)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-bar">
        <el-pagination
          background
          layout="total, prev, pager, next"
          :total="pagination.total"
          :current-page="pagination.page"
          :page-size="pagination.page_size"
          @current-change="handlePageChange"
        />
      </div>
    </div>

    <ArticleEditorDialog
      v-model="editorVisible"
      :article-id="currentArticleId"
      :categories="categories"
      @saved="handleSaved"
    />
  </section>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { fetchArticlesApi, deleteArticleApi } from '@/api/articles'
import { fetchCategoriesApi } from '@/api/categories'
import type { Article } from '@/types/article'
import type { Category } from '@/types/category'
import ArticleEditorDialog from '@/components/article/ArticleEditorDialog.vue'

const loading = ref(false)
const articles = ref<Article[]>([])
const categories = ref<Category[]>([])
const editorVisible = ref(false)
const currentArticleId = ref<number | null>(null)

const filters = reactive({
  search: '',
  status: '',
  category_id: ''
})

const pagination = reactive({
  page: 1,
  page_size: 10,
  total: 0
})

async function loadCategories() {
  categories.value = await fetchCategoriesApi()
}

async function loadArticles() {
  loading.value = true
  try {
    const result = await fetchArticlesApi({
      page: pagination.page,
      page_size: pagination.page_size,
      search: filters.search || undefined,
      status: filters.status || undefined,
      category_id: filters.category_id || undefined
    })

    articles.value = result.list
    pagination.total = result.total
  } catch (error: any) {
    ElMessage.error(
      error?.response?.data?.error ||
        error?.response?.data?.message ||
        error?.message ||
        '加载文章列表失败'
    )
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  pagination.page = 1
  loadArticles()
}

function handleReset() {
  filters.search = ''
  filters.status = ''
  filters.category_id = ''
  pagination.page = 1
  loadArticles()
}

function handlePageChange(page: number) {
  pagination.page = page
  loadArticles()
}

function handleCreate() {
  currentArticleId.value = null
  editorVisible.value = true
}

function handleEdit(id: number) {
  currentArticleId.value = id
  editorVisible.value = true
}

async function handleSaved() {
  editorVisible.value = false
  await loadArticles()
}

async function handleDelete(id: number) {
  try {
    await ElMessageBox.confirm('删除后不可恢复，确定要删除这篇文章吗？', '删除确认', {
      type: 'warning'
    })

    await deleteArticleApi(id)
    ElMessage.success('删除成功')
    await loadArticles()
  } catch (error: any) {
    if (error === 'cancel' || error?.action === 'cancel') return

    ElMessage.error(
      error?.response?.data?.error ||
        error?.response?.data?.message ||
        error?.message ||
        '删除失败'
    )
  }
}

onMounted(async () => {
  await loadCategories()
  await loadArticles()
})
</script>
