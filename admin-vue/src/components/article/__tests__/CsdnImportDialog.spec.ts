import { flushPromises, mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'

import CsdnImportDialog from '@/components/article/CsdnImportDialog.vue'

const articleApiMocks = vi.hoisted(() => ({
  previewImportCsdnApi: vi.fn(),
  importCsdnArticleApi: vi.fn()
}))

const elementPlusMocks = vi.hoisted(() => ({
  success: vi.fn(),
  error: vi.fn()
}))

vi.mock('@/api/articles', () => ({
  previewImportCsdnApi: articleApiMocks.previewImportCsdnApi,
  importCsdnArticleApi: articleApiMocks.importCsdnArticleApi
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

describe('CsdnImportDialog', () => {
  it('loads preview and submits import as draft', async () => {
    articleApiMocks.previewImportCsdnApi.mockResolvedValueOnce({
      title: 'Go 并发实战',
      summary: '讲清 goroutine',
      content: '## Go 并发实战',
      cover_image: 'https://img.example.com/cover.png',
      tags: 'Go,并发',
      source_url: 'https://blog.csdn.net/test/article/details/123',
      source_platform: 'csdn'
    })
    articleApiMocks.importCsdnArticleApi.mockResolvedValueOnce({ id: 99, title: 'Go 并发实战' })

    const wrapper = mount(CsdnImportDialog, {
      props: {
        modelValue: true,
        categories: [{ id: 7, name: 'Golang' }]
      },
      global: {
        directives: {
          loading: { mounted() {} }
        },
        stubs: {
          ElDialog: { template: '<div><slot /><slot name="footer" /></div>' },
          ElEmpty: { template: '<div class="el-empty-stub"></div>' },
          ElForm: { template: '<form><slot /></form>' },
          ElFormItem: { template: '<div><slot /></div>' },
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
          },
          ElRadioGroup: { template: '<div><slot /></div>' },
          ElRadioButton: { template: '<button type="button"><slot /></button>' },
          ElButton: {
            emits: ['click'],
            template: '<button type="button" @click="$emit(\'click\')"><slot /></button>'
          },
          ElAlert: { template: '<div><slot /></div>' },
          ElTag: { template: '<span><slot /></span>' },
          ElScrollbar: { template: '<div><slot /></div>' }
        }
      }
    })

    const inputs = wrapper.findAll('input')
    await inputs[0].setValue('https://blog.csdn.net/test/article/details/123')
    await wrapper.findAll('button')[0].trigger('click')
    await flushPromises()

    expect(articleApiMocks.previewImportCsdnApi).toHaveBeenCalledWith({
      url: 'https://blog.csdn.net/test/article/details/123'
    })
    expect(wrapper.text()).toContain('Go 并发实战')

    const select = wrapper.find('select')
    await select.setValue('7')
    const importButton = wrapper
      .findAll('button')
      .find((button) => button.text().includes('立即导入'))

    expect(importButton).toBeDefined()
    await importButton!.trigger('click')
    await flushPromises()

    expect(articleApiMocks.importCsdnArticleApi).toHaveBeenCalledWith({
      url: 'https://blog.csdn.net/test/article/details/123',
      category_id: 7,
      status: 'draft'
    })
    expect(elementPlusMocks.success).toHaveBeenCalled()
    expect(wrapper.emitted('imported')).toBeTruthy()
  })
})
