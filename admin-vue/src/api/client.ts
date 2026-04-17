import axios from 'axios'
import router from '@/router'
import { useAuthStore } from '@/stores/auth'
import { getToken } from '@/utils/auth'

const client = axios.create({
  baseURL: import.meta.env.VITE_API_BASE,
  timeout: 15000,
  headers: {
    'Content-Type': 'application/json'
  }
})

client.interceptors.request.use((config) => {
  const token = getToken()
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

client.interceptors.response.use(
  (response) => response,
  async (error) => {
    const status = error?.response?.status
    const forcePasswordChange = Boolean(error?.response?.data?.force_password_change)
    const authStore = useAuthStore()

    if (status === 401) {
      authStore.logout()
      const currentPath = router.currentRoute.value.path
      if (currentPath !== '/login') {
        await router.replace('/login')
      }
    }

    if (status === 403 && forcePasswordChange) {
      authStore.setForcePasswordChange(true)
      if (router.currentRoute.value.path !== '/change-password') {
        await router.replace('/change-password')
      }
    }

    return Promise.reject(error)
  }
)

export default client
