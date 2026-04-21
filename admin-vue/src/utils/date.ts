export const DATE_TIME_FALLBACK = '-'

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

export function formatDateTime(value?: string | number | Date | null, fallback = '-') {
  const date = parseDate(value)
  if (!date) return fallback

  const { year, month, day, hour, minute, second } = formatDateParts(date)
  return `${year}-${month}-${day} ${hour}:${minute}:${second}`
}

export function formatDate(value?: string | number | Date | null, fallback = '-') {
  const date = parseDate(value)
  if (!date) return fallback

  const { year, month, day } = formatDateParts(date)
  return `${year}-${month}-${day}`
}

export function formatTime(value?: string | number | Date | null, fallback = '-') {
  const date = parseDate(value)
  if (!date) return fallback

  const { hour, minute, second } = formatDateParts(date)
  return `${hour}:${minute}:${second}`
}
