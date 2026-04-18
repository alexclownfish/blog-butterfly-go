import { apiClient } from './client'
import type { ArticleListResponse, CategoryListResponse } from '@/types/content'

export interface FetchArticlesParams {
  page: number
  page_size: number
  search?: string
  category_id?: number
  status?: 'published' | 'draft'
}

export async function fetchArticles(params: FetchArticlesParams) {
  const { data } = await apiClient.get<ArticleListResponse>('/articles', { params })
  return data
}

export async function fetchCategories() {
  const { data } = await apiClient.get<CategoryListResponse>('/categories')
  return data
}
