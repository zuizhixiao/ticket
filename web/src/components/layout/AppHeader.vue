<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const router = useRouter()
const route = useRoute()

const dropdownOpen = ref(false)
const menuOpen = ref(false)

function toggleDropdown(e: Event) {
  e.stopPropagation()
  dropdownOpen.value = !dropdownOpen.value
  menuOpen.value = false
}

function toggleMenu(e: Event) {
  e.stopPropagation()
  menuOpen.value = !menuOpen.value
  dropdownOpen.value = false
}

function closeAll() {
  dropdownOpen.value = false
  menuOpen.value = false
}

function logout() {
  closeAll()
  auth.logout()
  router.push('/')
}

function onDocClick() {
  closeAll()
}

// 路由变化时收起菜单
watch(
  () => route.fullPath,
  () => closeAll()
)

onMounted(() => document.addEventListener('click', onDocClick))
onBeforeUnmount(() => document.removeEventListener('click', onDocClick))

const initial = (auth.user?.nickname || '用').slice(0, 1)
</script>

<template>
  <header class="header">
    <div class="header-inner">
      <RouterLink to="/" class="brand" @click="closeAll">
        <span class="brand-mark">🎬</span>
        <span class="brand-name">
          电影纪念票根
          <span class="brand-sub">Cinema Ticket</span>
        </span>
      </RouterLink>

      <!-- 桌面横排 / 小屏下拉菜单(登录入口走右上角独立按钮,不在此重复) -->
      <nav class="nav" :class="{ open: menuOpen }">
        <RouterLink v-if="auth.isAuthed" class="nav-link" to="/products" @click="closeAll">
          我的成品
        </RouterLink>
        <RouterLink v-if="auth.isAdmin" class="nav-link" to="/admin/templates" @click="closeAll">
          模板管理
        </RouterLink>
      </nav>

      <div class="header-right">
        <!-- 汉堡(仅小屏且已登录时显示) -->
        <button
          v-if="auth.isAuthed"
          class="nav-toggle"
          :class="{ open: menuOpen }"
          aria-label="菜单"
          @click="toggleMenu"
        >
          <span></span>
          <span></span>
          <span></span>
        </button>

        <div class="user-zone">
          <template v-if="auth.isAuthed">
            <button class="user-chip" @click="toggleDropdown">
              <img v-if="auth.user?.avatar" class="avatar" :src="auth.user.avatar" alt="头像" />
              <span v-else class="avatar-ph">{{ initial }}</span>
              <span class="chip-name">{{ auth.user?.nickname }}</span>
              <span class="chip-arrow" style="color: var(--text-muted)">▾</span>
            </button>
            <div v-if="dropdownOpen" class="dropdown">
              <RouterLink class="dropdown-item" to="/me" @click="closeAll">个人中心</RouterLink>
              <RouterLink class="dropdown-item" to="/products" @click="closeAll">我的成品</RouterLink>
              <RouterLink
                v-if="auth.isAdmin"
                class="dropdown-item"
                to="/admin/templates"
                @click="closeAll"
              >
                模板管理
              </RouterLink>
              <button class="dropdown-item danger" @click="logout">退出登录</button>
            </div>
          </template>
          <RouterLink v-else class="btn btn-primary btn-sm login-btn" to="/login">登录</RouterLink>
        </div>
      </div>
    </div>
  </header>
</template>
