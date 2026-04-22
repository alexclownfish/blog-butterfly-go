import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import ArticleListView from '@/views/articles/ArticleListView.vue'

const articleApiMocks = vi.hoisted(() => ({
  fetchArticlesApi: vi.fn(),
  deleteArticleApi: vi.fn()
}))

const categoryApiMocks = vi.hoisted(() => ({
  fetchCategoriesApi: vi.fn()
}))

const elementPlusMocks = vi.hoisted(() => ({
  success: vi.fn(),
  error: vi.fn(),
  confirm: vi.fn()
}))

vi.mock('@/api/articles', () => ({
  fetchArticlesApi: articleApiMocks.fetchArticlesApi,
  deleteArticleApi: articleApiMocks.deleteArticleApi
}))

vi.mock('@/api/categories', () => ({
  fetchCategoriesApi: categoryApiMocks.fetchCategoriesApi
}))

vi.mock('element-plus', async () => {
  const actual = await vi.importActual<typeof import('element-plus')>('element-plus')
  return {
    ...actual,
    ElMessage: {
      success: elementPlusMocks.success,
      error: elementPlusMocks.error
    },
    ElMessageBox: { confirm: elementPlusMocks.confirm }
  }
})

vi.mock('@/components/article/ArticleEditorDialog.vue', () => ({
  default: {
    name: 'ArticleEditorDialog',
    props: ['modelValue', 'articleId', 'categories'],
    emits: ['update:modelValue', 'saved'],
    template: '<div class="article-editor-dialog-stub"></div>'
  }
}))

vi.mock('@/components/article/CsdnImportDialog.vue', () => ({
  default: {
    name: 'CsdnImportDialog',
    props: ['modelValue', 'categories'],
    emits: ['update:modelValue', 'imported'],
    template: '<div class="csdn-import-dialog-stub"></div>'
  }
}))

describe('ArticleListView', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    categoryApiMocks.fetchCategoriesApi.mockResolvedValue([{ id: 7, name: 'Golang' }])
    articleApiMocks.fetchArticlesApi.mockResolvedValue({
      list: [
        {
          id: 1,
          title: 'Go 并发实战',
          category: 'Golang',
          status: 'draft',
          is_top: false,
          updated_at: '2026-04-22T10:00:00Z'
        }
      ],
      total: 1
    })
  })

  it('renders csdn import entry and refreshes list after imported event', async () => {
    const wrapper = mount(ArticleListView, {
      global: {
        directives: {
          loading: { mounted() {} }
        },
        stubs: {
          ElButton: {
            emits: ['click'],
            template: '<button type="button" @click="$emit(\'click\')"><slot /></button>'
          },
          ElInput: {
            props: ['modelValue'],
            emits: ['update:modelValue'],
            template: '<input :value="modelValue" @input="$emit(\'update:modelValue\', $event.target.value)" />'
          },
          ElSelect: {
            props: ['modelValue'],
            emits: ['update:modelValue'],
            template: '<select :value="modelValue" @change="$emit(\'update:modelValue\', $event.target.value)"><slot /></select>'
          },
          ElOption: {
            props: ['value', 'label'],
            template: '<option :value="value">{{ label }}</option>'
          },
          ElTable: { template: '<div><slot /></div>' },
          ElTableColumn: { template: '<div><slot :row="{}" /></div>' },
          ElTag: { template: '<span><slot /></span>' },
          ElPagination: { template: '<div />' }
        }
      }
    })

    await flushPromises()

    expect(categoryApiMocks.fetchCategoriesApi).toHaveBeenCalledTimes(1)
    expect(articleApiMocks.fetchArticlesApi).toHaveBeenCalledTimes(1)
    expect(wrapper.text()).toContain('导入 CSDN')

    const csdnDialog = wrapper.findComponent({ name: 'CsdnImportDialog' })
    expect(csdnDialog.exists()).toBe(true)
    expect(csdnDialog.props('categories')).toEqual([{ id: 7, name: 'Golang' }])

    await csdnDialog.vm.$emit('imported', { id: 99, title: '导入文章' })
    await flushPromises()

    expect(articleApiMocks.fetchArticlesApi).toHaveBeenCalledTimes(2)
  })
})
