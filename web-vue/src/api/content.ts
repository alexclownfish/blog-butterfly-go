import { apiClient } from './client'
import type { Article, ArticleListResponse, CategoryListResponse, TagListResponse } from '@/types/content'

export interface FetchArticlesParams {
  page: number
  page_size: number
  search?: string
  category_id?: number
  status?: 'published' | 'draft'
  tag?: string
}

export async function fetchArticles(params: FetchArticlesParams) {
  const { data } = await apiClient.get<ArticleListResponse>('/articles', { params })
  return data
}

export async function fetchCategories() {
  const { data } = await apiClient.get<CategoryListResponse>('/categories')
  return data
}

export async function fetchTags() {
  const { data } = await apiClient.get<TagListResponse>('/tags')
  return data
}

export async function fetchArticleDetail(id: number) {
  const { data } = await apiClient.get<{ data?: Article; error?: string }>(`/articles/${id}`)
  if (!data?.data) {
    throw new Error(data?.error || '未获取到文章详情')
  }
  return data.data
}
