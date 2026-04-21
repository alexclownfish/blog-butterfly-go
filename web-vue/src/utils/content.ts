import type { Article } from '@/types/content'

const DATE_FALLBACK = '日期待同步'
const DATE_TIME_FALLBACK = '-'

function pad(value: number) {
  return String(value).padStart(2, '0')
}

function parseDate(value?: string | number | Date | null) {
  if (value === null || value === undefined || value === '') return null
  const date = value instanceof Date ? value : new Date(value)
  if (Number.isNaN(date.getTime())) return null
  return date
}

function formatDateParts(date: Date) {
  return {
    year: String(date.getFullYear()),
    month: pad(date.getMonth() + 1),
    day: pad(date.getDate()),
    hour: pad(date.getHours()),
    minute: pad(date.getMinutes()),
    second: pad(date.getSeconds())
  }
}

export function formatDate(dateString: string) {
  const date = parseDate(dateString)
  if (!date) return DATE_FALLBACK

  const { year, month, day } = formatDateParts(date)
  return `${year}-${month}-${day}`
}

export function formatDateTime(value?: string | number | Date | null, fallback = DATE_TIME_FALLBACK) {
  const date = parseDate(value)
  if (!date) return fallback

  const { year, month, day, hour, minute, second } = formatDateParts(date)
  return `${year}-${month}-${day} ${hour}:${minute}:${second}`
}

export function formatTime(value?: string | number | Date | null, fallback = DATE_TIME_FALLBACK) {
  const date = parseDate(value)
  if (!date) return fallback

  const { hour, minute, second } = formatDateParts(date)
  return `${hour}:${minute}:${second}`
}

export function tagsOf(article: Article) {
  return (article.tags || '')
    .split(',')
    .map((item) => item.trim())
    .filter(Boolean)
    .slice(0, 6)
}

export function plainSummary(article: Article) {
  const normalizedContent = article.content
    ?.replace(/[#>*`\-\n]/g, ' ')
    .replace(/\s+/g, ' ')
    .trim()

  const base = article.summary?.trim() || normalizedContent || '这篇文章还没来得及写摘要，但已经准备好让你继续深挖。'
  return base.slice(0, 140)
}

export function articleDetailPath(articleId: number) {
  return `/posts/${articleId}.html`
}

export function tagPath(tagName: string) {
  return `/tags/${encodeURIComponent(tagName)}`
}

export function withFromQuery(path: string, fromPath: string) {
  return `${path}?from=${encodeURIComponent(fromPath)}`
}
