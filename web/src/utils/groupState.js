export function getGroupStateForPlaylist(storageKey, playlistID) {
  const all = loadGroupState(storageKey)
  return all[String(playlistID || 0)] || null
}

export function setGroupStateForPlaylist(storageKey, playlistID, state) {
  const all = loadGroupState(storageKey)
  all[String(playlistID || 0)] = state
  localStorage.setItem(storageKey, JSON.stringify(all))
}

export function mergeGroupState(defaultState, savedState, fallbackOpen) {
  const next = savedState && typeof savedState === 'object' ? { ...savedState } : {}
  for (const name of Object.keys(defaultState)) {
    if (savedState && Object.prototype.hasOwnProperty.call(savedState, name)) {
      next[name] = savedState[name] !== false
      continue
    }
    next[name] = fallbackOpen
  }
  return next
}

function loadGroupState(storageKey) {
  try {
    const raw = localStorage.getItem(storageKey)
    const parsed = raw ? JSON.parse(raw) : {}
    return parsed && typeof parsed === 'object' ? parsed : {}
  } catch {
    return {}
  }
}
