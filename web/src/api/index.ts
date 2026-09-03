import { http } from './http'
import type {
  AdminImageRow,
  CaptchaResult,
  LoginResult,
  Paged,
  ProductImage,
  Template,
  UploadResult,
  User
} from '@/types'

// ---------- auth ----------
export const authApi = {
  captcha: () => http.post<CaptchaResult>('/auth/captcha'),
  register: (nickname: string, password: string, captchaId: string, captcha: string) =>
    http.post<User>('/auth/register', { nickname, password, captchaId, captcha }),
  login: (nickname: string, password: string) =>
    http.post<LoginResult>('/auth/login', { nickname, password }),
  resetPassword: (nickname: string, password: string, captchaId: string, captcha: string) =>
    http.post('/auth/reset-password', { nickname, password, captchaId, captcha }),
  me: () => http.get<User>('/auth/me'),
  updateProfile: (payload: { nickname?: string; avatar?: string }) =>
    http.put<User>('/auth/profile', payload)
}

// ---------- wechat ----------
export const wechatApi = {
  login: (code: string) => http.post<LoginResult & { isNew: boolean }>('/wechat/login', { code })
}

// ---------- templates ----------
export const templateApi = {
  listPublic: () => http.get<{ list: Template[]; total: number }>('/templates')
}

// ---------- uploads / products ----------
export const uploadApi = {
  upload: (file: File | Blob, filename: string, type: string) =>
    http.upload<UploadResult>('/uploads', file, filename, type)
}

export const productApi = {
  list: (page = 1, size = 12) =>
    http.get<Paged<ProductImage>>(`/user/products?page=${page}&size=${size}`),
  remove: (id: number) => http.del(`/user/products/${id}`)
}

// ---------- admin ----------
export const adminApi = {
  templates: (status = 0) =>
    http.get<{ list: Template[]; total: number }>(`/admin/templates?status=${status}`),
  create: (payload: Partial<Template>) => http.post<Template>('/admin/templates', payload),
  update: (id: number, payload: Partial<Template>) =>
    http.put<Template>(`/admin/templates/${id}`, payload),
  remove: (id: number) => http.del(`/admin/templates/${id}`),
  images: (type: 'product' | 'poster', page = 1, size = 24) =>
    http.get<Paged<AdminImageRow>>(`/admin/images?type=${type}&page=${page}&size=${size}`),
  removeImage: (id: number) => http.del(`/admin/images/${id}`)
}
