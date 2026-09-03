<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import AppHeader from '@/components/layout/AppHeader.vue'
import { authApi } from '@/api'
import { useAuthStore } from '@/stores/auth'
import { toast } from '@/composables/toast'
import type { ApiError } from '@/api/http'

const props = defineProps<{ mode?: string }>()
const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const currentMode = computed(() => {
  const m = props.mode
  if (m === 'register' || m === 'forgot' || m === 'login') return m
  return 'login'
})

const way = ref<'account' | 'wechat'>('account')
const busy = ref(false)

// 公众号验证码登录入口:暂时隐藏,上线时改为 true 即可恢复
const showWechat = false

// 表单
const nickname = ref('')
const password = ref('')
const confirm = ref('')
const captchaCode = ref('')
const captchaId = ref('')
const captchaImg = ref('')
const wechatCode = ref('')

function switchMode(m: string) {
  router.push(m === 'login' ? '/login' : m === 'register' ? '/register' : '/forgot')
}

async function refreshCaptcha() {
  try {
    const data = await authApi.captcha()
    captchaId.value = data.captchaId
    captchaImg.value = data.captchaImg
    captchaCode.value = ''
  } catch {
    toast('验证码加载失败', 'error')
  }
}

function errMsg(e: unknown): string {
  return e instanceof Error ? e.message : '操作失败'
}

async function doLogin() {
  if (!nickname.value.trim() || !password.value) {
    toast('请输入昵称和密码', 'error')
    return
  }
  busy.value = true
  try {
    await auth.login(nickname.value.trim(), password.value)
    const redirect = (route.query.redirect as string) || '/'
    router.push(redirect)
  } catch (e) {
    toast(errMsg(e), 'error')
  } finally {
    busy.value = false
  }
}

async function doRegister() {
  if (nickname.value.trim().length < 2) {
    toast('昵称至少 2 个字符', 'error')
    return
  }
  if (password.value.length < 6) {
    toast('密码至少 6 位', 'error')
    return
  }
  if (password.value !== confirm.value) {
    toast('两次输入的密码不一致', 'error')
    return
  }
  if (!captchaCode.value.trim()) {
    toast('请输入图形验证码', 'error')
    return
  }
  busy.value = true
  try {
    await auth.register(nickname.value.trim(), password.value, captchaId.value, captchaCode.value.trim())
    toast('注册成功,请登录', 'success')
    router.push('/login')
  } catch (e) {
    toast(errMsg(e), 'error')
    refreshCaptcha()
  } finally {
    busy.value = false
  }
}

async function doForgot() {
  if (!nickname.value.trim() || password.value.length < 6) {
    toast('请填写昵称与新密码(至少 6 位)', 'error')
    return
  }
  if (password.value !== confirm.value) {
    toast('两次输入的密码不一致', 'error')
    return
  }
  if (!captchaCode.value.trim()) {
    toast('请输入图形验证码', 'error')
    return
  }
  busy.value = true
  try {
    await auth.resetPassword(nickname.value.trim(), password.value, captchaId.value, captchaCode.value.trim())
    toast('密码已重置,请用新密码登录', 'success')
    router.push('/login')
  } catch (e) {
    toast(errMsg(e), 'error')
    refreshCaptcha()
  } finally {
    busy.value = false
  }
}

async function doWechat() {
  if (!/^\d{6}$/.test(wechatCode.value.trim())) {
    toast('请输入公众号下发的 6 位验证码', 'error')
    return
  }
  busy.value = true
  try {
    await auth.wechatLogin(wechatCode.value.trim())
    toast('微信登录成功', 'success')
    const redirect = (route.query.redirect as string) || '/'
    router.push(redirect)
  } catch (e) {
    toast(errMsg(e), 'error')
  } finally {
    busy.value = false
  }
}

watch(currentMode, () => {
  password.value = ''
  confirm.value = ''
  if (currentMode.value !== 'login') {
    way.value = 'account'
    refreshCaptcha()
  }
})

onMounted(() => {
  if (currentMode.value !== 'login') refreshCaptcha()
})

const titles: Record<string, string> = {
  login: '欢迎回来',
  register: '创建账号',
  forgot: '找回密码'
}
</script>

<template>
  <div class="page">
    <AppHeader />

    <div class="auth-wrap">
      <div class="auth-card fade-up">
        <div class="auth-brand">
          <div class="auth-logo">🎬</div>
          <h1 class="auth-title text-gold-grad font-display">电影纪念票根</h1>
          <p class="auth-sub">CINEMA MEMORY TICKET</p>
        </div>

        <div class="auth-body">
          <div class="tab-row">
            <button class="tab-btn" :class="{ active: currentMode === 'login' }" @click="switchMode('login')">登录</button>
            <button class="tab-btn" :class="{ active: currentMode === 'register' }" @click="switchMode('register')">注册</button>
            <button class="tab-btn" :class="{ active: currentMode === 'forgot' }" @click="switchMode('forgot')">找回</button>
          </div>

          <template v-if="currentMode === 'login'">
            <!-- 账号密码登录(默认) -->
            <form class="row-between" style="flex-direction:column;gap:16px;align-items:stretch" @submit.prevent="doLogin">
              <div class="field">
                <label class="field-label">昵称</label>
                <input v-model="nickname" class="input" placeholder="你的昵称" autocomplete="username" />
              </div>
              <div class="field">
                <label class="field-label">密码</label>
                <input v-model="password" type="password" class="input" placeholder="登录密码" autocomplete="current-password" />
              </div>
              <button class="btn btn-primary btn-block" :disabled="busy">
                <span v-if="busy" class="spinner"></span>登 录
              </button>
            </form>

            <!-- 公众号验证码登录:暂时隐藏(showWechat=false),需要时改为 true -->
            <template v-if="showWechat">
              <div class="row" style="justify-content:center;gap:10px;margin:16px 0 4px">
                <span style="flex:1;height:1px;background:var(--border)"></span>
                <span class="small muted">或使用公众号登录</span>
                <span style="flex:1;height:1px;background:var(--border)"></span>
              </div>
              <div class="panel" style="padding:14px;background:rgba(212,175,55,.06);margin-bottom:14px">
                <p class="small" style="line-height:1.9">
                  ① 关注公众号并发送关键词 <b class="text-gold-grad">验证码</b><br />
                  ② 将收到的 6 位数字填入下方
                </p>
              </div>
              <div class="field mb-2">
                <label class="field-label">6 位验证码</label>
                <input v-model="wechatCode" class="input" maxlength="6" inputmode="numeric" placeholder="例如:483920" />
              </div>
              <button class="btn btn-ghost btn-block" :disabled="busy" @click="doWechat">
                <span v-if="busy" class="spinner"></span>微信登录
              </button>
            </template>
          </template>

          <!-- 注册 / 找回:昵称 + 密码 + 确认 + 验证码 -->
          <form
            v-else
            class="row-between"
            style="flex-direction:column;gap:16px;align-items:stretch"
            @submit.prevent="currentMode === 'register' ? doRegister() : doForgot()"
          >
            <div class="field">
              <label class="field-label">昵称</label>
              <input v-model="nickname" class="input" placeholder="2-24 个字符" autocomplete="username" />
            </div>
            <div class="field">
              <label class="field-label">{{ currentMode === 'register' ? '设置密码' : '新密码' }}</label>
              <input v-model="password" type="password" class="input" placeholder="至少 6 位" autocomplete="new-password" />
            </div>
            <div class="field">
              <label class="field-label">确认密码</label>
              <input v-model="confirm" type="password" class="input" placeholder="再次输入密码" autocomplete="new-password" />
            </div>
            <div class="field">
              <label class="field-label">图形验证码</label>
              <div class="captcha-row">
                <input v-model="captchaCode" class="input" maxlength="4" placeholder="输入右侧字符" />
                <img
                  :src="captchaImg"
                  class="captcha-img"
                  alt="验证码"
                  title="点击刷新"
                  @click="refreshCaptcha"
                />
              </div>
            </div>
            <button class="btn btn-primary btn-block" :disabled="busy">
              <span v-if="busy" class="spinner"></span>
              {{ currentMode === 'register' ? '注册并登录' : '重置密码' }}
            </button>
          </form>
        </div>

        <div class="auth-footer muted">
          {{ titles[currentMode] }}即表示同意本站的
          <span class="text-gold-grad">服务条款</span>
          <div v-if="currentMode === 'login'" class="mt-1">
            <RouterLink to="/admin/login" class="small" style="opacity: 0.85">🔐 管理员入口 →</RouterLink>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
