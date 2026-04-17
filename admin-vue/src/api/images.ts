import client from './client'
import type { ImageAsset, ImageListResponse, ImageUploadResponse } from '@/types/image'

function normalizeImageAsset(item: Partial<ImageAsset> | null | undefined): ImageAsset | null {
  const url = typeof item?.url === 'string' ? item.url.trim() : ''
  const key = typeof item?.key === 'string' ? item.key.trim() : ''

  if (!url) return null

  return {
    url,
    key,
    size: typeof item?.size === 'number' ? item.size : Number(item?.size) || 0,
    time: typeof item?.time === 'number' ? item.time : Number(item?.time) || 0
  }
}

export async function fetchImagesApi(): Promise<ImageAsset[]> {
  const { data } = await client.get<ImageListResponse>('/images')
  const list = Array.isArray(data?.data) ? data.data : []

  return list
    .map((item) => normalizeImageAsset(item))
    .filter((item): item is ImageAsset => Boolean(item))
}

export async function uploadImageApi(file: File): Promise<string> {
  const formData = new FormData()
  formData.append('image', file)

  const { data } = await client.post<ImageUploadResponse>('/upload', formData, {
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  })

  return typeof data?.url === 'string' ? data.url.trim() : ''
}
