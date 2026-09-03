<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'

import AppHeader from '@/components/layout/AppHeader.vue'
import Modal from '@/components/ui/Modal.vue'
import { adminApi, uploadApi } from '@/api'
import { toast } from '@/composables/toast'
import type { Template } from '@/types'
import { timeFmt } from '@/utils/format'

const rows = ref<Template[]>([])
const loading = ref(false)
const filter = ref<'0' | '1' | '2'>('0')

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
const draft = reactive<{ id: number | null; name: string; url: string; titleColor: string; textColor: string }>({
  id: null,
  name: '',
  url: '',
  titleColor: '#ffffff',
  textColor: '#f9d9c2'
})
const fileInput = ref<HTMLInputElement | null>(null)
const localFile = ref<File | null>(null)
const localPreview = ref('')

function openCreate() {
  Object.assign(draft, { id: null, name: '', url: '', titleColor: '#ffffff', textColor: '#f9d9c2' })
  localFile.value = null
  localPreview.value = ''
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
  localFile.value = null
  localPreview.value = ''
  showEdit.value = true
}

function pickImage(e: Event) {
  const input = e.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  if (!file || !file.type.startsWith('image/')) {
    toast('请选择图片文件', 'error')
    return
  }
  localFile.value = file
  localPreview.value = URL.createObjectURL(file)
  draft.url = ''
}

async function saveDraft() {
  if (draft.name.trim() && draft.name.trim().length > 24) {
    toast('模板名称请勿超过 24 字', 'error')
    return
  }
  if (!draft.url.trim() && !localFile.value) {
    toast('请上传模板背景图', 'error')
    return
  }
  saving.value = true
  try {
    let url = draft.url.trim()
    if (localFile.value) {
      const res = await uploadApi.upload(localFile.value, localFile.value.name, 'template')
      url = res.url
    }
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
  try {
    await adminApi.update(t.id, { status: next })
    toast(next === 1 ? '已上架' : '已下架', 'success')
    await load()
  } catch (e) {
    toast(e instanceof Error ? e.message : '操作失败', 'error')
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
        <div v-for="t in rows" :key="t.id" class="prod-card" style="cursor: default">
          <div class="prod-img">
            <img :src="t.url" :alt="t.name || `模板${t.id}`" loading="lazy" />
          </div>
          <div class="prod-meta">
            <span class="prod-name" :title="t.name">{{ t.name || `模板 #${t.id}` }}</span>
            <span class="badge" :class="t.status === 1 ? 'ok' : 'off'">{{ t.status === 1 ? '上架' : '下架' }}</span>
          </div>
          <div style="padding: 0 12px 8px" class="row wrap">
            <span class="badge" style="border-color: transparent; background: rgba(255,255,255,0.04)">
              字 <i :style="{ color: t.titleColor }">■</i> {{ t.titleColor }}
            </span>
            <span class="badge" style="border-color: transparent; background: rgba(255,255,255,0.04)">
              文 <i :style="{ color: t.textColor }">■</i> {{ t.textColor }}
            </span>
          </div>
          <div style="padding: 0 10px 10px" class="small muted">{{ timeFmt(t.createTime, false) }} 创建</div>
          <div class="prod-actions">
            <button class="btn btn-ghost btn-sm grow" @click="openEdit(t)">编辑</button>
            <button class="btn btn-sm" :class="t.status === 1 ? 'btn-danger' : 'btn-primary'" @click="toggleStatus(t)">
              {{ t.status === 1 ? '下架' : '上架' }}
            </button>
          </div>
        </div>
      </div>
    </main>

    <!-- 新增/编辑弹窗 -->
    <Modal :show="showEdit" :title="draft.id == null ? '新增模板' : '编辑模板'" @close="showEdit = false">
      <div class="row-between wrap mb-2 form-cols" style="align-items: flex-start">
        <div class="form-preview">
          <img v-if="localPreview || draft.url" :src="localPreview || draft.url" alt="模板预览" style="width: 100%; height: 100%; object-fit: cover" />
          <div v-else style="height:100%;display:grid;place-items:center;color:var(--text-muted)">未选择图片</div>
        </div>
        <div class="grow" style="min-width: 250px">
          <div class="field mb-2">
            <label class="field-label">背景图(建议 700×1400)</label>
            <input ref="fileInput" type="file" accept="image/*" hidden @change="pickImage" />
            <button class="btn btn-ghost btn-sm" @click="fileInput?.click()">
              {{ localFile ? '已选择:重新上传' : '上传图片' }}
            </button>
            <p v-if="localFile" class="small muted">{{ localFile.name }}</p>
          </div>
          <div class="field mb-2">
            <label class="field-label">模板名称(可空)</label>
            <input v-model="draft.name" class="input" placeholder="例如:经典蓝" maxlength="24" />
          </div>
          <div class="form-grid-2">
            <div class="field">
              <label class="field-label">标题文字颜色</label>
              <div class="row">
                <input v-model="draft.titleColor" type="color" style="width: 44px; height: 40px; border: none; border-radius: 8px; background: transparent; cursor: pointer" />
                <input v-model="draft.titleColor" class="input" maxlength="7" />
              </div>
            </div>
            <div class="field">
              <label class="field-label">正文文字颜色</label>
              <div class="row">
                <input v-model="draft.textColor" type="color" style="width: 44px; height: 40px; border: none; border-radius: 8px; background: transparent; cursor: pointer" />
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
