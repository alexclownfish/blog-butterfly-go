import type { Article } from '@/types/content'

export function formatDate(dateString: string) {
  const date = new Date(dateString)
  if (Number.isNaN(date.getTime())) return '日期待同步'
  return new Intl.DateTimeFormat('zh-CN', {
    year: 'numeric',
    month: 'short',
    day: 'numeric'
  }).format(date)
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
