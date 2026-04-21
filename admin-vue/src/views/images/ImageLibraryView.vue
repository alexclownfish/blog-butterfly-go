<template>
  <section class="page-section image-library-page">
    <div class="panel-card">
      <div class="section-head">
        <div>
          <div class="card-eyebrow">🖼️ Assets</div>
          <h2>素材管理</h2>
          <p>支持拖拽上传、多图整理、批量删除与快速复制链接，让配图不再满地乱跑。</p>
        </div>

        <div class="image-actions">
          <el-button :loading="refreshing" @click="loadImages">刷新列表</el-button>
          <el-button type="primary" @click="triggerFilePicker">上传图片</el-button>
        </div>
      </div>

      <input
        ref="fileInputRef"
        class="image-file-input"
        type="file"
        accept="image/*"
        multiple
        @change="handleFileChange"
      />

      <div
        class="upload-dropzone"
        :class="{ 'is-dragover': dragover }"
        @click="triggerFilePicker"
        @dragover.prevent="dragover = true"
        @dragleave.prevent="dragover = false"
        @drop.prevent="handleDrop"
      >
        <div class="upload-dropzone__icon">🪂</div>
        <h3>点击或拖拽图片到这里上传</h3>
        <p>支持多图连续上传。上传成功后可直接复制 URL，用于封面或正文插图。</p>
        <el-button type="primary" plain>选择图片</el-button>
      </div>

      <el-alert
        v-if="uploadSummary"
        :title="uploadSummary"
        :type="uploadSummaryType"
        show-icon
        class="upload-alert"
        :closable="false"
      />

      <div class="image-toolbar">
        <div class="image-toolbar__meta">
          <strong>素材库</strong>
          <span>已选择 {{ selectedKeys.length }} 张 / 共 {{ images.length }} 张</span>
        </div>

        <div class="image-toolbar__actions">
          <el-button @click="toggleSelectCurrentPage" :disabled="!pagedImages.length">
            {{ allCurrentPageSelected ? '取消本页全选' : '本页全选' }}
          </el-button>
          <el-button
            type="danger"
            :disabled="!selectedKeys.length || deleting"
            :loading="deleting"
            @click="handleDeleteSelected"
          >
            删除选中
          </el-button>
        </div>
      </div>

      <el-empty v-if="!loading && !images.length" description="当前素材库还是空空如也，先上传几张图吧～" />

      <div v-else v-loading="loading" class="image-grid">
        <article v-for="image in pagedImages" :key="image.key || image.url" class="image-card">
          <label class="image-select">
            <input
              type="checkbox"
              :checked="selectedKeySet.has(image.key)"
              @change="toggleSelection(image.key)"
            />
            <span>选择</span>
          </label>

          <div class="image-preview">
            <img :src="image.url" :alt="image.key || '素材图片'" loading="lazy" />
          </div>

          <div class="image-card__body">
            <div class="image-card__title" :title="image.key || image.url">
              {{ image.key || '未命名素材' }}
            </div>
            <div class="image-card__meta">
              <span>{{ formatFileSize(image.size) }}</span>
              <span>{{ formatDateTime(image.time || null) }}</span>
            </div>
            <div class="image-card__url" :title="image.url">{{ image.url }}</div>
          </div>

          <div class="image-card__actions">
            <el-button size="small" @click.stop="handlePreview(image.url)">预览</el-button>
            <el-button size="small" type="primary" @click.stop="handleCopy(image.url)">复制链接</el-button>
            <el-button size="small" type="danger" @click.stop="handleDeleteSingle(image)">删除</el-button>
          </div>
        </article>
      </div>

      <div class="pagination-bar" v-if="images.length > pageSize">
        <el-pagination
          background
          layout="total, prev, pager, next"
          :total="images.length"
          :current-page="page"
          :page-size="pageSize"
          @current-change="handlePageChange"
        />
      </div>
    </div>

    <el-dialog v-model="previewVisible" title="素材预览" width="760px" destroy-on-close>
      <div class="preview-dialog">
        <img :src="previewUrl" alt="素材预览" />
      </div>
    </el-dialog>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'

import { deleteImageApi, fetchImagesApi, uploadImageApi } from '@/api/images'
import type { ImageAsset } from '@/types/image'
import { formatDateTime } from '@/utils/date'

const loading = ref(false)
const refreshing = ref(false)
const deleting = ref(false)
const dragover = ref(false)
const images = ref<ImageAsset[]>([])
const selectedKeys = ref<string[]>([])
const page = ref(1)
const pageSize = 12
const fileInputRef = ref<HTMLInputElement | null>(null)
const uploadSummary = ref('')
const uploadSummaryType = ref<'success' | 'warning'>('success')
const previewVisible = ref(false)
const previewUrl = ref('')

const selectedKeySet = computed(() => new Set(selectedKeys.value))
const pagedImages = computed(() => {
  const start = (page.value - 1) * pageSize
  return images.value.slice(start, start + pageSize)
})
const allCurrentPageSelected = computed(() => {
  if (!pagedImages.value.length) return false
  return pagedImages.value.every((item) => item.key && selectedKeySet.value.has(item.key))
})

function normalizeImages(list: ImageAsset[]) {
  return [...list].sort((a, b) => (Number(b.time) || 0) - (Number(a.time) || 0))
}

function clampPage() {
  const totalPages = Math.max(1, Math.ceil(images.value.length / pageSize))
  if (page.value > totalPages) page.value = totalPages
}

function reconcileSelection() {
  const validKeys = new Set(images.value.map((item) => item.key).filter(Boolean))
  selectedKeys.value = selectedKeys.value.filter((key) => validKeys.has(key))
}

async function loadImages(options: { silent?: boolean } = {}) {
  const { silent = false } = options
  if (silent) {
    refreshing.value = true
  } else {
    loading.value = true
  }

  try {
    images.value = normalizeImages(await fetchImagesApi())
    reconcileSelection()
    clampPage()
  } catch (error: any) {
    ElMessage.error(
      error?.response?.data?.error ||
        error?.response?.data?.message ||
        error?.message ||
        '加载素材列表失败'
    )
  } finally {
    loading.value = false
    refreshing.value = false
  }
}

function triggerFilePicker() {
  fileInputRef.value?.click()
}

function resetFileInput() {
  if (fileInputRef.value) fileInputRef.value.value = ''
}

async function uploadFiles(fileList: FileList | File[]) {
  const files = Array.from(fileList || []).filter((file) => file.type.startsWith('image/'))
  if (!files.length) {
    ElMessage.warning('请选择图片文件再上传')
    return
  }

  uploadSummary.value = `准备上传 ${files.length} 张图片，请稍等～`
  uploadSummaryType.value = 'success'

  const results = await Promise.all(
    files.map(async (file) => {
      try {
        await uploadImageApi(file)
        return { success: true, name: file.name }
      } catch (error: any) {
        return {
          success: false,
          name: file.name,
          message:
            error?.response?.data?.error ||
            error?.response?.data?.message ||
            error?.message ||
            '上传失败'
        }
      }
    })
  )

  const successCount = results.filter((item) => item.success).length
  const failedItems = results.filter((item) => !item.success)

  if (failedItems.length === 0) {
    uploadSummary.value = `上传完成：成功 ${successCount} 张，素材库已补货。`
    uploadSummaryType.value = 'success'
    ElMessage.success(`成功上传 ${successCount} 张图片`)
  } else {
    const failedSummary = failedItems
      .slice(0, 3)
      .map((item) => `${item.name}：${item.message}`)
      .join('；')
    uploadSummary.value = `上传完成：成功 ${successCount} 张，失败 ${failedItems.length} 张。${failedSummary}`
    uploadSummaryType.value = 'warning'
    ElMessage.warning(`上传结果：成功 ${successCount} 张，失败 ${failedItems.length} 张`)
  }

  page.value = 1
  await loadImages({ silent: true })
  resetFileInput()
}

function handleFileChange(event: Event) {
  const input = event.target as HTMLInputElement
  if (!input.files?.length) return
  void uploadFiles(input.files)
}

function handleDrop(event: DragEvent) {
  dragover.value = false
  const files = event.dataTransfer?.files
  if (!files?.length) return
  void uploadFiles(files)
}

function toggleSelection(key: string) {
  if (!key) return
  const next = new Set(selectedKeys.value)
  if (next.has(key)) {
    next.delete(key)
  } else {
    next.add(key)
  }
  selectedKeys.value = [...next]
}

function toggleSelectCurrentPage() {
  const currentKeys = pagedImages.value.map((item) => item.key).filter(Boolean)
  const next = new Set(selectedKeys.value)

  if (allCurrentPageSelected.value) {
    currentKeys.forEach((key) => next.delete(key))
  } else {
    currentKeys.forEach((key) => next.add(key))
  }

  selectedKeys.value = [...next]
}

async function copyTextToClipboard(text: string) {
  if (navigator.clipboard?.writeText) {
    await navigator.clipboard.writeText(text)
    return
  }

  const textarea = document.createElement('textarea')
  textarea.value = text
  textarea.setAttribute('readonly', 'true')
  textarea.style.position = 'fixed'
  textarea.style.opacity = '0'
  document.body.appendChild(textarea)
  textarea.focus()
  textarea.select()

  try {
    const successful = document.execCommand('copy')
    if (!successful) {
      throw new Error('浏览器未允许复制到剪贴板')
    }
  } finally {
    document.body.removeChild(textarea)
  }
}

async function handleCopy(url: string) {
  try {
    await copyTextToClipboard(url)
    ElMessage.success('素材链接已复制')
  } catch (error: any) {
    ElMessage.error(error?.message || '复制素材链接失败')
  }
}

function handlePreview(url: string) {
  previewUrl.value = url
  previewVisible.value = true
}

async function handleDeleteByKeys(keys: string[]) {
  if (!keys.length) return
  deleting.value = true
  try {
    await Promise.all(keys.map((key) => deleteImageApi(key)))
    selectedKeys.value = selectedKeys.value.filter((key) => !keys.includes(key))
    ElMessage.success(`已删除 ${keys.length} 张素材`)
    await loadImages({ silent: true })
  } catch (error: any) {
    ElMessage.error(
      error?.response?.data?.error ||
        error?.response?.data?.message ||
        error?.message ||
        '删除素材失败'
    )
  } finally {
    deleting.value = false
  }
}

async function handleDeleteSelected() {
  if (!selectedKeys.value.length) return

  try {
    await ElMessageBox.confirm(
      `确定删除已选中的 ${selectedKeys.value.length} 张素材吗？删除后无法恢复。`,
      '批量删除素材',
      {
        type: 'warning',
        confirmButtonText: '删除',
        cancelButtonText: '取消'
      }
    )
  } catch {
    return
  }

  await handleDeleteByKeys([...selectedKeys.value])
}

async function handleDeleteSingle(image: ImageAsset) {
  if (!image.key) {
    ElMessage.error('当前素材缺少唯一 key，无法删除')
    return
  }

  try {
    await ElMessageBox.confirm(
      `确定删除素材「${image.key || image.url}」吗？删除后将无法继续复用这张图片。`,
      '删除素材',
      {
        type: 'warning',
        confirmButtonText: '删除',
        cancelButtonText: '取消'
      }
    )
  } catch {
    return
  }

  await handleDeleteByKeys([image.key])
}

function handlePageChange(nextPage: number) {
  page.value = nextPage
}

function formatFileSize(size?: number) {
  const value = Number(size) || 0
  if (!value) return '-'
  if (value < 1024) return `${value} B`
  if (value < 1024 * 1024) return `${(value / 1024).toFixed(1)} KB`
  if (value < 1024 * 1024 * 1024) return `${(value / (1024 * 1024)).toFixed(1)} MB`
  return `${(value / (1024 * 1024 * 1024)).toFixed(1)} GB`
}

onMounted(() => {
  void loadImages()
})
</script>

<style scoped>
.image-library-page {
  min-width: 0;
}

.image-actions {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.image-file-input {
  display: none;
}

.upload-dropzone {
  margin-bottom: 20px;
  border: 1.5px dashed rgba(99, 102, 241, 0.38);
  border-radius: 24px;
  padding: 28px 24px;
  text-align: center;
  background: rgba(99, 102, 241, 0.08);
  transition: 0.2s ease;
  cursor: pointer;
}

.upload-dropzone:hover,
.upload-dropzone.is-dragover {
  border-color: rgba(129, 140, 248, 0.72);
  background: rgba(99, 102, 241, 0.14);
  transform: translateY(-1px);
}

.upload-dropzone__icon {
  font-size: 38px;
  margin-bottom: 8px;
}

.upload-dropzone h3 {
  margin: 0 0 8px;
}

.upload-dropzone p {
  margin: 0 0 16px;
  color: var(--muted-strong);
}

.upload-alert {
  margin-bottom: 20px;
}

.image-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
  flex-wrap: wrap;
  margin-bottom: 20px;
}

.image-toolbar__meta {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.image-toolbar__meta span {
  color: var(--muted-strong);
  font-size: 14px;
}

.image-toolbar__actions {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.image-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
  gap: 16px;
}

.image-card {
  border-radius: 20px;
  border: 1px solid rgba(148, 163, 184, 0.18);
  background: rgba(15, 23, 42, 0.42);
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.image-select {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 14px 0;
  color: var(--muted-strong);
  font-size: 13px;
}

.image-preview {
  aspect-ratio: 16 / 10;
  padding: 14px;
}

.image-preview img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  border-radius: 16px;
  display: block;
  background: rgba(15, 23, 42, 0.65);
}

.image-card__body {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 0 14px 14px;
}

.image-card__title {
  font-weight: 700;
  line-height: 1.4;
  word-break: break-all;
}

.image-card__meta {
  display: flex;
  justify-content: space-between;
  gap: 8px;
  color: var(--muted-strong);
  font-size: 12px;
}

.image-card__url {
  color: var(--muted-strong);
  font-size: 12px;
  line-height: 1.5;
  word-break: break-all;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.image-card__actions {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 8px;
  padding: 0 14px 14px;
}

.preview-dialog {
  display: flex;
  justify-content: center;
}

.preview-dialog img {
  max-width: 100%;
  max-height: 70vh;
  border-radius: 16px;
  object-fit: contain;
}

@media (max-width: 768px) {
  .image-actions,
  .image-toolbar__actions,
  .image-card__actions {
    grid-template-columns: 1fr;
    width: 100%;
  }

  .image-card__actions {
    display: grid;
  }
}
</style>
