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
