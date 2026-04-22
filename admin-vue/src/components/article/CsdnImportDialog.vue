<template>
  <el-dialog
    :model-value="modelValue"
    title="导入 CSDN 文章"
    width="880px"
    destroy-on-close
    @close="handleClose"
  >
    <div v-loading="previewLoading || importLoading" class="csdn-import-dialog">
      <el-alert
        type="info"
        show-icon
        :closable="false"
        class="csdn-import-dialog__tip"
        title="粘贴 CSDN 文章链接后可先预览，再导入为草稿或已发布文章。"
      />

      <el-form label-position="top" class="csdn-import-dialog__form">
        <el-form-item label="CSDN 文章链接" required>
          <div class="csdn-import-dialog__url-row">
            <el-input
              v-model="form.url"
              placeholder="https://blog.csdn.net/..."
              clearable
              @keyup.enter="handlePreview"
            />
            <el-button type="primary" :loading="previewLoading" @click="handlePreview">
              解析预览
            </el-button>
          </div>
        </el-form-item>

        <el-form-item label="导入分类" required>
          <el-select v-model="form.category_id" placeholder="请选择分类" style="width: 100%">
            <el-option
              v-for="item in categories"
              :key="item.id"
              :label="item.name"
              :value="item.id"
            />
          </el-select>
        </el-form-item>

        <el-form-item label="导入状态">
          <el-radio-group v-model="form.status">
            <el-radio-button label="draft">草稿</el-radio-button>
            <el-radio-button label="published">已发布</el-radio-button>
          </el-radio-group>
        </el-form-item>
      </el-form>

      <div v-if="preview" class="csdn-preview-card">
        <div class="csdn-preview-card__head">
          <div>
            <p class="csdn-preview-card__eyebrow">预览结果</p>
            <h3>{{ preview.title || '未识别标题' }}</h3>
          </div>
          <el-tag type="success">{{ preview.source_platform || 'csdn' }}</el-tag>
        </div>

        <p v-if="preview.summary" class="csdn-preview-card__summary">{{ preview.summary }}</p>

        <div v-if="preview.cover_image" class="csdn-preview-card__cover">
          <img :src="preview.cover_image" alt="CSDN 文章封面" />
        </div>

        <div class="csdn-preview-card__meta">
          <div>
            <span class="label">来源链接</span>
            <span class="value">{{ preview.source_url || form.url }}</span>
          </div>
          <div>
            <span class="label">标签</span>
            <span class="value">{{ preview.tags || '未识别标签' }}</span>
          </div>
        </div>

        <div class="csdn-preview-card__content">
          <span class="label">正文预览</span>
          <pre>{{ preview.content || '未识别正文内容' }}</pre>
        </div>
      </div>

      <el-empty v-else description="先解析预览，确认内容没翻车再导入～" />
    </div>

    <template #footer>
      <div class="csdn-import-dialog__footer">
        <el-button @click="handleClose">取消</el-button>
        <el-button type="primary" :disabled="!preview" :loading="importLoading" @click="handleImport">
          立即导入
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'

import { importCsdnArticleApi, previewImportCsdnApi } from '@/api/articles'
import type { Article, ArticleStatus, CsdnArticlePreview } from '@/types/article'
import type { Category } from '@/types/category'

interface Props {
  modelValue: boolean
  categories?: Category[]
}

const props = withDefaults(defineProps<Props>(), {
  categories: () => []
})

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
  (e: 'imported', article: Article): void
}>()

const previewLoading = ref(false)
const importLoading = ref(false)
const preview = ref<CsdnArticlePreview | null>(null)

const form = reactive({
  url: '',
  category_id: null as number | null,
  status: 'draft' as ArticleStatus
})

const categories = computed(() => props.categories ?? [])

watch(
  () => props.modelValue,
  (visible) => {
    if (visible) {
      if (!form.category_id && categories.value.length) {
        form.category_id = categories.value[0].id
      }
      return
    }

    resetState()
  }
)

function resetState() {
  form.url = ''
  form.category_id = categories.value[0]?.id ?? null
  form.status = 'draft'
  preview.value = null
  previewLoading.value = false
  importLoading.value = false
}

function handleClose() {
  emit('update:modelValue', false)
}

async function handlePreview() {
  const url = form.url.trim()
  if (!url) {
    ElMessage.error('请先输入 CSDN 文章链接')
    return
  }

  previewLoading.value = true
  try {
    preview.value = await previewImportCsdnApi({ url })
    ElMessage.success('预览加载成功，内容已帮你抓回来啦')
  } catch (error: any) {
    ElMessage.error(
      error?.response?.data?.error ||
        error?.response?.data?.message ||
        error?.message ||
        '解析 CSDN 文章失败'
    )
  } finally {
    previewLoading.value = false
  }
}

async function handleImport() {
  const url = form.url.trim()
  if (!url) {
    ElMessage.error('请先输入 CSDN 文章链接')
    return
  }
  if (!form.category_id) {
    ElMessage.error('请选择导入分类')
    return
  }

  importLoading.value = true
  try {
    const article = await importCsdnArticleApi({
      url,
      category_id: form.category_id,
      status: form.status
    })
    ElMessage.success('导入成功，文章已经进后台待命啦')
    emit('imported', article)
    emit('update:modelValue', false)
  } catch (error: any) {
    ElMessage.error(
      error?.response?.data?.error ||
        error?.response?.data?.message ||
        error?.message ||
        '导入 CSDN 文章失败'
    )
  } finally {
    importLoading.value = false
  }
}
</script>

<style scoped>
.csdn-import-dialog {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.csdn-import-dialog__tip {
  margin-bottom: 4px;
}

.csdn-import-dialog__form {
  display: grid;
  gap: 8px;
}

.csdn-import-dialog__url-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 12px;
  align-items: center;
}

.csdn-preview-card {
  border: 1px solid rgba(91, 143, 249, 0.18);
  background: rgba(247, 249, 255, 0.95);
  border-radius: 20px;
  padding: 20px;
  display: grid;
  gap: 16px;
}

.csdn-preview-card__head {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  align-items: flex-start;
}

.csdn-preview-card__eyebrow {
  margin: 0 0 6px;
  font-size: 12px;
  color: #7c8db5;
  text-transform: uppercase;
  letter-spacing: 0.08em;
}

.csdn-preview-card__head h3 {
  margin: 0;
  font-size: 24px;
  color: #22304a;
}

.csdn-preview-card__summary {
  margin: 0;
  color: #51627d;
  line-height: 1.7;
}

.csdn-preview-card__cover {
  overflow: hidden;
  border-radius: 16px;
  border: 1px solid rgba(148, 163, 184, 0.18);
  max-height: 260px;
}

.csdn-preview-card__cover img {
  width: 100%;
  display: block;
  object-fit: cover;
}

.csdn-preview-card__meta {
  display: grid;
  gap: 12px;
}

.csdn-preview-card__meta > div,
.csdn-preview-card__content {
  display: grid;
  gap: 6px;
}

.csdn-preview-card .label {
  font-size: 12px;
  color: #7c8db5;
}

.csdn-preview-card .value {
  color: #22304a;
  word-break: break-all;
}

.csdn-preview-card__content pre {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-word;
  max-height: 240px;
  overflow: auto;
  padding: 14px;
  border-radius: 14px;
  background: #0f172a;
  color: #e2e8f0;
  font-size: 13px;
  line-height: 1.7;
}

.csdn-import-dialog__footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

@media (max-width: 768px) {
  .csdn-import-dialog__url-row {
    grid-template-columns: 1fr;
  }

  .csdn-preview-card__head {
    flex-direction: column;
  }
}
</style>
