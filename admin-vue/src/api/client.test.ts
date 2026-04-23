import { afterEach, describe, expect, it, vi } from 'vitest'

vi.mock('@/router', () => ({
  default: {
    currentRoute: { value: { path: '/' } },
    replace: vi.fn()
  }
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => ({
    logout: vi.fn(),
    setForcePasswordChange: vi.fn()
  })
}))

vi.mock('@/utils/auth', () => ({
  getToken: () => ''
}))

describe('api client runtime baseURL', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
  })

  function stubBrowser(getItemImpl: (key: string) => string | null) {
    const baseDocument = globalThis.document
    const baseWindow = globalThis.window
    vi.stubGlobal('window', Object.assign({}, baseWindow, {
      localStorage: { getItem: vi.fn(getItemImpl) },
      APP_CONFIG: undefined,
      API_BASE: undefined,
      location: baseWindow?.location || { href: 'http://localhost/' }
    }))
    vi.stubGlobal('document', Object.assign({}, baseDocument, {
      location: baseDocument?.location || { href: 'http://localhost/' },
      documentElement: { dataset: {} },
      createElement: baseDocument?.createElement?.bind(baseDocument)
    }))
  }

  it('prefers localStorage api_base override in browser runtime', async () => {
    stubBrowser((key) => (key === 'api_base' ? 'http://127.0.0.1:43083/api' : null))

    const { default: client } = await import('./client')
    const cfg = await (client.interceptors.request as any).handlers[0].fulfilled({ headers: {} })

    expect(cfg.baseURL).toBe('http://127.0.0.1:43083/api')
  })

  it('falls back to vite env api base when no runtime override exists', async () => {
    stubBrowser(() => null)

    const { default: client } = await import('./client')
    const cfg = await (client.interceptors.request as any).handlers[0].fulfilled({ headers: {} })

    expect(cfg.baseURL).toBe('/api')
  })
})
