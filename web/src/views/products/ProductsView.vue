<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'

import AppHeader from '@/components/layout/AppHeader.vue'
import Modal from '@/components/ui/Modal.vue'
import { productApi } from '@/api'
import { toast } from '@/composables/toast'
import type { ProductImage } from '@/types'
import { timeFmt } from '@/utils/format'
import { downloadBlob } from '@/lib/ticket/media'

const router = useRouter()

const list = ref<ProductImage[]>([])
const total = ref(0)
const page = ref(1)
const size = 12
const loading = ref(false)
const removingId = ref(0)

const lightbox = ref<ProductImage | null>(null)
const downloading = ref(false)

async function load(p = page.value) {
  loading.value = true
  try {
    const data = await productApi.list(p, size)
    list.value = data.list
    total.value = data.total
    page.value = p
  } catch (e) {
    toast(e instanceof Error ? e.message : '加载失败', 'error')
  } finally {
    loading.value = false
  }
}

const totalPages = ref(1)

async function removeItem(item: ProductImage) {
  if (!window.confirm('确定删除这张票根吗?删除后不可恢复。')) return
  removingId.value = item.id
  try {
    await productApi.remove(item.id)
    toast('已删除', 'success')
    if (list.value.length === 1 && page.value > 1) page.value -= 1
    await load(page.value)
  } catch (e) {
    toast(e instanceof Error ? e.message : '删除失败', 'error')
  } finally {
    removingId.value = 0
  }
}

async function downloadRemote(item: ProductImage) {
  downloading.value = true
  try {
    const resp = await fetch(item.url)
    const blob = await resp.blob()
    downloadBlob(blob, item.filename || `ticket_${item.id}.png`)
  } catch {
    window.open(item.url, '_blank')
  } finally {
    downloading.value = false
  }
}

onMounted(async () => {
  await load(1)
  totalPages.value = Math.max(1, Math.ceil(total.value / size))
})
</script>

<template>
  <div class="page">
    <AppHeader />
    <main class="page-inner fade-up" style="padding-top: 30px; padding-bottom: 60px">
      <div class="row-between wrap mb-2">
        <h2 class="font-display" style="font-size: 26px">🎞️ 我的成品</h2>
        <button class="btn btn-primary btn-sm" @click="router.push('/')">＋ 创作新票根</button>
      </div>

      <div v-if="loading && !list.length" class="gallery">
        <div v-for="n in 6" :key="n" class="skel" style="aspect-ratio: 700/1400"></div>
      </div>

      <div v-else-if="!list.length" class="panel">
        <div class="empty">
          <div class="empty-icon">🎬</div>
          <h3>还没有成品</h3>
          <p>去创作第一张属于你的电影纪念票根吧</p>
          <button class="btn btn-primary" @click="router.push('/')">开始创作</button>
        </div>
      </div>

      <div v-else class="gallery">
        <div v-for="item in list" :key="item.id" class="prod-card" @click="lightbox = item">
          <div class="prod-img">
            <img :src="item.url" :alt="item.filename" loading="lazy" />
          </div>
          <div class="prod-meta">
            <span class="prod-name" :title="item.filename">{{ item.filename || '票根' }}</span>
            <span class="prod-date">{{ timeFmt(item.createTime, false) }}</span>
          </div>
          <div class="prod-actions" @click.stop>
            <button class="btn btn-ghost btn-sm grow" @click="downloadRemote(item)">下载</button>
            <button class="btn btn-danger btn-sm" :disabled="removingId === item.id" @click="removeItem(item)">
              {{ removingId === item.id ? '…' : '删除' }}
            </button>
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
    </main>

    <Modal v-if="lightbox" :show="!!lightbox" :title="lightbox?.filename || '票根'" @close="lightbox = null">
      <img v-if="lightbox" :src="lightbox.url" class="lightbox-img" alt="票根" />
      <div class="row mt-2">
        <button class="btn btn-ghost" @click="lightbox = null">关闭</button>
        <button v-if="lightbox" class="btn btn-primary grow" :disabled="downloading" @click="downloadRemote(lightbox)">
          <span v-if="downloading" class="spinner"></span>💾 下载原图
        </button>
      </div>
    </Modal>
  </div>
</template>
