<script setup lang="ts">
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { useAuthStore } from '@/stores/auth'
import { toast } from '@/composables/toast'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const nickname = ref('')
const password = ref('')
const busy = ref(false)

async function doLogin() {
  if (!nickname.value.trim() || !password.value) {
    toast('请输入账号与密码', 'error')
    return
  }
  busy.value = true
  try {
    await auth.login(nickname.value.trim(), password.value)
    if (auth.user?.role !== 1) {
      // 普通账号不允许进入管理后台:清掉登录态并提示
      auth.logout()
      toast('该账号不是管理员,无法登录管理后台', 'error')
      return
    }
    toast('管理员登录成功', 'success')
    const redirect = (route.query.redirect as string) || '/admin'
    router.push(redirect)
  } catch (e) {
    toast(e instanceof Error ? e.message : '登录失败', 'error')
  } finally {
    busy.value = false
  }
}
</script>

<template>
  <div class="page" style="min-height: 100vh">
    <div class="auth-wrap">
      <div class="auth-card fade-up">
        <div class="auth-brand">
          <div class="auth-logo">🛠️</div>
          <h1 class="auth-title text-gold-grad font-display">管理后台</h1>
          <p class="auth-sub">ADMIN CONSOLE · 仅限管理员</p>
        </div>

        <div class="auth-body">
          <form
            class="row-between"
            style="flex-direction: column; gap: 16px; align-items: stretch"
            @submit.prevent="doLogin"
          >
            <div class="field">
              <label class="field-label">管理员账号</label>
              <input
                v-model="nickname"
                class="input"
                placeholder="请输入管理员昵称"
                autocomplete="username"
              />
            </div>
            <div class="field">
              <label class="field-label">密码</label>
              <input
                v-model="password"
                type="password"
                class="input"
                placeholder="管理员密码"
                autocomplete="current-password"
              />
            </div>
            <button class="btn btn-primary btn-block" :disabled="busy">
              <span v-if="busy" class="spinner"></span>登 录 管 理 后 台
            </button>
          </form>
        </div>
      </div>
    </div>
  </div>
</template>
