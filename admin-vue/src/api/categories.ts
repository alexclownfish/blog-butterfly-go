import client from './client'
import type { Category, CategoryPayload } from '@/types/category'

interface CategoryResponse {
  data?: Category[] | Category
  message?: string
}

export async function fetchCategoriesApi(): Promise<Category[]> {
  const { data } = await client.get<CategoryResponse>('/categories')
  return Array.isArray(data?.data) ? data.data : []
}

export async function createCategoryApi(payload: CategoryPayload): Promise<Category> {
  const { data } = await client.post<CategoryResponse>('/categories', payload)
  if (!data?.data || Array.isArray(data.data)) {
    throw new Error(data?.message || '创建分类失败')
  }
  return data.data
}

export async function updateCategoryApi(id: number, payload: CategoryPayload): Promise<Category> {
  const { data } = await client.put<CategoryResponse>(`/categories/${id}`, payload)
  if (!data?.data || Array.isArray(data.data)) {
    throw new Error(data?.message || '更新分类失败')
  }
  return data.data
}

export async function deleteCategoryApi(id: number): Promise<string> {
  const { data } = await client.delete<{ message?: string }>(`/categories/${id}`)
  return data?.message || '删除成功'
}
