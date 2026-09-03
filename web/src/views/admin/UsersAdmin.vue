<script setup lang="ts">
import { onMounted, ref } from 'vue'

import AppHeader from '@/components/layout/AppHeader.vue'
import Modal from '@/components/ui/Modal.vue'
import { adminApi } from '@/api'
import { toast } from '@/composables/toast'
import { useAuthStore } from '@/stores/auth'
import type { AdminUser } from '@/types'
import { timeFmt } from '@/utils/format'

const auth = useAuthStore()
const SIZE = 20

const rows = ref<AdminUser[]>([])
const total = ref(0)
const page = ref(1)
const keyword = ref('')
const loading = ref(false)
const busyId = ref(0)

const resetTarget = ref<AdminUser | null>(null)
const newPassword = ref('')
const resetting = ref(false)

async function load(p = page.value, kw = keyword.value) {
  loading.value = true
  try {
    const data = await adminApi.users(kw.trim(), p, SIZE)
    rows.value = data.list
    total.value = data.total
    totalPages.value = Math.max(1, Math.ceil(data.total / SIZE))
    page.value = p
  } catch (e) {
    toast(e instanceof Error ? e.message : '加载失败', 'error')
  } finally {
    loading.value = false
  }
}

function search() {
  load(1, keyword.value)
}

async function toggleFreeze(u: AdminUser) {
  const next = u.status === 1 ? 0 : 1
  const act = next === 0 ? '冻结' : '恢复'
  if (!window.confirm(`确定${act}账号「${u.nickname}」吗?`)) return
  busyId.value = u.id
  try {
    await adminApi.userStatus(u.id, next)
    u.status = next
    toast(next === 0 ? '已冻结' : '已恢复', 'success')
  } catch (e) {
    toast(e instanceof Error ? e.message : '操作失败', 'error')
  } finally {
    busyId.value = 0
  }
}

function openReset(u: AdminUser) {
  resetTarget.value = u
  newPassword.value = ''
}

async function doReset() {
  if (!resetTarget.value) return
  if (newPassword.value.length < 6 || newPassword.value.length > 64) {
    toast('新密码需为 6-64 位', 'error')
    return
  }
  resetting.value = true
  try {
    await adminApi.userPassword(resetTarget.value.id, newPassword.value)
    toast(`已重置「${resetTarget.value.nickname}」的密码`, 'success')
    resetTarget.value = null
  } catch (e) {
    toast(e instanceof Error ? e.message : '重置失败', 'error')
  } finally {
    resetting.value = false
  }
}

function initialOf(u: AdminUser) {
  return (u.nickname || '?').slice(0, 1)
}

const totalPages = ref(1)

onMounted(() => {
  load(1)
})
</script>

<template>
  <div class="page">
    <AppHeader />
    <main class="page-inner fade-up" style="padding-top: 28px; padding-bottom: 60px">
      <section class="panel" style="padding: 18px">
        <div class="row-between wrap mb-2">
          <h3 class="panel-title" style="margin: 0">👥 用户列表</h3>
          <span class="badge gold">共 {{ total }} 个账号</span>
        </div>

        <div class="row mb-2 wrap">
          <input
            v-model="keyword"
            class="input"
            style="max-width: 260px"
            placeholder="按昵称搜索"
            @keyup.enter="search"
          />
          <button class="btn btn-ghost btn-sm" @click="search">搜索</button>
        </div>

        <div v-if="loading && !rows.length" class="empty" style="padding: 30px 12px">
          <div class="empty-icon">👥</div>
          <p>加载中…</p>
        </div>

        <div v-else-if="!rows.length" class="empty" style="padding: 30px 12px">
          <div class="empty-icon">👥</div>
          <p>暂无用户</p>
        </div>

        <div v-else class="row-between" style="flex-direction: column; align-items: stretch; gap: 10px">
          <div
            v-for="u in rows"
            :key="u.id"
            class="panel"
            style="padding: 12px 14px; background: var(--surface)"
          >
            <div class="row-between wrap" style="gap: 10px">
              <div class="row" style="min-width: 0">
                <span class="avatar-ph" style="width: 38px; height: 38px; font-size: 16px">
                  {{ initialOf(u) }}
                </span>
                <div style="min-width: 0">
                  <div class="row wrap" style="gap: 8px">
                    <b>{{ u.nickname }}</b>
                    <span v-if="u.role === 1" class="badge gold">管理员</span>
                    <span v-if="u.status === 0" class="badge off">已冻结</span>
                  </div>
                  <div class="small muted">
                    ID {{ u.id }} · 注册 {{ timeFmt(u.createTime, false) }}
                    <template v-if="u.lastLoginTime"> · 最近登录 {{ timeFmt(u.lastLoginTime) }}</template>
                    <template v-if="u.openid"> · 已绑公众号</template>
                  </div>
                </div>
              </div>

              <div class="row wrap" style="flex-shrink: 0">
                <template v-if="u.role !== 1">
                  <button
                    class="btn btn-sm"
                    :class="u.status === 1 ? 'btn-danger' : 'btn-primary'"
                    :disabled="busyId === u.id || u.id === auth.user?.id"
                    @click="toggleFreeze(u)"
                  >
                    {{ u.status === 1 ? '冻结' : '恢复' }}
                  </button>
                  <button class="btn btn-ghost btn-sm" :disabled="u.id === auth.user?.id" @click="openReset(u)">
                    重置密码
                  </button>
                </template>
                <span v-else class="small muted">管理员账号受保护</span>
              </div>
            </div>
          </div>
        </div>

        <div v-if="totalPages > 1" class="pager">
          <button class="page-num" :disabled="page <= 1" @click="load(page - 1)">‹</button>
          <button
            v-for="p in totalPages"
            :key="p"
            class="page-num"
            :class="{ active: p === page }"
            @click="load(p)"
          >
            {{ p }}
          </button>
          <button class="page-num" :disabled="page >= totalPages" @click="load(page + 1)">›</button>
        </div>
      </section>
    </main>

    <!-- 重置密码 -->
    <Modal :show="!!resetTarget" :title="`重置密码 · ${resetTarget?.nickname || ''}`" @close="resetTarget = null">
      <div class="field">
        <label class="field-label">新密码(6-64 位)</label>
        <input v-model="newPassword" type="text" class="input" placeholder="输入新密码" autocomplete="new-password" />
      </div>
      <div class="row mt-2">
        <button class="btn btn-ghost" @click="resetTarget = null">取消</button>
        <button class="btn btn-primary grow" :disabled="resetting" @click="doReset">
          <span v-if="resetting" class="spinner"></span>确认重置
        </button>
      </div>
    </Modal>
  </div>
</template>
