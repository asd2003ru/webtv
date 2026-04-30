export function channelInitial(name) {
  const first = (name || '').trim().charAt(0)
  return first ? first.toUpperCase() : 'TV'
}

export function fuzzyScore(text, query) {
  const source = String(text || '').toLowerCase()
  const needle = String(query || '').toLowerCase().trim()
  if (!needle) return 1
  let idx = 0
  for (const ch of needle) {
    idx = source.indexOf(ch, idx)
    if (idx < 0) return 0
    idx += 1
  }
  return 1
}

export function favoriteKey(playlistID, channelID) {
  return `${playlistID}:${channelID}`
}

export function favoriteExternalKey(playlistID, externalID) {
  return `${playlistID}:ext:${externalID || ''}`
}

export function favoriteStorageKey(item) {
  if (item?.external_id) {
    return favoriteExternalKey(item.playlist_id, item.external_id)
  }
  return favoriteKey(item?.playlist_id, item?.id)
}

export function findChannelByFavorite(item, list) {
  if (!item || !item.playlist_id || !Array.isArray(list)) return null
  if (item.external_id) {
    const byExternal = list.find((c) => c.playlist_id === item.playlist_id && c.external_id === item.external_id)
    if (byExternal) return byExternal
  }
  if (item.id) {
    const byID = list.find((c) => c.playlist_id === item.playlist_id && c.id === item.id)
    if (byID) return byID
  }
  return null
}
