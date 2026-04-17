import { defineStore } from 'pinia'
import { changePasswordApi, loginApi } from '@/api/auth'
import {
  getForcePasswordChange,
  getToken,
  removeForcePasswordChangeStorage,
  removeToken,
  setForcePasswordChangeStorage,
  setToken
} from '@/utils/auth'
import type { ChangePasswordPayload, LoginPayload } from '@/types/auth'

interface AuthState {
  token: string
  loading: boolean
  forcePasswordChange: boolean
}

export const useAuthStore = defineStore('auth', {
  state: (): AuthState => ({
    token: '',
    loading: false,
    forcePasswordChange: false
  }),

  getters: {
    isLoggedIn: (state) => Boolean(state.token)
  },

  actions: {
    restoreToken() {
      this.token = getToken()
      this.forcePasswordChange = getForcePasswordChange()
    },

    setTokenValue(token: string) {
      this.token = token
      setToken(token)
    },

    setForcePasswordChange(value: boolean) {
      this.forcePasswordChange = value
      setForcePasswordChangeStorage(value)
    },

    clearToken() {
      this.token = ''
      this.forcePasswordChange = false
      removeToken()
      removeForcePasswordChangeStorage()
    },

    async login(payload: LoginPayload) {
      this.loading = true
      try {
        const result = await loginApi(payload)
        this.setTokenValue(result.token)
        this.setForcePasswordChange(result.forcePasswordChange)
        return result
      } finally {
        this.loading = false
      }
    },

    async changePassword(payload: ChangePasswordPayload) {
      this.loading = true
      try {
        const result = await changePasswordApi(payload)
        this.setForcePasswordChange(false)
        return result
      } finally {
        this.loading = false
      }
    },

    logout() {
      this.clearToken()
    }
  }
})
