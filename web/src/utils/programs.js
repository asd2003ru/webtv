export function findNowProgram(epgItems) {
  const now = Date.now()
  return epgItems.find((item) => {
    const start = new Date(item.start_at).getTime()
    const end = new Date(item.end_at).getTime()
    return start <= now && now < end
  }) || null
}

export function findProgramByNowHint(epgItems, nowItem) {
  if (!nowItem || !Array.isArray(epgItems) || epgItems.length === 0) return null
  const hintStart = new Date(nowItem.start_at).getTime()
  const hintEnd = new Date(nowItem.end_at).getTime()
  if (Number.isFinite(hintStart) && Number.isFinite(hintEnd) && hintEnd > hintStart) {
    const midpoint = hintStart + Math.floor((hintEnd - hintStart) / 2)
    const byRange = epgItems.find((item) => {
      const start = new Date(item.start_at).getTime()
      const end = new Date(item.end_at).getTime()
      return Number.isFinite(start) && Number.isFinite(end) && start <= midpoint && midpoint < end
    })
    if (byRange) return byRange
  }
  if (nowItem.start_at) {
    const byStart = epgItems.find((item) => item.start_at === nowItem.start_at)
    if (byStart) return byStart
  }
  return null
}

export function findClosestProgramToNow(epgItems) {
  if (!Array.isArray(epgItems) || epgItems.length === 0) return null
  const now = Date.now()
  let lastPast = null
  let firstFuture = null
  for (const item of epgItems) {
    const start = new Date(item.start_at).getTime()
    const end = new Date(item.end_at).getTime()
    if (!Number.isFinite(start) || !Number.isFinite(end)) continue
    if (end <= now) {
      lastPast = item
      continue
    }
    if (start > now) {
      firstFuture = item
      break
    }
  }
  return lastPast || firstFuture || epgItems[0]
}

export function formatProgramDate(value) {
  const date = new Date(value)
  return new Intl.DateTimeFormat(undefined, {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  }).format(date)
}

export function formatSyncDateTime(value) {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '-'
  return new Intl.DateTimeFormat(undefined, {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  }).format(date)
}

export function formatOffset(totalSeconds) {
  const sec = Math.max(0, Math.floor(totalSeconds))
  const hh = String(Math.floor(sec / 3600)).padStart(2, '0')
  const mm = String(Math.floor((sec % 3600) / 60)).padStart(2, '0')
  const ss = String(sec % 60).padStart(2, '0')
  return `${hh}:${mm}:${ss}`
}

export function programKey(program) {
  return `${program.channel_id}-${program.start_at}`
}

export function isCurrentProgram(program, now = Date.now()) {
  const start = new Date(program.start_at).getTime()
  const end = new Date(program.end_at).getTime()
  return start <= now && now < end
}

export function isPastProgram(program, now = Date.now()) {
  const end = new Date(program.end_at).getTime()
  return end <= now
}

export function isFutureProgram(program, now = Date.now()) {
  const start = new Date(program.start_at).getTime()
  return start > now
}

export function programProgressPercent(program, now = Date.now()) {
  const start = new Date(program.start_at).getTime()
  const end = new Date(program.end_at).getTime()
  const duration = end - start
  if (duration <= 0) return 0
  const elapsed = Math.max(0, Math.min(now - start, duration))
  return Math.round((elapsed / duration) * 100)
}

export function programDescriptionTooltip(program) {
  const desc = (program.description || '').trim()
  return desc || 'Описание отсутствует'
}

export function currentOffsetForProgram(program, now = Date.now()) {
  const start = new Date(program.start_at).getTime()
  const offset = Math.floor((now - start) / 1000)
  return offset > 0 ? offset : 0
}
