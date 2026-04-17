import axios from 'axios'
import router from '@/router'
import { getToken, removeToken } from '@/utils/auth'

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

    if (status === 401) {
      removeToken()
      const currentPath = router.currentRoute.value.path
      if (currentPath !== '/login') {
        await router.replace('/login')
      }
    }

    return Promise.reject(error)
  }
)

export default client
