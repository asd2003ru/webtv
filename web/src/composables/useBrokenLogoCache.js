import { ref } from 'vue'

const BROKEN_LOGOS_STORAGE_KEY = 'webtv_broken_logos_v1'
const BROKEN_LOGO_TTL_MS = 30 * 24 * 60 * 60 * 1000
const BROKEN_LOGO_MAX_ENTRIES = 2000

function logoKey(url) {
  return `url:${url}`
}

export function useBrokenLogoCache() {
  const brokenLogos = ref({})
  let brokenLogoExpiresAt = {}
  let brokenLogoSaveTimer = null

  function showChannelLogo(channel) {
    const url = (channel.logo || '').trim()
    const isRenderableLogo = url.startsWith('/api/logos/') || url.startsWith('data:image/') || url.startsWith('blob:')
    return isRenderableLogo && !brokenLogos.value[logoKey(url)]
  }

  function markLogoError(channel) {
    const url = (channel?.logo || '').trim()
    if (!url) return
    const key = logoKey(url)
    brokenLogos.value[key] = true
    brokenLogoExpiresAt[key] = Date.now() + BROKEN_LOGO_TTL_MS
    scheduleBrokenLogosSave()
  }

  function loadBrokenLogos() {
    try {
      const raw = localStorage.getItem(BROKEN_LOGOS_STORAGE_KEY)
      const parsed = raw ? JSON.parse(raw) : {}
      const entries = parsed && typeof parsed === 'object' && parsed.entries && typeof parsed.entries === 'object'
        ? parsed.entries
        : {}
      const now = Date.now()
      const nextBrokenLogos = {}
      const nextExpiresAt = {}
      for (const [key, expiresAt] of Object.entries(entries)) {
        if (!key.startsWith('url:')) continue
        const expiry = Number(expiresAt)
        if (!Number.isFinite(expiry) || expiry <= now) continue
        nextBrokenLogos[key] = true
        nextExpiresAt[key] = expiry
      }
      brokenLogos.value = nextBrokenLogos
      brokenLogoExpiresAt = nextExpiresAt
      if (Object.keys(entries).length !== Object.keys(nextBrokenLogos).length) {
        scheduleBrokenLogosSave()
      }
    } catch {
      brokenLogos.value = {}
      brokenLogoExpiresAt = {}
    }
  }

  function scheduleBrokenLogosSave() {
    if (brokenLogoSaveTimer) return
    brokenLogoSaveTimer = window.setTimeout(() => {
      brokenLogoSaveTimer = null
      saveBrokenLogos()
    }, 500)
  }

  function saveBrokenLogos() {
    try {
      const now = Date.now()
      const entries = Object.entries(brokenLogoExpiresAt)
        .filter(([key, expiresAt]) => brokenLogos.value[key] && Number(expiresAt) > now)
        .sort((a, b) => Number(b[1]) - Number(a[1]))
        .slice(0, BROKEN_LOGO_MAX_ENTRIES)
      const nextExpiresAt = Object.fromEntries(entries)
      const nextBrokenLogos = Object.fromEntries(entries.map(([key]) => [key, true]))
      brokenLogoExpiresAt = nextExpiresAt
      brokenLogos.value = nextBrokenLogos
      localStorage.setItem(BROKEN_LOGOS_STORAGE_KEY, JSON.stringify({ entries: nextExpiresAt }))
    } catch {
      // Cache is a performance hint; quota or privacy-mode failures can be ignored.
    }
  }

  function flushBrokenLogos() {
    if (!brokenLogoSaveTimer) return
    clearTimeout(brokenLogoSaveTimer)
    brokenLogoSaveTimer = null
    saveBrokenLogos()
  }

  async function clearLogoCaches() {
    if (brokenLogoSaveTimer) {
      clearTimeout(brokenLogoSaveTimer)
      brokenLogoSaveTimer = null
    }
    brokenLogos.value = {}
    brokenLogoExpiresAt = {}
    try {
      localStorage.removeItem(BROKEN_LOGOS_STORAGE_KEY)
    } catch {
      // Cache is a performance hint; storage failures can be ignored.
    }
    if (!window.caches?.keys) return
    try {
      const keys = await window.caches.keys()
      await Promise.all(
        keys
          .filter((key) => key.startsWith('webtv-logo-cache-'))
          .map((key) => window.caches.delete(key))
      )
    } catch {
      // Cache API may be unavailable in private mode or older browsers.
    }
  }

  return {
    showChannelLogo,
    markLogoError,
    loadBrokenLogos,
    clearLogoCaches,
    flushBrokenLogos
  }
}
