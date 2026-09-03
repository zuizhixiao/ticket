import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(),
  scrollBehavior: () => ({ top: 0 }),
  routes: [
    {
      path: '/',
      name: 'editor',
      component: () => import('@/views/editor/EditorView.vue'),
      meta: { title: '创作票根' }
    },
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/auth/AuthView.vue'),
      props: { mode: 'login' },
      meta: { title: '登录' }
    },
    {
      path: '/register',
      name: 'register',
      component: () => import('@/views/auth/AuthView.vue'),
      props: { mode: 'register' },
      meta: { title: '注册' }
    },
    {
      path: '/forgot',
      name: 'forgot',
      component: () => import('@/views/auth/AuthView.vue'),
      props: { mode: 'forgot' },
      meta: { title: '找回密码' }
    },
    {
      path: '/me',
      name: 'me',
      component: () => import('@/views/me/MeView.vue'),
      meta: { title: '个人中心', auth: true }
    },
    {
      path: '/products',
      name: 'products',
      component: () => import('@/views/products/ProductsView.vue'),
      meta: { title: '我的成品', auth: true }
    },
    {
      path: '/admin/templates',
      name: 'admin-templates',
      component: () => import('@/views/admin/TemplatesAdmin.vue'),
      meta: { title: '模板管理', auth: true, admin: true }
    },
    { path: '/:pathMatch(.*)*', redirect: '/' }
  ]
})

router.beforeEach(async (to) => {
  const auth = useAuthStore()

  if (to.meta.auth && !auth.token) {
    return { path: '/login', query: { redirect: to.fullPath } }
  }

  // 有 token 但本地无用户信息时拉取一次
  if (auth.token && !auth.user) {
    try {
      await auth.fetchMe()
    } catch {
      /* 已由 http 层处理 401 跳转 */
    }
  }

  if (to.meta.admin && !auth.isAdmin) {
    return { path: '/' }
  }
  return true
})

router.afterEach((to) => {
  const base = '电影纪念票根'
  document.title = to.meta.title ? `${to.meta.title as string} · ${base}` : base
})

export default router
