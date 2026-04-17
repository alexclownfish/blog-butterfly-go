export interface LoginPayload {
  username: string
  password: string
}

export interface LoginApiResponse {
  token?: string
  data?: {
    token?: string
  }
  message?: string
}
