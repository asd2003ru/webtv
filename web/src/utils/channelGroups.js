export function buildSettingsGroupedChannels(channels) {
  const groups = new Map()
  for (const channel of channels) {
    const groupName = channel.group || 'Без группы'
    if (!groups.has(groupName)) groups.set(groupName, [])
    groups.get(groupName).push(channel)
  }
  return [...groups.entries()]
    .sort((a, b) => a[0].localeCompare(b[0]))
    .map(([name, list]) => ({
      name,
      channels: [...list].sort((a, b) => (a.name || '').localeCompare(b.name || ''))
    }))
}

export function buildGroupedChannels(channels, { isHiddenChannel, isHiddenGroup, matchesChannelSearch }) {
  const groups = new Map()
  for (const channel of channels) {
    if (isHiddenChannel(channel.playlist_id, channel.id, channel.external_id)) continue
    const groupName = channel.group || 'Без группы'
    if (isHiddenGroup(groupName)) continue
    if (!matchesChannelSearch(channel.name || '')) continue
    if (!groups.has(groupName)) groups.set(groupName, [])
    groups.get(groupName).push(channel)
  }
  return [...groups.entries()].map(([name, list]) => ({
    name,
    channels: list
  }))
}

export function buildPlaylistVisibleCountById(playlists, visibleCountByPlaylist) {
  const counts = Object.fromEntries(playlists.map((playlist) => [playlist.id, Number(playlist.channel_count || 0)]))
  for (const [playlistID, visibleCount] of Object.entries(visibleCountByPlaylist)) {
    counts[Number(playlistID)] = Number(visibleCount || 0)
  }
  return counts
}
