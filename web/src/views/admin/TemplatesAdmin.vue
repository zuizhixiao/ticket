<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'

import AppHeader from '@/components/layout/AppHeader.vue'
import Modal from '@/components/ui/Modal.vue'
import { adminApi } from '@/api'
import { toast } from '@/composables/toast'
import type { Template } from '@/types'

const rows = ref<Template[]>([])
const loading = ref(false)
const filter = ref<'0' | '1' | '2'>('0')
const busyId = ref(0)

async function load() {
  loading.value = true
  try {
    const data = await adminApi.templates(Number(filter.value))
    rows.value = data.list
  } catch (e) {
    toast(e instanceof Error ? e.message : '加载失败', 'error')
  } finally {
    loading.value = false
  }
}

onMounted(load)

// ---------- 新增 / 编辑 ----------
const showEdit = ref(false)
const saving = ref(false)
const previewFailed = ref(false)
const draft = reactive<{ id: number | null; name: string; url: string; titleColor: string; textColor: string }>({
  id: null,
  name: '',
  url: '',
  titleColor: '#ffffff',
  textColor: '#f9d9c2'
})

function openCreate() {
  Object.assign(draft, { id: null, name: '', url: '', titleColor: '#ffffff', textColor: '#f9d9c2' })
  previewFailed.value = false
  showEdit.value = true
}

function openEdit(t: Template) {
  Object.assign(draft, {
    id: t.id,
    name: t.name || '',
    url: t.url,
    titleColor: t.titleColor || '#ffffff',
    textColor: t.textColor || '#f9d9c2'
  })
  previewFailed.value = false
  showEdit.value = true
}

async function saveDraft() {
  const url = draft.url.trim()
  if (!url) {
    toast('请填写图片地址', 'error')
    return
  }
  if (draft.name.trim() && draft.name.trim().length > 24) {
    toast('模板名称请勿超过 24 字', 'error')
    return
  }
  saving.value = true
  try {
    const payload = { name: draft.name.trim(), url, titleColor: draft.titleColor, textColor: draft.textColor }
    if (draft.id == null) {
      await adminApi.create(payload)
      toast('模板已创建', 'success')
    } else {
      await adminApi.update(draft.id, payload)
      toast('模板已保存', 'success')
    }
    showEdit.value = false
    await load()
  } catch (e) {
    toast(e instanceof Error ? e.message : '保存失败', 'error')
  } finally {
    saving.value = false
  }
}

async function toggleStatus(t: Template) {
  const next = t.status === 1 ? 2 : 1
  busyId.value = t.id
  try {
    await adminApi.update(t.id, { status: next })
    toast(next === 1 ? '已上架' : '已下架', 'success')
    await load()
  } catch (e) {
    toast(e instanceof Error ? e.message : '操作失败', 'error')
  } finally {
    busyId.value = 0
  }
}

async function removeItem(t: Template) {
  if (!window.confirm(`确定删除模板「${t.name || `#${t.id}`}」吗?删除后不可恢复。`)) return
  busyId.value = t.id
  try {
    await adminApi.remove(t.id)
    toast('已删除', 'success')
    await load()
  } catch (e) {
    toast(e instanceof Error ? e.message : '删除失败', 'error')
  } finally {
    busyId.value = 0
  }
}

// ---------- 拖拽排序 ----------
const dragId = ref<number | null>(null)

function onDragStart(t: Template) {
  dragId.value = t.id
}

function onDropAt(target: Template) {
  const from = dragId.value
  dragId.value = null
  if (from == null || from === target.id) return
  const arr = [...rows.value]
  const fi = arr.findIndex((x) => x.id === from)
  const ti = arr.findIndex((x) => x.id === target.id)
  if (fi < 0 || ti < 0) return
  const [moved] = arr.splice(fi, 1)
  arr.splice(ti, 0, moved)
  rows.value = arr
  persistOrder(arr)
}

async function persistOrder(arr: Template[]) {
  try {
    await adminApi.reorderSort(arr.map((x) => x.id))
    toast('排序已保存', 'success')
  } catch (e) {
    toast(e instanceof Error ? e.message : '排序保存失败', 'error')
    await load()
  }
}
</script>

<template>
  <div class="page">
    <AppHeader />
    <main class="page-inner fade-up" style="padding-top: 30px; padding-bottom: 60px">
      <div class="row-between mb-2 wrap">
        <h2 class="font-display" style="font-size: 26px">🖼️ 系统模板管理</h2>
        <div class="row">
          <select v-model="filter" class="select" style="width: 130px" @change="load">
            <option value="0">全部</option>
            <option value="1">上架中</option>
            <option value="2">已下架</option>
          </select>
          <button class="btn btn-primary btn-sm" @click="openCreate">＋ 新增模板</button>
        </div>
      </div>

      <div v-if="loading && !rows.length" class="row wrap" style="gap: 14px">
        <div v-for="n in 4" :key="n" class="skel" style="width: 200px; height: 400px"></div>
      </div>

      <div v-else-if="!rows.length" class="panel">
        <div class="empty">
          <div class="empty-icon">🖼️</div>
          <h3>暂无模板</h3>
          <button class="btn btn-primary" @click="openCreate">新增第一个模板</button>
        </div>
      </div>

      <div v-else class="gallery admin-grid">
        <div
          v-for="t in rows"
          :key="t.id"
          class="tpl-admin-card"
          :class="{ dragging: dragId === t.id }"
          draggable="true"
          @dragstart="onDragStart(t)"
          @dragover.prevent
          @drop="onDropAt(t)"
          @dragend="dragId = null"
          :title="t.name ? `${t.name} — 拖动排序` : '拖动排序'"
        >
          <!-- 完整展示的背景图 -->
          <div class="tpl-card-thumb">
            <img :src="t.url" :alt="t.name || `模板${t.id}`" loading="lazy" />
            <span class="tpl-dot" :class="t.status === 1 ? 'on' : 'off'" :title="t.status === 1 ? '上架中' : '已下架'"></span>
          </div>

          <!-- 图标操作(无文字) -->
          <div class="tpl-card-actions" draggable="false">
            <button class="tpl-icon" data-tip="编辑" @click="openEdit(t)">✎</button>
            <button
              class="tpl-icon"
              :data-tip="t.status === 1 ? '下架' : '上架'"
              :disabled="busyId === t.id"
              @click="toggleStatus(t)"
            >
              {{ t.status === 1 ? '▼' : '▲' }}
            </button>
            <button class="tpl-icon danger" data-tip="删除" :disabled="busyId === t.id" @click="removeItem(t)">
              🗑
            </button>
          </div>
        </div>
      </div>
    </main>

    <!-- 新增/编辑弹窗(填写图片地址) -->
    <Modal :show="showEdit" :title="draft.id == null ? '新增模板' : '编辑模板'" @close="showEdit = false">
      <div class="row-between wrap mb-2 form-cols" style="align-items: flex-start">
        <!-- 带示例文字的实时预览(调色方便) -->
        <div class="form-preview tpl-preview" style="width: 180px; height: 360px; flex-shrink: 0">
          <template v-if="draft.url && !previewFailed">
            <img :src="draft.url" alt="预览" @error="previewFailed = true" />
            <div class="pv-title" :style="{ color: draft.titleColor }">电影标题</div>
            <div class="pv-line" style="top: 76%" :style="{ color: draft.textColor }">国语 2D | 星光影城</div>
            <div class="pv-line" style="top: 80.5%" :style="{ color: draft.textColor }">3号厅 7排12座</div>
            <div class="pv-line" style="top: 85%" :style="{ color: draft.textColor }">2026-02-14 19:30</div>
          </template>
          <div v-else style="height: 100%; display: grid; place-items: center; color: var(--text-muted); font-size: 12px; text-align: center; padding: 8px">
            填写图片地址后<br />显示预览与文字
          </div>
        </div>

        <div class="grow" style="min-width: 250px">
          <div class="field mb-2">
            <label class="field-label">图片地址(必填)</label>
            <input v-model="draft.url" class="input" placeholder="https://…/template.png" @input="previewFailed = false" />
          </div>
          <div class="field mb-2">
            <label class="field-label">模板名称(可空)</label>
            <input v-model="draft.name" class="input" placeholder="例如:经典蓝" maxlength="24" />
          </div>
          <div class="form-grid-2">
            <div class="field">
              <label class="field-label">标题文字颜色</label>
              <div class="row">
                <input v-model="draft.titleColor" type="color" style="width: 40px; height: 36px; border: none; border-radius: 8px; background: transparent; cursor: pointer" />
                <input v-model="draft.titleColor" class="input" maxlength="7" />
              </div>
            </div>
            <div class="field">
              <label class="field-label">正文文字颜色</label>
              <div class="row">
                <input v-model="draft.textColor" type="color" style="width: 40px; height: 36px; border: none; border-radius: 8px; background: transparent; cursor: pointer" />
                <input v-model="draft.textColor" class="input" maxlength="7" />
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="row">
        <button class="btn btn-ghost" @click="showEdit = false">取消</button>
        <button class="btn btn-primary grow" :disabled="saving" @click="saveDraft">
          <span v-if="saving" class="spinner"></span>
          {{ draft.id == null ? '创建模板' : '保存修改' }}
        </button>
      </div>
    </Modal>
  </div>
</template>
