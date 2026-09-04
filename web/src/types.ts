// 共享领域类型(与后端 JSON 契约一致)
export interface User {
  id: number
  nickname: string
  avatar: string
  role: number // 0 普通 1 管理员
  openid: string
  createTime?: number
}

export interface LoginResult {
  token: string
  user: User
}

export interface Template {
  id: number
  name: string
  url: string
  titleColor: string
  textColor: string
  status: number
  createTime?: number
}

export interface ProductImage {
  id: number
  type: string
  filename: string
  url: string
  ip?: string
  createTime: number
}

// 管理后台列表行(图片 + 上传者昵称)
export interface AdminImageRow {
  id: number
  type: string
  userId: number
  filename: string
  url: string
  thumbUrl?: string
  ip?: string
  nickname: string
  createTime: number
}

// 管理后台用户行
export interface AdminUser {
  id: number
  nickname: string
  avatar: string
  openid: string
  role: number // 0 普通 / 1 管理员
  status: number // 1 正常 / 0 冻结
  createTime?: number
  lastLoginTime?: number | null
}

export interface Paged<T> {
  list: T[]
  total: number
  page: number
  size: number
}

export interface UploadResult {
  id: number
  url: string
  filename: string
  type: string
  createTime: number
}

export interface CaptchaResult {
  captchaId: string
  captchaImg: string
}

export interface TicketFields {
  movieTitle: string
  language: string
  format: string
  cinemaName: string
  hallSeat: string
  startTime: string
  exclusiveName: string
  titleFontSize: number
  textFontSize: number
}

export const DEFAULT_FIELDS: TicketFields = {
  movieTitle: '',
  language: '国语',
  format: '2D',
  cinemaName: '',
  hallSeat: '',
  startTime: '',
  exclusiveName: '',
  titleFontSize: 48,
  textFontSize: 26
}
