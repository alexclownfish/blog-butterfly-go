import { defineStore } from 'pinia'
import { loginApi } from '@/api/auth'
import { getToken, removeToken, setToken } from '@/utils/auth'
import type { LoginPayload } from '@/types/auth'

interface AuthState {
  token: string
  loading: boolean
}

export const useAuthStore = defineStore('auth', {
  state: (): AuthState => ({
    token: '',
    loading: false
  }),

  getters: {
    isLoggedIn: (state) => Boolean(state.token)
  },

  actions: {
    restoreToken() {
      this.token = getToken()
    },

    setTokenValue(token: string) {
      this.token = token
      setToken(token)
    },

    clearToken() {
      this.token = ''
      removeToken()
    },

    async login(payload: LoginPayload) {
      this.loading = true
      try {
        const token = await loginApi(payload)
        this.setTokenValue(token)
        return token
      } finally {
        this.loading = false
      }
    },

    logout() {
      this.clearToken()
    }
  }
})
