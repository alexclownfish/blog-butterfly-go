export interface ImageAsset {
  url: string
  key: string
  size?: number
  time?: number
}

export interface ImageListResponse {
  data?: Array<Partial<ImageAsset>>
  message?: string
  error?: string
}

export interface ImageUploadResponse {
  url?: string
  message?: string
  error?: string
}

export interface ImageDeleteResponse {
  message?: string
  error?: string
}
