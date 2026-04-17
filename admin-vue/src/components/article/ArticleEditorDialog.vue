<template>
  <el-dialog
    :model-value="modelValue"
    :title="dialogTitle"
    width="1080px"
    destroy-on-close
    class="article-editor-dialog"
    @open="handleOpen"
    @close="handleClose"
  >
    <div ref="dialogBodyRef" v-loading="detailLoading" class="editor-body">
      <el-alert
        v-if="draftRecoveryAvailable"
        type="warning"
        show-icon
        :closable="false"
        class="draft-alert"
      >
        <template #title>
          检测到{{ isEditMode ? '本地草稿' : '未完成的新文章草稿' }}
        </template>
        <div class="draft-alert__content">
          <span>{{ draftRecoveryMessage }}</span>
          <div class="draft-alert__actions">
            <el-button size="small" @click="discardLocalDraft">忽略本地草稿</el-button>
            <el-button size="small" type="warning" @click="restoreLocalDraft">恢复本地草稿</el-button>
          </div>
        </div>
      </el-alert>

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

          <div class="editor-status-panel">
            <el-tag size="small" :type="serverSaveTagType">{{ serverSaveLabel }}</el-tag>
            <el-tag size="small" :type="localDraftTagType">{{ localDraftLabel }}</el-tag>
            <span class="shortcut-hint">⌘/Ctrl + S 保存到服务器</span>
          </div>
        </div>

        <el-form-item label="正文内容" prop="content" class="content-form-item">
          <div class="markdown-editor-shell">
            <div class="markdown-toolbar">
              <div class="markdown-toolbar__left">
                <el-button-group>
                  <el-button size="small" @click="insertMarkdownSyntax('**', '**', '粗体文本')">
                    粗体
                  </el-button>
                  <el-button size="small" @click="insertMarkdownSyntax('*', '*', '斜体文本')">
                    斜体
                  </el-button>
                  <el-button size="small" @click="insertMarkdownSyntax('## ', '', '二级标题')">
                    标题
                  </el-button>
                  <el-button
                    size="small"
                    @click="insertMarkdownSyntax('- ', '', '列表项', { multilinePrefix: '- ' })"
                  >
                    列表
                  </el-button>
                  <el-button
                    size="small"
                    @click="insertMarkdownSyntax('> ', '', '引用内容', { multilinePrefix: '> ' })"
                  >
                    引用
                  </el-button>
                  <el-button
                    size="small"
                    @click="insertMarkdownSyntax('`', '`', 'inline-code')"
                  >
                    行内代码
                  </el-button>
                  <el-button
                    size="small"
                    @click="insertMarkdownSyntax('```\n', '\n```', 'code block')"
                  >
                    代码块
                  </el-button>
                  <el-button
                    size="small"
                    @click="insertMarkdownSyntax('[', '](https://example.com)', '链接文字')"
                  >
                    链接
                  </el-button>
                </el-button-group>
              </div>
              <div class="markdown-toolbar__right">
                <el-radio-group v-model="previewMode" size="small">
                  <el-radio-button label="edit">编辑</el-radio-button>
                  <el-radio-button label="split">分栏预览</el-radio-button>
                  <el-radio-button label="preview">仅预览</el-radio-button>
                </el-radio-group>
              </div>
            </div>

            <div class="markdown-tips">
              支持 Markdown 编写；本地草稿会自动保存，服务器保存仍需点击按钮或按 ⌘/Ctrl + S。
            </div>

            <div class="markdown-workspace" :class="`mode-${previewMode}`">
              <div v-show="previewMode !== 'preview'" class="markdown-pane markdown-pane--editor">
                <textarea
                  ref="contentTextareaRef"
                  v-model="form.content"
                  class="markdown-textarea"
                  placeholder="# 从这里开始写作\n\n- 支持 Markdown 语法\n- 可使用分栏实时预览\n- 本地草稿会自动保存"
                />
              </div>

              <div v-show="previewMode !== 'edit'" class="markdown-pane markdown-pane--preview">
                <div v-if="form.content.trim()" class="markdown-preview" v-html="renderedMarkdown"></div>
                <el-empty v-else description="开始输入 Markdown 后，这里会显示预览" />
              </div>
            </div>
          </div>
        </el-form-item>
      </el-form>
    </div>

    <template #footer>
      <div class="dialog-footer">
        <el-button v-if="hasLocalDraft" @click="discardLocalDraft">清除本地草稿</el-button>
        <el-button @click="handleClose">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSubmit">
          {{ isEditMode ? '保存修改' : '创建文章' }}
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, reactive, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { marked } from 'marked'

import { fetchArticleDetailApi, createArticleApi, updateArticleApi } from '@/api/articles'
import type { ArticleEditorForm } from '@/types/article'
import { createDefaultArticleForm } from '@/types/article'
import type { Category } from '@/types/category'

interface Props {
  modelValue: boolean
  articleId?: number | null
  categories?: Category[]
}

interface LocalDraftSnapshot extends ArticleEditorForm {
  updated_at: string
  article_id: number | null
}

type PreviewMode = 'edit' | 'split' | 'preview'
type LocalDraftState = 'clean' | 'dirty' | 'saving' | 'saved' | 'error' | 'restored'
type ServerSaveState = 'idle' | 'saving' | 'saved' | 'error'

const AUTOSAVE_DELAY = 1200
const NEW_ARTICLE_DRAFT_KEY = 'admin-vue:article-editor:new'

const props = withDefaults(defineProps<Props>(), {
  articleId: null,
  categories: () => []
})

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
  (e: 'saved'): void
}>()

const formRef = ref<FormInstance>()
const dialogBodyRef = ref<HTMLElement>()
const contentTextareaRef = ref<HTMLTextAreaElement>()
const detailLoading = ref(false)
const saving = ref(false)
const previewMode = ref<PreviewMode>('split')
const serverSaveState = ref<ServerSaveState>('idle')
const localDraftState = ref<LocalDraftState>('clean')
const localDraftTimestamp = ref('')
const draftRecoverySnapshot = ref<LocalDraftSnapshot | null>(null)
const serverUpdatedAt = ref('')
const autosaveTimer = ref<number | null>(null)
const suppressAutosave = ref(false)
const dialogOpened = ref(false)
const currentDraftExists = ref(false)

const form = reactive<ArticleEditorForm>(createDefaultArticleForm())

const isEditMode = computed(() => Boolean(props.articleId))
const dialogTitle = computed(() => (isEditMode.value ? '编辑文章' : '新建文章'))
const categories = computed(() => props.categories || [])
const hasLocalDraft = computed(() => currentDraftExists.value)
const draftRecoveryAvailable = computed(() => Boolean(draftRecoverySnapshot.value))

const draftRecoveryMessage = computed(() => {
  if (!draftRecoverySnapshot.value) return ''

  const localLabel = formatDateTime(draftRecoverySnapshot.value.updated_at)
  if (!isEditMode.value || !serverUpdatedAt.value) {
    return `本地保存时间：${localLabel}`
  }

  return `本地保存时间：${localLabel}；服务器最近更新时间：${formatDateTime(serverUpdatedAt.value)}`
})

const renderedMarkdown = computed(() => marked.parse(form.content || '', { breaks: true }) as string)

const serverSaveTagType = computed(() => {
  switch (serverSaveState.value) {
    case 'saved':
      return 'success'
    case 'error':
      return 'danger'
    case 'saving':
      return 'warning'
    default:
      return 'info'
  }
})

const serverSaveLabel = computed(() => {
  switch (serverSaveState.value) {
    case 'saving':
      return '正在保存到服务器'
    case 'saved':
      return '已保存到服务器'
    case 'error':
      return '服务器保存失败'
    default:
      return '尚未保存到服务器'
  }
})

const localDraftTagType = computed(() => {
  switch (localDraftState.value) {
    case 'saved':
    case 'restored':
      return 'success'
    case 'saving':
    case 'dirty':
      return 'warning'
    case 'error':
      return 'danger'
    default:
      return 'info'
  }
})

const localDraftLabel = computed(() => {
  switch (localDraftState.value) {
    case 'saving':
      return '正在自动保存本地草稿'
    case 'dirty':
      return '本地草稿待保存'
    case 'saved':
      return localDraftTimestamp.value
        ? `本地草稿已保存 ${formatDateTime(localDraftTimestamp.value)}`
        : '本地草稿已保存'
    case 'restored':
      return localDraftTimestamp.value
        ? `已恢复本地草稿 ${formatDateTime(localDraftTimestamp.value)}`
        : '已恢复本地草稿'
    case 'error':
      return '本地草稿保存失败'
    default:
      return '尚无本地草稿'
  }
})

const rules: FormRules<ArticleEditorForm> = {
  title: [{ required: true, message: '请输入文章标题', trigger: 'blur' }],
  content: [{ required: true, message: '请输入正文内容', trigger: 'blur' }],
  status: [{ required: true, message: '请选择文章状态', trigger: 'change' }]
}

function resetForm() {
  suppressAutosave.value = true
  clearAutosaveTimer()
  Object.assign(form, createDefaultArticleForm())
  formRef.value?.clearValidate()
  previewMode.value = 'split'
  serverSaveState.value = 'idle'
  localDraftState.value = 'clean'
  localDraftTimestamp.value = ''
  draftRecoverySnapshot.value = null
  serverUpdatedAt.value = ''
  suppressAutosave.value = false
}

function getDraftStorageKey(articleId = props.articleId ?? null) {
  return articleId ? `admin-vue:article-editor:${articleId}` : NEW_ARTICLE_DRAFT_KEY
}

function buildSnapshot(): LocalDraftSnapshot {
  return {
    ...normalizePayload(),
    updated_at: new Date().toISOString(),
    article_id: props.articleId ?? null
  }
}

function loadLocalDraft(articleId = props.articleId ?? null): LocalDraftSnapshot | null {
  const raw = window.localStorage.getItem(getDraftStorageKey(articleId))
  if (!raw) return null

  try {
    const parsed = JSON.parse(raw) as LocalDraftSnapshot
    return parsed && typeof parsed === 'object' ? parsed : null
  } catch {
    return null
  }
}

function saveLocalDraft() {
  try {
    const snapshot = buildSnapshot()
    window.localStorage.setItem(getDraftStorageKey(), JSON.stringify(snapshot))
    localDraftTimestamp.value = snapshot.updated_at
    localDraftState.value = 'saved'
    currentDraftExists.value = true
    updateDraftRecoverySnapshot()
  } catch {
    localDraftState.value = 'error'
  }
}

function clearLocalDraft(articleId = props.articleId ?? null) {
  window.localStorage.removeItem(getDraftStorageKey(articleId))
  if ((draftRecoverySnapshot.value?.article_id ?? null) === articleId) {
    draftRecoverySnapshot.value = null
  }
  if (!loadLocalDraft(articleId)) {
    currentDraftExists.value = false
    localDraftTimestamp.value = ''
    localDraftState.value = 'clean'
  }
}

function clearAutosaveTimer() {
  if (autosaveTimer.value !== null) {
    window.clearTimeout(autosaveTimer.value)
    autosaveTimer.value = null
  }
}

function queueAutosave() {
  if (!dialogOpened.value || suppressAutosave.value || detailLoading.value) return

  localDraftState.value = 'saving'
  clearAutosaveTimer()
  autosaveTimer.value = window.setTimeout(() => {
    saveLocalDraft()
    autosaveTimer.value = null
  }, AUTOSAVE_DELAY)
}

function updateDraftRecoverySnapshot() {
  const snapshot = loadLocalDraft()
  currentDraftExists.value = Boolean(snapshot)
  if (!snapshot) {
    draftRecoverySnapshot.value = null
    return
  }

  draftRecoverySnapshot.value = snapshot
}

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

function formatDateTime(value: string) {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value

  return date.toLocaleString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

async function loadDetail(id: number) {
  detailLoading.value = true
  try {
    const detail = await fetchArticleDetailApi(id)
    serverUpdatedAt.value = detail.updated_at || ''

    suppressAutosave.value = true
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
    suppressAutosave.value = false
    updateDraftRecoverySnapshot()
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

async function initializeDialog() {
  resetForm()

  if (props.articleId) {
    await loadDetail(props.articleId)
  } else {
    updateDraftRecoverySnapshot()
  }

  await nextTick()
  contentTextareaRef.value?.focus()
}

function applyDraft(snapshot: LocalDraftSnapshot) {
  suppressAutosave.value = true
  Object.assign(form, createDefaultArticleForm(), {
    title: snapshot.title || '',
    summary: snapshot.summary || '',
    content: snapshot.content || '',
    cover_image: snapshot.cover_image || '',
    category_id:
      snapshot.category_id === undefined || snapshot.category_id === null
        ? null
        : Number(snapshot.category_id),
    tags: snapshot.tags || '',
    is_top: Boolean(snapshot.is_top),
    status: snapshot.status === 'published' ? 'published' : 'draft'
  })
  suppressAutosave.value = false
  localDraftTimestamp.value = snapshot.updated_at || ''
  localDraftState.value = 'restored'
  draftRecoverySnapshot.value = null
}

function restoreLocalDraft() {
  const snapshot = draftRecoverySnapshot.value || loadLocalDraft()
  if (!snapshot) {
    ElMessage.info('没有可恢复的本地草稿')
    return
  }

  applyDraft(snapshot)
  ElMessage.success('已恢复本地草稿')
}

function discardLocalDraft() {
  clearAutosaveTimer()
  clearLocalDraft()
  draftRecoverySnapshot.value = null
  ElMessage.success('本地草稿已清除')
}

async function handleSubmit() {
  if (!formRef.value || saving.value) return

  try {
    await formRef.value.validate()
  } catch {
    return
  }

  saving.value = true
  serverSaveState.value = 'saving'

  try {
    const payload = normalizePayload()

    if (isEditMode.value && props.articleId) {
      await updateArticleApi(props.articleId, payload)
      ElMessage.success('文章修改成功')
    } else {
      await createArticleApi(payload)
      ElMessage.success('文章创建成功')
    }

    serverSaveState.value = 'saved'
    clearAutosaveTimer()
    clearLocalDraft()
    emit('saved')
    emit('update:modelValue', false)
  } catch (error: any) {
    serverSaveState.value = 'error'
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

function insertMarkdownSyntax(
  prefix: string,
  suffix: string,
  placeholder: string,
  options?: { multilinePrefix?: string }
) {
  const textarea = contentTextareaRef.value
  if (!textarea) return

  const start = textarea.selectionStart
  const end = textarea.selectionEnd
  const selectedText = form.content.slice(start, end)
  const content = selectedText || placeholder
  const nextContent = options?.multilinePrefix
    ? content
        .split('\n')
        .map((line) => `${options.multilinePrefix}${line}`)
        .join('\n')
    : content

  form.content = `${form.content.slice(0, start)}${prefix}${nextContent}${suffix}${form.content.slice(end)}`

  nextTick(() => {
    textarea.focus()
    const selectionStart = start + prefix.length
    const selectionEnd = selectionStart + nextContent.length
    textarea.setSelectionRange(selectionStart, selectionEnd)
  })
}

function handleDialogKeydown(event: KeyboardEvent) {
  if (!props.modelValue) return
  if (!(event.ctrlKey || event.metaKey) || event.key.toLowerCase() !== 's') return

  const dialogElement = dialogBodyRef.value
  const target = event.target as Node | null
  if (dialogElement && target && !dialogElement.contains(target)) return

  event.preventDefault()
  void handleSubmit()
}

function handleOpen() {
  dialogOpened.value = true
  void initializeDialog()
}

function persistDraftBeforeClose() {
  if (localDraftState.value === 'dirty' || autosaveTimer.value !== null) {
    clearAutosaveTimer()
    saveLocalDraft()
  }
}

function handleClose() {
  persistDraftBeforeClose()
  dialogOpened.value = false
  emit('update:modelValue', false)
}

watch(
  () => props.modelValue,
  (visible) => {
    if (!visible) {
      dialogOpened.value = false
      clearAutosaveTimer()
    }
  }
)

watch(
  form,
  () => {
    if (suppressAutosave.value || !dialogOpened.value) return
    localDraftState.value = 'dirty'
    if (serverSaveState.value === 'saved') {
      serverSaveState.value = 'idle'
    }
    queueAutosave()
  },
  { deep: true }
)

watch(
  () => props.articleId,
  () => {
    if (props.modelValue) {
      void initializeDialog()
    }
  }
)

window.addEventListener('keydown', handleDialogKeydown)

onBeforeUnmount(() => {
  persistDraftBeforeClose()
  clearAutosaveTimer()
  window.removeEventListener('keydown', handleDialogKeydown)
})
</script>

<style scoped>
.editor-body {
  min-height: 160px;
}

.draft-alert {
  margin-bottom: 16px;
}

.draft-alert__content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap;
}

.draft-alert__actions {
  display: flex;
  gap: 8px;
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

.editor-status-panel {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  padding-top: 30px;
}

.shortcut-hint {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.content-form-item :deep(.el-form-item__content) {
  display: block;
}

.markdown-editor-shell {
  border: 1px solid var(--el-border-color);
  border-radius: 10px;
  overflow: hidden;
  background: var(--el-bg-color);
}

.markdown-toolbar {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  align-items: center;
  padding: 12px;
  border-bottom: 1px solid var(--el-border-color-light);
  background: var(--el-fill-color-light);
  flex-wrap: wrap;
}

.markdown-toolbar__left,
.markdown-toolbar__right {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.markdown-tips {
  padding: 8px 12px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
  border-bottom: 1px solid var(--el-border-color-light);
}

.markdown-workspace {
  display: grid;
  min-height: 440px;
}

.markdown-workspace.mode-split {
  grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
}

.markdown-workspace.mode-edit,
.markdown-workspace.mode-preview {
  grid-template-columns: minmax(0, 1fr);
}

.markdown-pane {
  min-width: 0;
}

.markdown-pane--editor {
  border-right: 1px solid var(--el-border-color-light);
}

.markdown-workspace.mode-edit .markdown-pane--editor,
.markdown-workspace.mode-preview .markdown-pane--editor {
  border-right: none;
}

.markdown-textarea {
  width: 100%;
  min-height: 440px;
  padding: 16px;
  border: none;
  resize: vertical;
  outline: none;
  font: inherit;
  line-height: 1.7;
  color: var(--el-text-color-primary);
  background: transparent;
}

.markdown-preview {
  min-height: 440px;
  padding: 16px;
  overflow: auto;
  line-height: 1.7;
  word-break: break-word;
}

.markdown-preview :deep(h1),
.markdown-preview :deep(h2),
.markdown-preview :deep(h3),
.markdown-preview :deep(h4) {
  margin: 1.2em 0 0.6em;
}

.markdown-preview :deep(p),
.markdown-preview :deep(ul),
.markdown-preview :deep(ol),
.markdown-preview :deep(blockquote) {
  margin: 0.8em 0;
}

.markdown-preview :deep(pre) {
  margin: 1em 0;
  padding: 12px;
  overflow: auto;
  border-radius: 8px;
  background: var(--el-fill-color-dark);
  color: #f5f7fa;
}

.markdown-preview :deep(code) {
  padding: 0.15em 0.35em;
  border-radius: 4px;
  background: var(--el-fill-color-light);
}

.markdown-preview :deep(pre code) {
  padding: 0;
  background: transparent;
}

.markdown-preview :deep(blockquote) {
  padding-left: 12px;
  border-left: 4px solid var(--el-color-primary-light-5);
  color: var(--el-text-color-regular);
}

.markdown-preview :deep(a) {
  color: var(--el-color-primary);
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  flex-wrap: wrap;
}

@media (max-width: 768px) {
  .editor-grid {
    grid-template-columns: 1fr;
  }

  .span-2 {
    grid-column: span 1;
  }

  .markdown-workspace.mode-split {
    grid-template-columns: 1fr;
  }

  .markdown-pane--editor {
    border-right: none;
    border-bottom: 1px solid var(--el-border-color-light);
  }
}
</style>
