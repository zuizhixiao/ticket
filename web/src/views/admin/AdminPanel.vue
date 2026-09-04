<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'

import AppHeader from '@/components/layout/AppHeader.vue'
import Modal from '@/components/ui/Modal.vue'
import { adminApi } from '@/api'
import { toast } from '@/composables/toast'
import type { AdminImageRow } from '@/types'
import { timeFmt } from '@/utils/format'

const route = useRoute()
const SIZE = 20

type ListKind = 'product' | 'poster'

// ---------- 数据(成品/海报各自缓存) ----------
const products = ref<AdminImageRow[]>([])
const productTotal = ref(0)
const productLoading = ref(false)
const productPage = ref(1)

const posters = ref<AdminImageRow[]>([])
const posterTotal = ref(0)
const posterLoading = ref(false)
const posterPage = ref(1)

// 当前展示的列表类型:由顶部导航 /admin?view=product|poster 控制
const active = ref<ListKind>((route.query.view as string) === 'poster' ? 'poster' : 'product')

async function loadList(kind: ListKind, page: number, reset: boolean) {
  const isP = kind === 'product'
  const loading = isP ? productLoading : posterLoading
  loading.value = true
  try {
    const data = await adminApi.images(kind, page, SIZE)
    const rows = isP ? products : posters
    const total = isP ? productTotal : posterTotal
    const pageRef = isP ? productPage : posterPage
    if (reset) {
      rows.value = data.list
      pageRef.value = 1
    } else {
      rows.value = [...rows.value, ...data.list]
      pageRef.value = page
    }
    total.value = data.total
  } catch (e) {
    toast(e instanceof Error ? e.message : '加载失败', 'error')
  } finally {
    loading.value = false
  }
}

function hasMore(kind: ListKind): boolean {
  const page = kind === 'product' ? productPage.value : posterPage.value
  const total = kind === 'product' ? productTotal.value : posterTotal.value
  return page * SIZE < total
}

async function ensureView(kind: ListKind) {
  active.value = kind
  const rows = kind === 'product' ? products.value : posters.value
  if (!rows.length && !(kind === 'product' ? productLoading.value : posterLoading.value)) {
    await loadList(kind, 1, true)
  }
}

// 顶部导航切换 /admin?view=… 时,自动切换展示的列表
watch(
  () => route.query.view,
  (v) => {
    if (v === 'product' || v === 'poster') ensureView(v)
  }
)

// 当前面板(展开为普通值,模板直接用)
const panel = computed(() => {
  const isP = active.value === 'product'
  return {
    kind: active.value,
    icon: isP ? '🎞️' : '🖼️',
    title: isP ? '全部成品' : '全部海报',
    rows: isP ? products.value : posters.value,
    total: isP ? productTotal.value : posterTotal.value,
    loading: isP ? productLoading.value : posterLoading.value,
    emptyText: isP ? '暂无成品' : '暂无海报',
    hasMore: hasMore(active.value),
    onLoadMore: () => loadList(active.value, (isP ? productPage.value : posterPage.value) + 1, false)
  }
})

// ---------- 删除 / 预览 ----------
const deletingId = ref(0)
const lightbox = ref<AdminImageRow | null>(null)

async function removeItem(item: AdminImageRow) {
  const label = item.type === 'product' ? '成品' : '海报'
  if (!window.confirm(`确定删除这张${label}吗?删除后不可恢复。`)) return
  deletingId.value = item.id
  try {
    await adminApi.removeImage(item.id)
    if (active.value === 'product') {
      products.value = products.value.filter((x) => x.id !== item.id)
      productTotal.value -= 1
    } else {
      posters.value = posters.value.filter((x) => x.id !== item.id)
      posterTotal.value -= 1
    }
    if (lightbox.value?.id === item.id) lightbox.value = null
    toast('已删除', 'success')
  } catch (e) {
    toast(e instanceof Error ? e.message : '删除失败', 'error')
  } finally {
    deletingId.value = 0
  }
}

function openLightbox(item: AdminImageRow) {
  lightbox.value = item
}

onMounted(() => {
  loadList('product', 1, true)
  loadList('poster', 1, true)
  if ((route.query.view as string) === 'poster') active.value = 'poster'
})
</script>

<template>
  <div class="page">
    <AppHeader />
    <main class="page-inner fade-up" style="padding-top: 28px; padding-bottom: 60px">
      <!-- 当前列表(管理后台标题与「更多」菜单在顶部导航栏) -->
      <section class="panel" style="padding: 18px">
        <div class="row-between mb-2 wrap">
          <h3 class="panel-title" style="margin: 0">
            {{ panel.icon }} {{ panel.title }}
          </h3>
          <span class="badge gold">共 {{ panel.total }} 张</span>
        </div>

        <div v-if="!panel.rows.length && !panel.loading" class="empty" style="padding: 34px 12px">
          <div class="empty-icon">{{ panel.icon }}</div>
          <p>{{ panel.emptyText }}</p>
        </div>

        <div v-else class="admin-imgs">
          <div
            v-for="item in panel.rows"
            :key="item.id"
            class="admin-img-card"
            @click="openLightbox(item)"
          >
            <img :src="item.thumbUrl || item.url" :alt="item.filename" loading="lazy" />
            <button
              class="admin-rm"
              title="删除"
              :disabled="deletingId === item.id"
              @click.stop="removeItem(item)"
            >
              ✕
            </button>
          </div>
        </div>

        <!-- 手动加载更多 -->
        <div v-if="panel.hasMore" class="row" style="justify-content: center; margin-top: 14px">
          <button class="btn btn-ghost btn-sm" :disabled="panel.loading" @click="panel.onLoadMore">
            {{ panel.loading ? '加载中…' : '加载更多' }}
          </button>
        </div>
        <div v-else-if="panel.rows.length" class="row" style="justify-content: center; margin-top: 10px">
          <span class="small muted">已全部加载</span>
        </div>
      </section>
    </main>

    <!-- 大图预览 -->
    <Modal v-if="lightbox" :show="!!lightbox" :title="lightbox?.filename || '图片'" @close="lightbox = null">
      <img v-if="lightbox" :src="lightbox.url" class="lightbox-img" alt="预览" style="max-height: 62vh" />
      <div class="row mt-2">
        <span class="small muted grow">
          上传者:{{ lightbox?.nickname || '游客' }} · {{ timeFmt(lightbox?.createTime) }}
        </span>
        <button class="btn btn-ghost" @click="lightbox = null">关闭</button>
        <a v-if="lightbox" class="btn btn-primary" :href="lightbox.url" target="_blank" rel="noopener">
          新窗口打开
        </a>
      </div>
    </Modal>
  </div>
</template>
