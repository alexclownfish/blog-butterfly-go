export interface Category {
  id: number
  name: string
}

export interface Tag {
  id: number
  name: string
}

export interface Article {
  id: number
  title: string
  content: string
  summary: string
  cover_image: string
  category_id: number
  category?: Category | null
  tags: string
  is_top: boolean
  status: 'draft' | 'published'
  views: number
  created_at: string
  updated_at: string
}

export interface ArticleListResponse {
  data: Article[]
  total: number
  page: number
  page_size: number
}

export interface CategoryListResponse {
  data: Category[]
}

export interface TagListResponse {
  data: Tag[]
}
