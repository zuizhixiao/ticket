// 票根 Canvas 绘制管线。
// 布局数值为既有模板调校结果(画布 700×1400),移植保持 1:1,勿臆改。

import type { TicketFields } from '@/types'

export const CANVAS_W = 700
export const CANVAS_H = 1400

export interface TicketTemplateStyle {
  name: string | null // 模板名(URL 文件名去扩展名),决定装饰分支
  titleColor: string
  textColor: string
}

export const DEFAULT_TITLE_COLOR = '#ffffff'
export const DEFAULT_TEXT_COLOR = '#f9d9c2'

const BODY_FONT = '"Microsoft YaHei","PingFang SC",sans-serif'

function roundRectPath(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  w: number,
  h: number,
  r: number
) {
  ctx.beginPath()
  ctx.moveTo(x + r, y)
  ctx.lineTo(x + w - r, y)
  ctx.quadraticCurveTo(x + w, y, x + w, y + r)
  ctx.lineTo(x + w, y + h - r)
  ctx.quadraticCurveTo(x + w, y + h, x + w - r, y + h)
  ctx.lineTo(x + r, y + h)
  ctx.quadraticCurveTo(x, y + h, x, y + h - r)
  ctx.lineTo(x, y + r)
  ctx.quadraticCurveTo(x, y, x + r, y)
  ctx.closePath()
}

/** 模板文件名 → 装饰分支名(与旧逻辑一致)。 */
export function templateStyleNameFromUrl(url: string): string | null {
  const filename = (url.split('/').pop() || '').split('?')[0]
  if (!filename) return null
  const name = filename.replace(/\.[^.]+$/, '')
  const known = ['classic', 'modern', 'vintage', 'elegant']
  return known.includes(name) ? name : null
}

export interface DrawTicketParams {
  ctx: CanvasRenderingContext2D
  width: number
  height: number
  templateImg: HTMLImageElement | null
  style: TicketTemplateStyle
  poster: HTMLImageElement | null
  fields: TicketFields
}

export function drawTicket(p: DrawTicketParams): void {
  const { ctx, width, height, templateImg, style, poster, fields } = p
  ctx.clearRect(0, 0, width, height)

  // ---- 背景模板 ----
  if (templateImg) {
    ctx.drawImage(templateImg, 0, 0, width, height)
  } else {
    ctx.fillStyle = '#ffffff'
    ctx.fillRect(0, 0, width, height)
    ctx.strokeStyle = '#e9ecef'
    ctx.lineWidth = 2
    ctx.strokeRect(10, 10, width - 20, height - 20)
  }

  // ---- 海报 ----
  if (poster) {
    const posterWidth = 617
    const posterHeight = 845
    const posterX = (width - posterWidth) / 2
    const posterY = 40
    const borderRadius = 20

    ctx.save()
    roundRectPath(ctx, posterX, posterY, posterWidth, posterHeight, borderRadius)
    ctx.shadowColor = 'rgba(0,0,0,0.3)'
    ctx.shadowBlur = 15
    ctx.shadowOffsetX = 8
    ctx.shadowOffsetY = 8
    ctx.clip()
    ctx.drawImage(poster, posterX, posterY, posterWidth, posterHeight)
    ctx.restore()

    // 白色描边
    ctx.save()
    roundRectPath(ctx, posterX, posterY, posterWidth, posterHeight, borderRadius)
    ctx.strokeStyle = '#ffffff'
    ctx.lineWidth = 4
    ctx.stroke()
    ctx.restore()
  }

  drawTicketInfo(ctx, width, style, fields)
  drawDecorations(ctx, width, height, style.name)
}

function drawTicketInfo(
  ctx: CanvasRenderingContext2D,
  width: number,
  style: TicketTemplateStyle,
  fields: TicketFields
) {
  const movieTitle = fields.movieTitle.trim() || '电影标题'
  const cinemaName = fields.cinemaName.trim() || '影院名称'
  const hallSeat = fields.hallSeat.trim() || '影厅号及座位号'
  const startTime = fields.startTime.trim() || '开始时间'
  const exclusiveName = fields.exclusiveName.trim() || '特殊ID'
  const timeText = startTime

  const textColor = style.textColor || DEFAULT_TEXT_COLOR
  const titleColor = style.titleColor || DEFAULT_TITLE_COLOR

  const titleFontSize = fields.titleFontSize || 48
  const textFontSize = fields.textFontSize || 26
  const nameFontSize = 35 // 特殊ID字号固定
  const tailFontSize = 24 // 底部小字固定

  const textStartX = 34
  const firstLineY = 1017
  const tailLineY = 1306
  const firstBottomMargin = 35
  const textBottomMargin = 17
  const nameBottomMargin = 7

  ctx.textAlign = 'left'
  ctx.textBaseline = 'alphabetic'

  // 电影标题
  ctx.font = `bold ${titleFontSize}px ${BODY_FONT}`
  ctx.fillStyle = titleColor
  ctx.fillText(movieTitle, textStartX, firstLineY)

  // 语言 + 格式 | 影院
  const secondText = `${fields.language} ${fields.format} | ${cinemaName}`
  const secondLineY = firstLineY + titleFontSize + firstBottomMargin
  ctx.font = `${textFontSize}px ${BODY_FONT}`
  ctx.fillStyle = textColor
  ctx.fillText(secondText, textStartX, secondLineY)
  ctx.fillText(hallSeat, textStartX, secondLineY + textFontSize + textBottomMargin)
  ctx.fillText(timeText, textStartX, secondLineY + (textFontSize + textBottomMargin) * 2)

  // 专属名称与落款(居中)
  ctx.textAlign = 'center'
  ctx.font = `${nameFontSize}px ${BODY_FONT}`
  ctx.fillStyle = titleColor
  ctx.fillText(`@${exclusiveName}`, width / 2, tailLineY)
  ctx.font = `${tailFontSize}px ${BODY_FONT}`
  ctx.fillStyle = textColor
  ctx.fillText('专属纪念票', width / 2, tailLineY + nameBottomMargin + nameFontSize)
}

function drawDecorations(
  ctx: CanvasRenderingContext2D,
  width: number,
  height: number,
  templateName: string | null
) {
  switch (templateName) {
    case 'classic':
      drawCorner(ctx, '#007bff', 20, 20)
      drawCorner(ctx, '#007bff', width - 20, 20)
      drawCorner(ctx, '#007bff', 20, height - 20)
      drawCorner(ctx, '#007bff', width - 20, height - 20)
      break
    case 'modern': {
      ctx.save()
      ctx.strokeStyle = 'rgba(255,255,255,0.3)'
      ctx.lineWidth = 2
      for (let i = 0; i < 3; i++) {
        const x = 80 + i * 80
        const y = 900
        ctx.beginPath()
        ctx.arc(x, y, 40, 0, Math.PI * 2)
        ctx.stroke()
      }
      ctx.restore()
      break
    }
    case 'vintage': {
      ctx.save()
      ctx.strokeStyle = '#8b4513'
      ctx.lineWidth = 2
      ctx.globalAlpha = 0.6
      for (let i = 0; i < 5; i++) {
        drawVintage(ctx, 80 + i * 120, 900, 25)
      }
      ctx.restore()
      break
    }
    case 'elegant': {
      ctx.save()
      ctx.strokeStyle = 'rgba(255,255,255,0.4)'
      ctx.lineWidth = 3
      ctx.beginPath()
      ctx.moveTo(80, 900)
      ctx.quadraticCurveTo(width / 2, 880, width - 80, 900)
      ctx.stroke()
      ctx.restore()
      break
    }
    default:
      break
  }
}

function drawCorner(ctx: CanvasRenderingContext2D, color: string, x: number, y: number) {
  const size = 15
  ctx.save()
  ctx.translate(x, y)
  ctx.strokeStyle = color
  ctx.lineWidth = 2
  ctx.beginPath()
  ctx.moveTo(-size / 2, -size / 2)
  ctx.lineTo(size / 2, -size / 2)
  ctx.moveTo(-size / 2, -size / 2)
  ctx.lineTo(-size / 2, size / 2)
  ctx.stroke()
  ctx.restore()
}

function drawVintage(ctx: CanvasRenderingContext2D, x: number, y: number, size: number) {
  ctx.beginPath()
  ctx.moveTo(x, y - size)
  ctx.lineTo(x + size / 2, y)
  ctx.lineTo(x, y + size)
  ctx.lineTo(x - size / 2, y)
  ctx.closePath()
  ctx.stroke()
}
