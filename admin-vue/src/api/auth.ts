import client from './client'
import type { ChangePasswordPayload, LoginPayload, LoginApiResponse, LoginResult } from '@/types/auth'

export async function loginApi(payload: LoginPayload): Promise<LoginResult> {
  const { data } = await client.post<LoginApiResponse>('/login', payload)
  const token = data?.token || data?.data?.token || ''

  if (!token) {
    throw new Error(data?.message || '登录成功但未返回 token')
  }

  return {
    token,
    forcePasswordChange: Boolean(data?.force_password_change ?? data?.data?.force_password_change)
  }
}

export async function changePasswordApi(payload: ChangePasswordPayload) {
  const { data } = await client.post('/change-password', payload)
  return data
}
