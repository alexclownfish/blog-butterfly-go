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

export interface CsdnArticlePreview {
  title: string
  summary?: string
  content?: string
  cover_image?: string
  tags?: string
  source_url?: string
  source_platform?: string
}

export interface CsdnArticleImportPayload {
  url: string
  category_id: number
  status: ArticleStatus
}

export type CsdnSyncSessionStatus = 'pending' | 'scanned' | 'authorized' | 'expired' | 'failed'

export interface CsdnSyncRemoteArticle {
  id: string
  title: string
  summary?: string
  cover_image?: string
  source_url?: string
  published_at?: string
}

export interface CsdnSyncSession {
  id: string
  user_id?: number
  provider: string
  provider_mode: string
  provider_session?: string
  status: CsdnSyncSessionStatus
  message?: string
  error_message?: string
  qr_code_data_url?: string
  expires_at?: string
  created_at?: string
  updated_at?: string
  articles?: CsdnSyncRemoteArticle[]
}

export interface CsdnSyncArticleImportPayload {
  session_id: string
  article_id: string
  category_id: number
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
