<template>
  <section class="page-section">
    <div class="panel-card">
      <div class="section-head">
        <div>
          <div class="card-eyebrow">🗂️ Taxonomy</div>
          <h2>分类管理</h2>
          <p>维护文章分类，给创作台先把收纳盒摆整齐。</p>
        </div>

        <el-button type="primary" @click="openCreateDialog">新建分类</el-button>
      </div>

      <el-table :data="categories" v-loading="loading" empty-text="暂无分类数据">
        <el-table-column prop="name" label="分类名称" min-width="220" />
        <el-table-column prop="created_at" label="创建时间" min-width="180">
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column prop="updated_at" label="更新时间" min-width="180">
          <template #default="{ row }">
            {{ formatDate(row.updated_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openEditDialog(row)">编辑</el-button>
            <el-button link type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="480px" destroy-on-close>
      <el-form ref="formRef" :model="form" :rules="rules" label-position="top">
        <el-form-item label="分类名称" prop="name">
          <el-input
            v-model="form.name"
            maxlength="50"
            show-word-limit
            placeholder="比如：Go、Kubernetes、瞎折腾日记"
            @keyup.enter="handleSubmit"
          />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSubmit">
          {{ editingId ? '保存修改' : '创建分类' }}
        </el-button>
      </template>
    </el-dialog>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'

import {
  createCategoryApi,
  deleteCategoryApi,
  fetchCategoriesApi,
  updateCategoryApi
} from '@/api/categories'
import type { Category, CategoryPayload } from '@/types/category'

const loading = ref(false)
const saving = ref(false)
const dialogVisible = ref(false)
const editingId = ref<number | null>(null)
const categories = ref<Category[]>([])
const formRef = ref<FormInstance>()
const form = reactive<CategoryPayload>({
  name: ''
})

const dialogTitle = computed(() => (editingId.value ? '编辑分类' : '新建分类'))

const rules: FormRules<CategoryPayload> = {
  name: [{ required: true, message: '请输入分类名称', trigger: 'blur' }]
}

function resetForm() {
  form.name = ''
  formRef.value?.clearValidate()
}

function normalizePayload(): CategoryPayload {
  return {
    name: form.name.trim()
  }
}

function formatDate(value?: string) {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString('zh-CN')
}

async function loadCategories() {
  loading.value = true
  try {
    categories.value = await fetchCategoriesApi()
  } catch (error: any) {
    ElMessage.error(
      error?.response?.data?.error ||
        error?.response?.data?.message ||
        error?.message ||
        '加载分类列表失败'
    )
  } finally {
    loading.value = false
  }
}

function openCreateDialog() {
  editingId.value = null
  resetForm()
  dialogVisible.value = true
}

function openEditDialog(category: Category) {
  editingId.value = category.id
  form.name = category.name || ''
  formRef.value?.clearValidate()
  dialogVisible.value = true
}

async function handleSubmit() {
  if (!formRef.value || saving.value) return

  try {
    await formRef.value.validate()
  } catch {
    return
  }

  saving.value = true
  try {
    const payload = normalizePayload()
    if (editingId.value) {
      await updateCategoryApi(editingId.value, payload)
      ElMessage.success('分类更新成功')
    } else {
      await createCategoryApi(payload)
      ElMessage.success('分类创建成功')
    }
    dialogVisible.value = false
    await loadCategories()
  } catch (error: any) {
    ElMessage.error(
      error?.response?.data?.error ||
        error?.response?.data?.message ||
        error?.message ||
        '保存分类失败'
    )
  } finally {
    saving.value = false
  }
}

async function handleDelete(category: Category) {
  try {
    await ElMessageBox.confirm(
      `确定删除分类「${category.name}」吗？删除后可能影响已有文章归类。`,
      '删除分类',
      {
        type: 'warning',
        confirmButtonText: '删除',
        cancelButtonText: '取消'
      }
    )
  } catch {
    return
  }

  try {
    const message = await deleteCategoryApi(category.id)
    ElMessage.success(message || '删除成功')
    await loadCategories()
  } catch (error: any) {
    ElMessage.error(
      error?.response?.data?.error ||
        error?.response?.data?.message ||
        error?.message ||
        '删除分类失败'
    )
  }
}

onMounted(() => {
  void loadCategories()
})
</script>
