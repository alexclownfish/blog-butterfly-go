import client from './client'
import type { LoginPayload, LoginApiResponse } from '@/types/auth'

export async function loginApi(payload: LoginPayload): Promise<string> {
  const { data } = await client.post<LoginApiResponse>('/login', payload)
  const token = data?.token || data?.data?.token || ''

  if (!token) {
    throw new Error(data?.message || '登录成功但未返回 token')
  }

  return token
}
