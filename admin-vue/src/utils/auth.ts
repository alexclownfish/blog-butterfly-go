const TOKEN_KEY = 'admin-vue:token'
const FORCE_PASSWORD_CHANGE_KEY = 'admin-vue:force-password-change'

export function getToken(): string {
  return localStorage.getItem(TOKEN_KEY) || ''
}

export function setToken(token: string) {
  localStorage.setItem(TOKEN_KEY, token)
}

export function removeToken() {
  localStorage.removeItem(TOKEN_KEY)
}

export function getForcePasswordChange(): boolean {
  return localStorage.getItem(FORCE_PASSWORD_CHANGE_KEY) === 'true'
}

export function setForcePasswordChangeStorage(value: boolean) {
  localStorage.setItem(FORCE_PASSWORD_CHANGE_KEY, String(Boolean(value)))
}

export function removeForcePasswordChangeStorage() {
  localStorage.removeItem(FORCE_PASSWORD_CHANGE_KEY)
}
