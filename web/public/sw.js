const LOGO_CACHE = 'webtv-logo-cache-v1'

self.addEventListener('install', () => {
  self.skipWaiting()
})

self.addEventListener('activate', (event) => {
  event.waitUntil((async () => {
    const keys = await caches.keys()
    await Promise.all(
      keys
        .filter((key) => key.startsWith('webtv-logo-cache-') && key !== LOGO_CACHE)
        .map((key) => caches.delete(key))
    )
    await self.clients.claim()
  })())
})

self.addEventListener('fetch', (event) => {
  const { request } = event
  if (request.method !== 'GET') return
  if (request.destination !== 'image') return

  let url
  try {
    url = new URL(request.url)
  } catch {
    return
  }
  if (url.protocol !== 'http:' && url.protocol !== 'https:') return

  event.respondWith((async () => {
    const cache = await caches.open(LOGO_CACHE)
    const cached = await cache.match(request)
    if (cached) return cached

    const response = await fetch(request)
    if (response && (response.ok || response.type === 'opaque')) {
      await cache.put(request, response.clone())
    }
    return response
  })())
})
