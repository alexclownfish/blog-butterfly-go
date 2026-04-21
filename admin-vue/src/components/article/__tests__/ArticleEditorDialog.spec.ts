import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'

import ArticleEditorDialog from '@/components/article/ArticleEditorDialog.vue'

vi.mock('@/api/articles', () => ({
  fetchArticleDetailApi: vi.fn(),
  createArticleApi: vi.fn(),
  updateArticleApi: vi.fn()
}))

vi.mock('@/api/images', () => ({
  fetchImagesApi: vi.fn().mockResolvedValue([]),
  uploadImageApi: vi.fn(),
  deleteImageApi: vi.fn()
}))

vi.mock('element-plus', async () => {
  const actual = await vi.importActual<typeof import('element-plus')>('element-plus')
  return {
    ...actual,
    ElMessage: {
      success: vi.fn(),
      error: vi.fn(),
      warning: vi.fn(),
      info: vi.fn()
    },
    ElMessageBox: {
      confirm: vi.fn()
    }
  }
})

describe('ArticleEditorDialog markdown textarea indentation', () => {
  it('shows live word count and estimated reading time for markdown content', async () => {
    const wrapper = mount(ArticleEditorDialog, {
      props: {
        modelValue: true,
        articleId: null,
        categories: []
      },
      global: {
        directives: {
          loading: {
            mounted() {}
          }
        },
        stubs: {
          ElDialog: {
            template: '<div><slot /><slot name="footer" /></div>'
          },
          ElAlert: {
            template: '<div><slot name="title" /><slot /></div>'
          },
          ElForm: {
            template: '<form><slot /></form>'
          },
          ElFormItem: {
            template: '<div><slot /></div>'
          },
          ElInput: {
            props: ['modelValue'],
            emits: ['update:modelValue'],
            methods: {
              onInput(event: Event) {
                this.$emit('update:modelValue', (event.target as HTMLInputElement).value)
              }
            },
            template: '<input :value="modelValue" @input="onInput" />'
          },
          ElSelect: {
            template: '<div><slot /></div>'
          },
          ElOption: {
            template: '<option><slot /></option>'
          },
          ElSwitch: {
            props: ['modelValue'],
            emits: ['update:modelValue'],
            methods: {
              onChange(event: Event) {
                this.$emit('update:modelValue', (event.target as HTMLInputElement).checked)
              }
            },
            template: '<input type="checkbox" :checked="modelValue" @change="onChange" />'
          },
          ElTag: {
            template: '<span><slot /></span>'
          },
          ElButton: {
            emits: ['click'],
            template: '<button type="button" @click="$emit(\'click\')"><slot /></button>'
          },
          ElButtonGroup: {
            template: '<div><slot /></div>'
          },
          ElRadioGroup: {
            template: '<div><slot /></div>'
          },
          ElRadioButton: {
            template: '<button type="button"><slot /></button>'
          },
          ElEmpty: {
            template: '<div><slot /></div>'
          },
          ElImage: {
            template: '<img />'
          },
          ElUpload: {
            template: '<div><slot /></div>'
          }
        }
      }
    })

    const textarea = wrapper.get('textarea')
    await textarea.setValue('# 标题\n\n这是用于统计阅读时长的内容。')

    expect(wrapper.text()).toContain('15 字')
    expect(wrapper.text()).toContain('预计阅读 1 分钟')
  })

  it('indents selected lines with Tab and unindents them with Shift+Tab', async () => {
    const wrapper = mount(ArticleEditorDialog, {
      props: {
        modelValue: true,
        articleId: null,
        categories: []
      },
      global: {
        directives: {
          loading: {
            mounted() {}
          }
        },
        stubs: {
          ElDialog: {
            template: '<div><slot /><slot name="footer" /></div>'
          },
          ElAlert: {
            template: '<div><slot name="title" /><slot /></div>'
          },
          ElForm: {
            template: '<form><slot /></form>'
          },
          ElFormItem: {
            template: '<div><slot /></div>'
          },
          ElInput: {
            props: ['modelValue'],
            emits: ['update:modelValue'],
            methods: {
              onInput(event: Event) {
                this.$emit('update:modelValue', (event.target as HTMLInputElement).value)
              }
            },
            template: '<input :value="modelValue" @input="onInput" />'
          },
          ElSelect: {
            template: '<div><slot /></div>'
          },
          ElOption: {
            template: '<option><slot /></option>'
          },
          ElSwitch: {
            props: ['modelValue'],
            emits: ['update:modelValue'],
            methods: {
              onChange(event: Event) {
                this.$emit('update:modelValue', (event.target as HTMLInputElement).checked)
              }
            },
            template: '<input type="checkbox" :checked="modelValue" @change="onChange" />'
          },
          ElTag: {
            template: '<span><slot /></span>'
          },
          ElButton: {
            emits: ['click'],
            template: '<button type="button" @click="$emit(\'click\')"><slot /></button>'
          },
          ElButtonGroup: {
            template: '<div><slot /></div>'
          },
          ElRadioGroup: {
            template: '<div><slot /></div>'
          },
          ElRadioButton: {
            template: '<button type="button"><slot /></button>'
          },
          ElEmpty: {
            template: '<div><slot /></div>'
          },
          ElImage: {
            template: '<img />'
          },
          ElUpload: {
            template: '<div><slot /></div>'
          }
        }
      }
    })

    const textarea = wrapper.get('textarea')
    await textarea.setValue('first\nsecond')

    const element = textarea.element as HTMLTextAreaElement
    element.setSelectionRange(0, element.value.length)

    const tabEvent = new KeyboardEvent('keydown', {
      key: 'Tab',
      bubbles: true,
      cancelable: true
    })
    element.dispatchEvent(tabEvent)
    await wrapper.vm.$nextTick()

    expect(tabEvent.defaultPrevented).toBe(true)
    expect(element.value).toBe('  first\n  second')
    expect(element.selectionStart).toBe(0)
    expect(element.selectionEnd).toBe(element.value.length)

    element.setSelectionRange(0, element.value.length)

    const shiftTabEvent = new KeyboardEvent('keydown', {
      key: 'Tab',
      shiftKey: true,
      bubbles: true,
      cancelable: true
    })
    element.dispatchEvent(shiftTabEvent)
    await wrapper.vm.$nextTick()

    expect(shiftTabEvent.defaultPrevented).toBe(true)
    expect(element.value).toBe('first\nsecond')
    expect(element.selectionStart).toBe(0)
    expect(element.selectionEnd).toBe(element.value.length)
  })
})

