import { defineStore } from 'pinia'
import { authApi, wechatApi } from '@/api'
import { getToken, setToken, clearAuthStorage } from '@/api/http'
import type { User } from '@/types'

function readUser(): User | null {
  try {
    const raw = localStorage.getItem('ticket_user')
    return raw ? (JSON.parse(raw) as User) : null
  } catch {
    return null
  }
}

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: getToken(),
    user: readUser()
  }),
  getters: {
    isAuthed: (s) => !!s.token,
    isAdmin: (s) => s.user?.role === 1
  },
  actions: {
    applyLogin(data: { token: string; user: User }) {
      this.token = data.token
      this.user = data.user
      setToken(data.token)
      localStorage.setItem('ticket_user', JSON.stringify(data.user))
    },
    async login(nickname: string, password: string) {
      const data = await authApi.login(nickname, password)
      this.applyLogin(data)
    },
    async register(nickname: string, password: string, captchaId: string, captcha: string) {
      await authApi.register(nickname, password, captchaId, captcha)
    },
    async resetPassword(nickname: string, password: string, captchaId: string, captcha: string) {
      await authApi.resetPassword(nickname, password, captchaId, captcha)
    },
    async wechatLogin(code: string) {
      const data = await wechatApi.login(code)
      this.applyLogin(data)
    },
    async fetchMe() {
      if (!this.token) return null
      const user = await authApi.me()
      this.user = user
      localStorage.setItem('ticket_user', JSON.stringify(user))
      return user
    },
    logout() {
      this.token = ''
      this.user = null
      clearAuthStorage()
    }
  }
})
