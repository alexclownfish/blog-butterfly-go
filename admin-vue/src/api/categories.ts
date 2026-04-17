import client from './client'
import type { Category } from '@/types/category'

interface CategoryResponse {
  data?: Category[]
  message?: string
}

export async function fetchCategoriesApi(): Promise<Category[]> {
  const { data } = await client.get<CategoryResponse>('/categories')
  return Array.isArray(data?.data) ? data.data : []
}
