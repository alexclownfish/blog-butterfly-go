export interface LoginPayload {
  username: string
  password: string
}

export interface ChangePasswordPayload {
  old_password: string
  new_password: string
  confirm_password: string
}

export interface LoginApiResponse {
  token?: string
  force_password_change?: boolean
  data?: {
    token?: string
    force_password_change?: boolean
  }
  message?: string
}

export interface LoginResult {
  token: string
  forcePasswordChange: boolean
}
