// 轻量 fetch 客户端:统一鉴权头、响应壳解包、错误消息化。
export const TOKEN_KEY = 'ticket_token'
export const USER_KEY = 'ticket_user'

export function getToken(): string {
  return localStorage.getItem(TOKEN_KEY) || ''
}

export function setToken(token: string): void {
  localStorage.setItem(TOKEN_KEY, token)
}

export function clearAuthStorage(): void {
  localStorage.removeItem(TOKEN_KEY)
  localStorage.removeItem(USER_KEY)
}

export class ApiError extends Error {
  code: number
  status: number
  constructor(code: number, status: number, message: string) {
    super(message)
    this.code = code
    this.status = status
  }
}

interface Envelope<T> {
  code: number
  message: string
  data: T
}

async function request<T>(method: string, path: string, body?: unknown, form?: FormData): Promise<T> {
  const headers: Record<string, string> = {}
  let payload: BodyInit | undefined
  if (form) {
    payload = form
  } else if (body !== undefined) {
    headers['Content-Type'] = 'application/json'
    payload = JSON.stringify(body)
  }
  const token = getToken()
  if (token) headers['Authorization'] = `Bearer ${token}`

  let resp: Response
  try {
    resp = await fetch('/api' + path, { method, headers, body: payload })
  } catch {
    throw new ApiError(-1, 0, '网络异常,请检查连接')
  }

  const text = await resp.text()
  let json: Envelope<T> | null = null
  try {
    json = text ? (JSON.parse(text) as Envelope<T>) : null
  } catch {
    /* 非 JSON(如微信公众号 XML) */
  }

  if (resp.status === 401) {
    clearAuthStorage()
    const current = window.location.pathname + window.location.search
    if (!current.startsWith('/login') && !current.startsWith('/register')) {
      window.location.assign('/login?redirect=' + encodeURIComponent(current))
    }
    throw new ApiError(json?.code ?? 401, resp.status, json?.message || '登录已过期')
  }

  if (!resp.ok || !json || json.code !== 0) {
    throw new ApiError(json?.code ?? resp.status, resp.status, json?.message || `请求失败(${resp.status})`)
  }
  return json.data
}

export const http = {
  get: <T>(path: string) => request<T>('GET', path),
  post: <T>(path: string, body?: unknown) => request<T>('POST', path, body),
  put: <T>(path: string, body?: unknown) => request<T>('PUT', path, body),
  del: <T>(path: string) => request<T>('DELETE', path),
  upload: <T>(path: string, file: File | Blob, filename: string, type: string) => {
    const fd = new FormData()
    fd.append('file', file, filename)
    fd.append('type', type)
    return request<T>('POST', path, undefined, fd)
  }
}
