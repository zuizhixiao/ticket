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

// 管理员「更多」:成品列表 / 海报列表 / 用户列表 / 模板管理 / 退出
function adminMore(kind: 'product' | 'poster' | 'users' | 'templates') {
  dropdownOpen.value = false
  if (kind === 'templates') router.push('/admin/templates')
  else if (kind === 'users') router.push('/admin/users')
  else router.push({ path: '/admin', query: { view: kind } })
}

function logout() {
  const wasAdmin = auth.isAdmin
  closeAll()
  auth.logout()
  // 管理员退出回管理员登录页;普通用户回首页
  router.push(wasAdmin ? '/admin/login' : '/')
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
      <RouterLink
        :to="auth.isAdmin ? { path: '/admin', query: { view: 'product' } } : '/'"
        class="brand"
        @click="closeAll"
      >
        <span class="brand-mark">{{ auth.isAdmin ? '🛠️' : '🎬' }}</span>
        <span class="brand-name">
          {{ auth.isAdmin ? '管理后台' : '电影纪念票根' }}
          <span class="brand-sub">{{ auth.isAdmin ? 'Admin Console' : 'Cinema Ticket' }}</span>
        </span>
      </RouterLink>

      <!-- 左侧直达链接(普通用户) -->
      <nav v-if="auth.isAuthed && !auth.isAdmin" class="quick-nav">
        <RouterLink class="nav-link" to="/products" @click="closeAll">我的成品</RouterLink>
      </nav>

      <!-- ============ 管理员:右上角三横线 = 「更多」菜单 ============ -->
      <div v-if="auth.isAdmin" class="header-right">
        <div class="user-zone">
          <button
            class="nav-toggle"
            :class="{ open: dropdownOpen }"
            aria-label="更多菜单"
            @click="toggleDropdown"
          >
            <span></span>
            <span></span>
            <span></span>
          </button>
          <div v-if="dropdownOpen" class="dropdown">
            <button class="dropdown-item" @click="adminMore('product')">🎞️ 成品列表</button>
            <button class="dropdown-item" @click="adminMore('poster')">🖼️ 海报列表</button>
            <button class="dropdown-item" @click="adminMore('users')">👥 用户列表</button>
            <button class="dropdown-item" @click="adminMore('templates')">🛠️ 模板管理</button>
            <div style="height: 1px; background: var(--border); margin: 6px 10px"></div>
            <button class="dropdown-item danger" @click="logout">退出登录</button>
          </div>
        </div>
      </div>

      <!-- ============ 普通登录用户 ============ -->
      <div v-else class="header-right">
        <button
          v-if="auth.isAuthed"
          class="nav-toggle"
          :class="{ open: menuOpen }"
          aria-label="更多菜单"
          @click="toggleMenu"
        >
          <span></span>
          <span></span>
          <span></span>
        </button>

        <nav v-if="menuOpen" class="menu-pop" @click.stop>
          <RouterLink class="dropdown-item" to="/products" @click="closeAll">🎞️ 我的成品</RouterLink>
          <RouterLink class="dropdown-item" to="/me" @click="closeAll">👤 个人中心</RouterLink>
          <div style="height: 1px; background: var(--border); margin: 6px 10px"></div>
          <button class="dropdown-item danger" @click="logout">退出登录</button>
        </nav>

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
              <button class="dropdown-item danger" @click="logout">退出登录</button>
            </div>
          </template>
          <RouterLink v-else class="btn btn-primary btn-sm login-btn" to="/login">登录</RouterLink>
        </div>
      </div>
    </div>
  </header>
</template>
