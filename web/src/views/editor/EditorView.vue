<script setup lang="ts">
import { onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { useRouter } from 'vue-router'

import AppHeader from '@/components/layout/AppHeader.vue'
import Modal from '@/components/ui/Modal.vue'
import { templateApi, uploadApi } from '@/api'
import { useAuthStore } from '@/stores/auth'
import { toast } from '@/composables/toast'
import { DEFAULT_FIELDS } from '@/types'
import type { Template, TicketFields } from '@/types'
import {
  CANVAS_H,
  CANVAS_W,
  DEFAULT_TEXT_COLOR,
  DEFAULT_TITLE_COLOR,
  drawTicket,
  templateStyleNameFromUrl
} from '@/lib/ticket/drawing'
import type { TicketTemplateStyle } from '@/lib/ticket/drawing'
import {
  canvasToBlob,
  downloadBlob,
  fileToImage,
  isWeChatBrowser,
  loadImage,
  safeFilename
} from '@/lib/ticket/media'

interface LoadedTpl {
  item: Template
  img: HTMLImageElement | null
}

const canvasRef = ref<HTMLCanvasElement | null>(null)
const posterInput = ref<HTMLInputElement | null>(null)
const fileInputEl = ref<HTMLInputElement | null>(null)

const templates = ref<LoadedTpl[]>([])
const loadingTemplates = ref(true)
const current = ref<LoadedTpl | null>(null)
const poster = ref<HTMLImageElement | null>(null)
const dragging = ref(false)

const fields = reactive<TicketFields>({ ...DEFAULT_FIELDS })
const saving = ref(false)

const previewUrl = ref('')
const previewName = ref('')
const wechat = isWeChatBrowser()

const auth = useAuthStore()
const router = useRouter()

let ctx: CanvasRenderingContext2D | null = null
let renderTimer: number | undefined

function render() {
  if (!ctx) return
  const style: TicketTemplateStyle = current.value
    ? {
        name: templateStyleNameFromUrl(current.value.item.url),
        titleColor: current.value.item.titleColor,
        textColor: current.value.item.textColor
      }
    : { name: null, titleColor: DEFAULT_TITLE_COLOR, textColor: DEFAULT_TEXT_COLOR }
  drawTicket({
    ctx,
    width: CANVAS_W,
    height: CANVAS_H,
    templateImg: current.value?.img ?? null,
    style,
    poster: poster.value,
    fields
  })
}

function scheduleRender() {
  if (renderTimer) window.clearTimeout(renderTimer)
  renderTimer = window.setTimeout(render, 60)
}

watch(fields, scheduleRender, { deep: true })

onMounted(async () => {
  ctx = canvasRef.value?.getContext('2d') ?? null
  render()
  await loadTemplates()
})

onBeforeUnmount(() => {
  if (renderTimer) window.clearTimeout(renderTimer)
})

// ---------- 模板 ----------
async function loadTemplates() {
  loadingTemplates.value = true
  try {
    const { list } = await templateApi.listPublic()
    const loaded: LoadedTpl[] = []
    for (const item of list) {
      let img: HTMLImageElement | null = null
      try {
        img = await loadImage(item.url)
      } catch {
        /* 单个模板失败不阻断 */
      }
      loaded.push({ item, img })
    }
    templates.value = loaded
    if (!current.value && loaded.length) selectTpl(loaded[0])
  } catch {
    toast('模板加载失败,请刷新重试', 'error')
  } finally {
    loadingTemplates.value = false
    render()
  }
}

function selectTpl(t: LoadedTpl) {
  current.value = t
  render()
}

// ---------- 海报 ----------
function pickPoster() {
  fileInputEl.value?.click()
}

async function onPosterFile(e: Event) {
  const input = e.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  if (file) await applyPoster(file)
}

async function applyPoster(file: File) {
  if (!file.type.startsWith('image/')) {
    toast('请选择图片文件', 'error')
    return
  }
  try {
    poster.value = await fileToImage(file)
    render()
    // 选择海报后立即异步上传(游客也上传,静默失败不影响本地预览)
    uploadApi.upload(file, file.name, 'poster').catch(() => undefined)
  } catch {
    toast('海报读取失败', 'error')
  }
}

function onDrop(e: DragEvent) {
  dragging.value = false
  const file = e.dataTransfer?.files?.[0]
  if (file) void applyPoster(file)
}

function removePoster() {
  poster.value = null
  render()
}

// ---------- 生成 ----------
function requireText(key: 'movieTitle' | 'hallSeat' | 'startTime' | 'exclusiveName', label: string): boolean {
  if (!fields[key].trim()) {
    toast(`请填写${label}`, 'error')
    return false
  }
  return true
}

async function generate() {
  if (!poster.value) {
    toast('请先上传电影海报', 'error')
    return
  }
  if (!requireText('movieTitle', '电影标题')) return
  if (!requireText('hallSeat', '影厅号及座位号')) return
  if (!requireText('startTime', '开始时间')) return
  if (!requireText('exclusiveName', '特殊ID')) return

  const canvas = canvasRef.value
  if (!canvas) return

  saving.value = true
  try {
    const blob = await canvasToBlob(canvas)
    const name = `${safeFilename(fields.movieTitle)}_${Date.now()}.png`
    previewName.value = name
    previewUrl.value = URL.createObjectURL(blob)

    // 1) 无论是否登录都上传成品图到云端(失败不阻断本地下载)
    try {
      const res = await uploadApi.upload(blob, name, 'product')
      if (auth.token) {
        toast('已保存到「我的成品」', 'success')
      } else {
        toast('图片已生成并上传云端', 'success')
      }
      if (res.url) previewUrl.value = res.url
    } catch {
      toast('云端保存失败,已保留本地图片', 'error')
    }

    // 2) 微信内长按保存 / 常规下载
    if (!wechat) downloadBlob(blob, name)
  } catch {
    toast('生成失败,请重试', 'error')
  } finally {
    saving.value = false
  }
}

function resetEditor() {
  Object.assign(fields, { ...DEFAULT_FIELDS })
  poster.value = null
  render()
}

function toLogin() {
  router.push({ path: '/login', query: { redirect: '/' } })
}

function downloadPreview() {
  if (!previewUrl.value) return
  const a = document.createElement('a')
  a.href = previewUrl.value
  a.download = previewName.value || 'ticket.png'
  document.body.appendChild(a)
  a.click()
  a.remove()
}
</script>

<template>
  <div class="page">
    <AppHeader />

    <main class="editor-main fade-up">
      <!-- 左:预览舞台 -->
      <section class="stage-wrap">
        <div class="stage">
          <canvas ref="canvasRef" width="700" height="1400"></canvas>
        </div>

        <div class="row">
          <button class="btn btn-primary" :class="{ grow: true }" :disabled="saving" @click="generate">
            <span v-if="saving" class="spinner"></span>
            <span v-else>✦</span>
            {{ saving ? '正在生成并保存…' : '生成票根并保存' }}
          </button>
          <button class="btn btn-ghost" @click="resetEditor">重置</button>
        </div>
        <p class="small muted" style="text-align: center">
          {{
            wechat
              ? '生成后长按图片即可保存到相册 📱'
              : auth.token
                ? '已生成,自动下载并同步到「我的成品」💾'
                : '无需登录,生成后同步保存到云端并自动下载 💾'
          }}
        </p>
      </section>

      <!-- 右:控制面板 -->
      <aside class="control-panel">
        <!-- 模板 -->
        <div class="panel control-section">
          <h3 class="panel-title">🎨 选择背景模板</h3>
          <div v-if="loadingTemplates" class="tpl-grid">
            <div v-for="n in 6" :key="n" class="skel" style="aspect-ratio: 2/4"></div>
          </div>
          <div v-else class="tpl-grid">
            <div
              v-for="t in templates"
              :key="t.item.id"
              class="tpl-item"
              :class="{ active: current === t }"
              @click="selectTpl(t)"
            >
              <img v-if="t.img" :src="t.item.url" alt="" loading="lazy" />
              <div v-else style="height:100%;background:#000"></div>
              <span class="tpl-tick">✓</span>
            </div>
          </div>
        </div>

        <!-- 海报 -->
        <div class="panel control-section">
          <h3 class="panel-title">🎭 上传电影海报</h3>
          <div
            v-if="!poster"
            class="poster-drop"
            :class="{ drag: dragging }"
            @click="pickPoster"
            @dragover.prevent="dragging = true"
            @dragleave="dragging = false"
            @drop.prevent="onDrop"
          >
            <div style="font-size: 34px">🖼️</div>
            <div>点击选择或拖拽海报到此处</div>
            <div class="small muted mt-1">支持 JPG / PNG / GIF / WebP</div>
          </div>
          <div v-else class="row wrap">
            <div class="poster-prev">
              <img :src="poster.src" alt="海报预览" />
              <button class="poster-rm" title="移除海报" @click="removePoster">✕</button>
            </div>
            <div class="small muted" style="flex:1;min-width:120px">
              <p>海报已就绪 ✓</p>
              <p style="font-size:12px">将按票面区域裁切显示</p>
            </div>
          </div>
          <input
            ref="fileInputEl"
            type="file"
            accept="image/*"
            hidden
            @change="onPosterFile"
          />
        </div>

        <!-- 信息 -->
        <div class="panel control-section">
          <h3 class="panel-title">🎫 编辑票根信息</h3>
          <div class="field mb-2">
            <label class="field-label" for="movieTitle">电影标题 *</label>
            <input v-model="fields.movieTitle" id="movieTitle" class="input" placeholder="例如:流浪地球" />
          </div>
          <div class="form-grid-2 mb-2">
            <div class="field">
              <label class="field-label">语言</label>
              <select v-model="fields.language" class="select">
                <option>国语</option>
                <option>粤语</option>
                <option>日语</option>
                <option>英语</option>
                <option>韩语</option>
                <option>法语</option>
              </select>
            </div>
            <div class="field">
              <label class="field-label">格式</label>
              <select v-model="fields.format" class="select">
                <option>2D</option>
                <option>3D</option>
                <option>IMAX</option>
                <option>杜比全景声</option>
              </select>
            </div>
          </div>
          <div class="field mb-2">
            <label class="field-label" for="cinemaName">影院名称</label>
            <input v-model="fields.cinemaName" id="cinemaName" class="input" placeholder="例如:星光国际影城" />
          </div>
          <div class="field mb-2">
            <label class="field-label" for="hallSeat">影厅号及座位号 *</label>
            <input v-model="fields.hallSeat" id="hallSeat" class="input" placeholder="例如:3号厅 7排12座" />
          </div>
          <div class="field mb-2">
            <label class="field-label" for="startTime">放映时间 *</label>
            <input v-model="fields.startTime" id="startTime" class="input" placeholder="例如:2026-02-14 19:30" />
          </div>
          <div class="field">
            <label class="field-label" for="exclusiveName">特殊 ID *</label>
            <input v-model="fields.exclusiveName" id="exclusiveName" class="input" placeholder="例如:LOVE0214" />
            <p class="small muted">将显示在票根底部 @你的专属ID</p>
          </div>
        </div>

        <!-- 排版 -->
        <div class="panel control-section">
          <h3 class="panel-title">🔤 字体大小</h3>
          <div class="row mb-2">
            <span class="field-label" style="width: 92px">电影标题</span>
            <input v-model.number="fields.titleFontSize" type="range" min="20" max="60" class="range" />
            <span class="small muted" style="width: 52px; text-align: right">{{ fields.titleFontSize }}px</span>
          </div>
          <div class="row">
            <span class="field-label" style="width: 92px">正文字体</span>
            <input v-model.number="fields.textFontSize" type="range" min="12" max="40" class="range" />
            <span class="small muted" style="width: 52px; text-align: right">{{ fields.textFontSize }}px</span>
          </div>
        </div>
      </aside>
    </main>

    <!-- 预览大图 -->
    <Modal :show="!!previewUrl" title="你的专属票根" @close="previewUrl = ''">
      <img v-if="previewUrl" :src="previewUrl" class="lightbox-img" alt="票根成品" />
      <div class="row mt-2">
        <button class="btn btn-ghost" @click="previewUrl = ''">关闭</button>
        <button class="btn btn-primary grow" @click="downloadPreview">💾 下载图片</button>
      </div>
      <div v-if="!auth.token" class="row mt-1" style="justify-content: center">
        <span class="small muted">想长期保存?</span>
        <a class="small text-gold-grad" style="cursor: pointer" @click="toLogin">
          登录并同步到「我的成品」→
        </a>
      </div>
      <p v-if="wechat" class="small muted mt-1" style="text-align:center">
        💡 在微信中,请长按上方图片选择「保存到手机」
      </p>
    </Modal>
  </div>
</template>
