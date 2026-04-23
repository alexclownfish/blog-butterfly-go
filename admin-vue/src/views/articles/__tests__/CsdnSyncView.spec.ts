import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import CsdnSyncView from '@/views/articles/CsdnSyncView.vue'

const articleApiMocks = vi.hoisted(() => ({
  startCsdnSyncLoginApi: vi.fn(),
  fetchCsdnSyncSessionApi: vi.fn(),
  importCsdnSyncArticleApi: vi.fn()
}))

const categoryApiMocks = vi.hoisted(() => ({
  fetchCategoriesApi: vi.fn()
}))

const elementPlusMocks = vi.hoisted(() => ({
  success: vi.fn(),
  error: vi.fn()
}))

vi.mock('@/api/articles', () => ({
  startCsdnSyncLoginApi: articleApiMocks.startCsdnSyncLoginApi,
  fetchCsdnSyncSessionApi: articleApiMocks.fetchCsdnSyncSessionApi,
  importCsdnSyncArticleApi: articleApiMocks.importCsdnSyncArticleApi
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
    }
  }
})

describe('CsdnSyncView', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    categoryApiMocks.fetchCategoriesApi.mockResolvedValue([{ id: 7, name: 'Golang' }])
    articleApiMocks.startCsdnSyncLoginApi.mockResolvedValue({
      id: 'sess-1',
      provider: 'csdn',
      provider_mode: 'stub',
      status: 'pending',
      message: '请使用 CSDN App 扫码',
      qr_code_data_url: 'data:image/svg+xml;utf8,test',
      expires_at: '2026-04-23T10:05:00Z',
      created_at: '2026-04-23T10:03:00Z',
      updated_at: '2026-04-23T10:03:00Z',
      articles: []
    })
    articleApiMocks.fetchCsdnSyncSessionApi.mockResolvedValue({
      id: 'sess-1',
      provider: 'csdn',
      provider_mode: 'stub',
      status: 'authorized',
      message: '已授权，可导入文章',
      qr_code_data_url: 'data:image/svg+xml;utf8,test',
      expires_at: '2026-04-23T10:05:00Z',
      created_at: '2026-04-23T10:03:00Z',
      updated_at: '2026-04-23T10:04:00Z',
      articles: [
        {
          id: 'remote-1',
          title: 'Go 并发实战',
          summary: '讲清 goroutine',
          source_url: 'https://blog.csdn.net/test/article/details/123',
          published_at: '2026-04-22T10:00:00Z'
        }
      ]
    })
    articleApiMocks.importCsdnSyncArticleApi.mockResolvedValue({ id: 99, title: 'Go 并发实战' })
  })

  it('starts login, refreshes authorized session and imports selected article', async () => {
    const wrapper = mount(CsdnSyncView, {
      global: {
        directives: {
          loading: { mounted() {} }
        },
        stubs: {
          ElButton: {
            emits: ['click'],
            template: '<button type="button" @click="$emit(\'click\')"><slot /></button>'
          },
          ElCard: { template: '<div><slot /></div>' },
          ElAlert: { template: '<div><slot /></div>' },
          ElTag: { template: '<span><slot /></span>' },
          ElEmpty: { template: '<div><slot /></div>' },
          ElDivider: { template: '<div><slot /></div>' },
          ElRadioGroup: { template: '<div><slot /></div>' },
          ElRadio: { template: '<label><slot /></label>' },
          ElInput: {
            props: ['modelValue'],
            emits: ['update:modelValue'],
            template: '<input :value="modelValue" @input="$emit(\'update:modelValue\', $event.target.value)" />'
          },
          ElSelect: {
            props: ['modelValue'],
            emits: ['update:modelValue'],
            template: '<select :value="modelValue" @change="$emit(\'update:modelValue\', Number($event.target.value))"><slot /></select>'
          },
          ElOption: {
            props: ['value', 'label'],
            template: '<option :value="value">{{ label }}</option>'
          }
        }
      }
    })

    await flushPromises()
    expect(categoryApiMocks.fetchCategoriesApi).toHaveBeenCalledTimes(1)

    const startButton = wrapper.findAll('button').find((button) => button.text().includes('开始扫码登录'))
    expect(startButton).toBeDefined()
    await startButton!.trigger('click')
    await flushPromises()

    expect(articleApiMocks.startCsdnSyncLoginApi).toHaveBeenCalledTimes(1)
    expect(wrapper.text()).toContain('请使用 CSDN App 扫码')

    const refreshButton = wrapper.findAll('button').find((button) => button.text().includes('刷新登录状态'))
    expect(refreshButton).toBeDefined()
    await refreshButton!.trigger('click')
    await flushPromises()

    expect(articleApiMocks.fetchCsdnSyncSessionApi).toHaveBeenCalledWith('sess-1')
    expect(wrapper.text()).toContain('Go 并发实战')

    const select = wrapper.find('select')
    await select.setValue('7')

    const importButton = wrapper.findAll('button').find((button) => button.text().includes('导入到当前博客'))
    expect(importButton).toBeDefined()
    await importButton!.trigger('click')
    await flushPromises()

    expect(articleApiMocks.importCsdnSyncArticleApi).toHaveBeenCalledWith({
      session_id: 'sess-1',
      article_id: 'remote-1',
      category_id: 7,
      status: 'draft'
    })
    expect(elementPlusMocks.success).toHaveBeenCalled()
  })
})
