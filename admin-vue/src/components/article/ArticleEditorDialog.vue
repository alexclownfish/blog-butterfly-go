<template>
  <el-dialog
    :model-value="modelValue"
    :title="dialogTitle"
    width="900px"
    destroy-on-close
    class="article-editor-dialog"
    @close="handleClose"
  >
    <div v-loading="detailLoading" class="editor-body">
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-position="top"
        class="editor-form"
      >
        <div class="editor-grid">
          <el-form-item label="标题" prop="title" class="span-2">
            <el-input
              v-model="form.title"
              placeholder="给文章取个一眼能记住的标题"
              maxlength="120"
              show-word-limit
            />
          </el-form-item>

          <el-form-item label="摘要" prop="summary" class="span-2">
            <el-input
              v-model="form.summary"
              placeholder="一句话概括这篇文章的重点"
              maxlength="200"
              show-word-limit
            />
          </el-form-item>

          <el-form-item label="分类" prop="category_id">
            <el-select
              v-model="form.category_id"
              placeholder="请选择分类"
              clearable
              style="width: 100%"
            >
              <el-option
                v-for="item in categories"
                :key="item.id"
                :label="item.name"
                :value="item.id"
              />
            </el-select>
          </el-form-item>

          <el-form-item label="标签" prop="tags">
            <el-input
              v-model="form.tags"
              placeholder="多个标签用逗号分隔，如 Docker,K8s,监控"
            />
          </el-form-item>

          <el-form-item label="封面图片 URL" prop="cover_image" class="span-2">
            <el-input
              v-model="form.cover_image"
              placeholder="先保留 URL 输入，第二阶段再接图床选择器"
            />
          </el-form-item>
        </div>

        <div class="editor-meta-row">
          <el-form-item label="文章状态" prop="status" class="meta-item">
            <el-select v-model="form.status" style="width: 180px">
              <el-option label="草稿" value="draft" />
              <el-option label="已发布" value="published" />
            </el-select>
          </el-form-item>

          <el-form-item label="置顶" prop="is_top" class="meta-item">
            <el-switch v-model="form.is_top" />
          </el-form-item>
        </div>

        <el-form-item label="正文内容" prop="content">
          <el-input
            v-model="form.content"
            type="textarea"
            :rows="16"
            resize="vertical"
            placeholder="第一阶段先用 textarea 跑通保存链路，下一阶段再升级 Markdown 编辑器。"
          />
        </el-form-item>
      </el-form>
    </div>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleClose">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSubmit">
          {{ isEditMode ? '保存修改' : '创建文章' }}
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'

import { fetchArticleDetailApi, createArticleApi, updateArticleApi } from '@/api/articles'
import type { ArticleEditorForm } from '@/types/article'
import { createDefaultArticleForm } from '@/types/article'
import type { Category } from '@/types/category'

interface Props {
  modelValue: boolean
  articleId?: number | null
  categories?: Category[]
}

const props = withDefaults(defineProps<Props>(), {
  articleId: null,
  categories: () => []
})

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
  (e: 'saved'): void
}>()

const formRef = ref<FormInstance>()
const detailLoading = ref(false)
const saving = ref(false)

const form = reactive<ArticleEditorForm>(createDefaultArticleForm())

const isEditMode = computed(() => Boolean(props.articleId))
const dialogTitle = computed(() => (isEditMode.value ? '编辑文章' : '新建文章'))
const categories = computed(() => props.categories || [])

const rules: FormRules<ArticleEditorForm> = {
  title: [{ required: true, message: '请输入文章标题', trigger: 'blur' }],
  content: [{ required: true, message: '请输入正文内容', trigger: 'blur' }],
  status: [{ required: true, message: '请选择文章状态', trigger: 'change' }]
}

function resetForm() {
  Object.assign(form, createDefaultArticleForm())
  formRef.value?.clearValidate()
}

async function loadDetail(id: number) {
  detailLoading.value = true
  try {
    const detail = await fetchArticleDetailApi(id)

    Object.assign(form, createDefaultArticleForm(), {
      title: detail.title || '',
      summary: detail.summary || '',
      content: detail.content || '',
      cover_image: detail.cover_image || '',
      category_id:
        detail.category_id === undefined || detail.category_id === null
          ? null
          : Number(detail.category_id),
      tags: detail.tags || '',
      is_top: Boolean(detail.is_top),
      status: detail.status === 'published' ? 'published' : 'draft'
    })
  } catch (error: any) {
    ElMessage.error(
      error?.response?.data?.error ||
        error?.response?.data?.message ||
        error?.message ||
        '加载文章详情失败'
    )
  } finally {
    detailLoading.value = false
  }
}

watch(
  () => props.modelValue,
  async (visible) => {
    if (!visible) return

    resetForm()

    if (props.articleId) {
      await loadDetail(props.articleId)
    }
  }
)

function normalizePayload(): ArticleEditorForm {
  return {
    title: form.title.trim(),
    summary: form.summary.trim(),
    content: form.content,
    cover_image: form.cover_image.trim(),
    category_id: form.category_id ? Number(form.category_id) : null,
    tags: form.tags.trim(),
    is_top: Boolean(form.is_top),
    status: form.status
  }
}

async function handleSubmit() {
  if (!formRef.value) return

  try {
    await formRef.value.validate()
  } catch {
    return
  }

  saving.value = true
  try {
    const payload = normalizePayload()

    if (isEditMode.value && props.articleId) {
      await updateArticleApi(props.articleId, payload)
      ElMessage.success('文章修改成功')
    } else {
      await createArticleApi(payload)
      ElMessage.success('文章创建成功')
    }

    emit('saved')
    emit('update:modelValue', false)
  } catch (error: any) {
    ElMessage.error(
      error?.response?.data?.error ||
        error?.response?.data?.message ||
        error?.message ||
        '保存文章失败'
    )
  } finally {
    saving.value = false
  }
}

function handleClose() {
  emit('update:modelValue', false)
}
</script>

<style scoped>
.editor-body {
  min-height: 160px;
}

.editor-form {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.editor-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px;
}

.span-2 {
  grid-column: span 2;
}

.editor-meta-row {
  display: flex;
  gap: 20px;
  align-items: flex-start;
  flex-wrap: wrap;
}

.meta-item {
  margin-bottom: 0;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

@media (max-width: 768px) {
  .editor-grid {
    grid-template-columns: 1fr;
  }

  .span-2 {
    grid-column: span 1;
  }
}
</style>
