export type ArticleStatus = 'draft' | 'published'

export interface Article {
  id: number
  title: string
  summary?: string
  content?: string
  cover_image?: string
  category_id?: number | null
  category?: string
  tags?: string
  is_top?: boolean
  status?: ArticleStatus
  created_at?: string
  updated_at?: string
}

export interface ArticleListQuery {
  page: number
  page_size: number
  status?: string
  search?: string
  category_id?: string | number
}

export interface ArticleListResponse {
  data?: {
    list?: Article[]
    total?: number
    page?: number
    page_size?: number
  }
  list?: Article[]
  total?: number
  message?: string
}

export interface ArticleEditorForm {
  title: string
  summary: string
  content: string
  cover_image: string
  category_id: number | null
  tags: string
  is_top: boolean
  status: ArticleStatus
}

export function createDefaultArticleForm(): ArticleEditorForm {
  return {
    title: '',
    summary: '',
    content: '',
    cover_image: '',
    category_id: null,
    tags: '',
    is_top: false,
    status: 'draft'
  }
}
