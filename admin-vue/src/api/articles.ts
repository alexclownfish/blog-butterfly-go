import client from './client'
import type {
  Article,
  ArticleEditorForm,
  ArticleListQuery,
  ArticleListResponse,
  CsdnArticleImportPayload,
  CsdnArticlePreview
} from '@/types/article'

export async function fetchArticlesApi(params: ArticleListQuery) {
  const { data } = await client.get<ArticleListResponse>('/articles', { params })

  const list = data?.data?.list || data?.list || data?.data || []
  const total = data?.data?.total || data?.total || (Array.isArray(list) ? list.length : 0)

  return {
    list: Array.isArray(list) ? list : [],
    total: Number(total) || 0
  }
}

export async function fetchArticleDetailApi(id: number): Promise<Article> {
  const { data } = await client.get<{ data?: Article; message?: string }>(`/articles/${id}`)
  if (!data?.data) {
    throw new Error(data?.message || '未获取到文章详情')
  }
  return data.data
}

export async function createArticleApi(payload: ArticleEditorForm) {
  const { data } = await client.post('/articles', payload)
  return data
}

export async function updateArticleApi(id: number, payload: ArticleEditorForm) {
  const { data } = await client.put(`/articles/${id}`, payload)
  return data
}

export async function deleteArticleApi(id: number) {
  const { data } = await client.delete(`/articles/${id}`)
  return data
}

export async function previewImportCsdnApi(payload: { url: string }): Promise<CsdnArticlePreview> {
  const { data } = await client.post<{ data?: CsdnArticlePreview; message?: string }>(
    '/articles/import/csdn/preview',
    payload
  )
  if (!data?.data) {
    throw new Error(data?.message || '未获取到 CSDN 预览结果')
  }
  return data.data
}

export async function importCsdnArticleApi(payload: CsdnArticleImportPayload): Promise<Article> {
  const { data } = await client.post<{ data?: Article; message?: string }>(
    '/articles/import/csdn',
    payload
  )
  if (!data?.data) {
    throw new Error(data?.message || '导入 CSDN 文章失败')
  }
  return data.data
}
