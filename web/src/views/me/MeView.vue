<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'

import AppHeader from '@/components/layout/AppHeader.vue'
import { authApi, uploadApi } from '@/api'
import { useAuthStore } from '@/stores/auth'
import { toast } from '@/composables/toast'
import { timeFmt } from '@/utils/format'

const auth = useAuthStore()
const router = useRouter()

const nickname = ref(auth.user?.nickname || '')
const avatarInput = ref<HTMLInputElement | null>(null)
const uploadingAvatar = ref(false)
const saving = ref(false)

const initial = (auth.user?.nickname || '用').slice(0, 1)

async function pickAvatar() {
  const file = avatarInput.value?.files?.[0]
  if (!file) return
  if (!file.type.startsWith('image/')) {
    toast('请选择图片文件', 'error')
    return
  }
  uploadingAvatar.value = true
  try {
    const res = await uploadApi.upload(file, file.name, 'avatar')
    await authApi.updateProfile({ avatar: res.url })
    auth.user = { ...auth.user!, avatar: res.url }
    localStorage.setItem('ticket_user', JSON.stringify(auth.user))
    toast('头像已更新', 'success')
  } catch (e) {
    toast(e instanceof Error ? e.message : '头像上传失败', 'error')
  } finally {
    uploadingAvatar.value = false
    if (avatarInput.value) avatarInput.value.value = ''
  }
}

async function saveNickname() {
  const n = nickname.value.trim()
  if (n.length < 2) {
    toast('昵称至少 2 个字符', 'error')
    return
  }
  saving.value = true
  try {
    const user = await authApi.updateProfile({ nickname: n })
    auth.user = user
    localStorage.setItem('ticket_user', JSON.stringify(user))
    toast('昵称已更新', 'success')
  } catch (e) {
    toast(e instanceof Error ? e.message : '保存失败', 'error')
  } finally {
    saving.value = false
  }
}

function logout() {
  auth.logout()
  router.push('/')
}
</script>

<template>
  <div class="page">
    <AppHeader />
    <main class="page-inner fade-up" style="padding-top: 34px; max-width: 860px">
      <div class="panel" style="padding: 30px">
        <div class="row" style="align-items: flex-start">
          <!-- 头像 -->
          <div style="text-align: center; width: 130px">
            <div style="position: relative; display: inline-block" @click="avatarInput?.click()">
              <img v-if="auth.user?.avatar" class="avatar" :src="auth.user.avatar" style="width: 96px; height: 96px; border: 2px solid var(--border-strong); cursor: pointer" alt="头像" />
              <span v-else class="avatar-ph" style="width: 96px; height: 96px; font-size: 40px; cursor: pointer">{{ initial }}</span>
              <span style="position:absolute;right:-2px;bottom:-2px;background:var(--gold-grad);border-radius:50%;width:30px;height:30px;display:grid;place-items:center;color:#1c1503;border:2px solid var(--bg-1)">📷</span>
            </div>
            <input ref="avatarInput" type="file" accept="image/*" hidden @change="pickAvatar" />
            <p class="small muted mt-1">{{ uploadingAvatar ? '上传中…' : '点击更换头像' }}</p>
          </div>

          <!-- 资料 -->
          <div class="grow" style="min-width: 240px">
            <h2 class="font-display mb-1" style="font-size: 24px">{{ auth.user?.nickname }}</h2>
            <div class="row wrap mb-2">
              <span class="badge gold">{{ auth.isAdmin ? '管理员' : '创作者' }}</span>
              <span class="badge">注册于 {{ timeFmt(auth.user?.createTime, false) }}</span>
              <span v-if="auth.user?.openid" class="badge ok">已绑定公众号</span>
            </div>

            <div class="field mb-2" style="max-width: 380px">
              <label class="field-label">昵称</label>
              <div class="row">
                <input v-model="nickname" class="input" maxlength="24" />
                <button class="btn btn-ghost" :disabled="saving" @click="saveNickname">保存</button>
              </div>
            </div>
          </div>
        </div>

        <div class="row wrap" style="margin-top: 22px; border-top: 1px dashed var(--border); padding-top: 20px">
          <button class="btn btn-primary" @click="router.push('/products')">🎞️ 我的成品</button>
          <button class="btn btn-ghost" @click="router.push('/')">✏️ 去创作票根</button>
          <button v-if="auth.isAdmin" class="btn btn-ghost" @click="router.push('/admin/templates')">🖼️ 模板管理</button>
          <button class="btn btn-danger" @click="logout">退出登录</button>
        </div>
      </div>
    </main>
  </div>
</template>
