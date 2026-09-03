// 轻量全局 toast(无依赖)。
import { reactive } from 'vue'

export type ToastType = 'success' | 'error' | 'info'

export interface ToastItem {
  id: number
  type: ToastType
  message: string
}

const state = reactive<{ list: ToastItem[] }>({ list: [] })
let seed = 0

function push(type: ToastType, message: string): void {
  const id = ++seed
  state.list.push({ id, type, message })
  window.setTimeout(() => dismiss(id), type === 'error' ? 4200 : 2600)
}

export function dismiss(id: number): void {
  const i = state.list.findIndex((t) => t.id === id)
  if (i >= 0) state.list.splice(i, 1)
}

export function toast(message: string, type: ToastType = 'info'): void {
  push(type, message)
}

export const toastState = state

export const success = (m: string) => toast(m, 'success')
export const error = (m: string) => toast(m, 'error')
