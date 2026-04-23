import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import AdminSidebar from '@/components/layout/AdminSidebar.vue'

describe('AdminSidebar', () => {
  it('renders the CSDN sync entry in navigation', () => {
    const wrapper = mount(AdminSidebar, {
      global: {
        stubs: {
          RouterLink: {
            props: ['to'],
            template: '<a :href="typeof to === \'string\' ? to : to.path"><slot /></a>'
          }
        }
      }
    })

    expect(wrapper.text()).toContain('CSDN 同步导入')
    expect(wrapper.html()).toContain('/articles/csdn-sync')
  })
})
