export function timeFmt(sec?: number, withTime = true): string {
  if (!sec) return '—'
  const d = new Date(sec * 1000)
  const p = (n: number) => String(n).padStart(2, '0')
  const date = `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())}`
  if (!withTime) return date
  return `${date} ${p(d.getHours())}:${p(d.getMinutes())}`
}
