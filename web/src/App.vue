<template>
  <main class="layout">
    <header class="topbar">
      <h1>{{ appTitle }}</h1>
    </header>

    <PlaylistsPanel
      :tab="tab"
      :t="t"
      :form="form"
      :editing-id="editingId"
      :playlists="playlists"
      :refreshing-playlist-ids="refreshingPlaylistIds"
      :show-debug-overlay="showDebugOverlay"
      :ui-language-mode="uiLanguageMode"
      :theme-mode="themeMode"
      :default-audio-language="defaultAudioLanguage"
      :default-audio-language-options="defaultAudioLanguageOptions"
      :selected-playlist="selectedPlaylist"
      :settings-grouped-channels="settingsGroupedChannels"
      :format-sync-date-time="formatSyncDateTime"
      :is-settings-group-open="isSettingsGroupOpen"
      :is-hidden-group="isHiddenGroup"
      :is-hidden-channel="isHiddenChannel"
      :on-close="() => { tab = 'player' }"
      :on-save-playlist="savePlaylist"
      :on-cancel-edit="cancelEdit"
      :on-refresh="refresh"
      :on-start-edit="startEdit"
      :on-remove-playlist="removePlaylist"
      :on-toggle-debug-overlay="toggleDebugOverlay"
      :on-language-mode-change="setLanguageMode"
      :on-theme-mode-change="setThemeMode"
      :on-default-audio-language-change="setDefaultAudioLanguage"
      :on-select-playlist="selectPlaylist"
      :on-toggle-settings-group="toggleSettingsGroup"
      :on-toggle-hidden-group="toggleHiddenGroup"
      :on-toggle-hidden-channel="toggleHiddenChannel"
    />

    <section v-show="tab==='player'" class="panel">
      <div class="playlist-switcher">
        <div class="playlist-switcher-list">
          <button
            v-for="p in playlists"
            :key="p.id"
            type="button"
            class="btn playlist-btn"
            :class="{ active: selectedPlaylist === p.id }"
            @click="selectPlaylist(p.id)"
          >
            <span>{{ p.name }}</span>
            <span class="playlist-btn-badge">{{ playlistVisibleCountById[p.id] ?? 0 }}</span>
          </button>
        </div>
        <div class="playlist-switcher-actions">
          <button
            type="button"
            class="settings-btn"
            :aria-label="t('settings')"
            :title="t('settings')"
            @click="tab='playlists'"
          >
            <svg viewBox="0 0 24 24" aria-hidden="true">
              <path d="M19.14 12.94a7.49 7.49 0 0 0 .05-.94 7.49 7.49 0 0 0-.05-.94l2.03-1.58a.5.5 0 0 0 .12-.64l-1.92-3.32a.5.5 0 0 0-.6-.22l-2.39.96a7.25 7.25 0 0 0-1.63-.94l-.36-2.54A.5.5 0 0 0 13.9 2h-3.8a.5.5 0 0 0-.49.42l-.36 2.54a7.25 7.25 0 0 0-1.63.94l-2.39-.96a.5.5 0 0 0-.6.22L2.71 8.48a.5.5 0 0 0 .12.64l2.03 1.58a7.49 7.49 0 0 0-.05.94c0 .32.02.63.05.94l-2.03 1.58a.5.5 0 0 0-.12.64l1.92 3.32a.5.5 0 0 0 .6.22l2.39-.96c.5.39 1.04.71 1.63.94l.36 2.54a.5.5 0 0 0 .49.42h3.8a.5.5 0 0 0 .49-.42l.36-2.54a7.25 7.25 0 0 0 1.63-.94l2.39.96a.5.5 0 0 0 .6-.22l1.92-3.32a.5.5 0 0 0-.12-.64l-2.03-1.58zM12 15.5A3.5 3.5 0 1 1 12 8a3.5 3.5 0 0 1 0 7.5z" />
            </svg>
          </button>
          <button
            type="button"
            class="btn channel-toggle-btn"
            :aria-label="sidebarOpen ? t('hide_channels') : t('show_channels')"
            :title="sidebarOpen ? t('hide_channels') : t('show_channels')"
            @click="sidebarOpen = !sidebarOpen"
          >
            <svg v-if="sidebarOpen" viewBox="0 0 24 24" aria-hidden="true">
              <path d="M4 6h16v2H4zm0 5h16v2H4zm0 5h16v2H4z" />
            </svg>
            <svg v-else viewBox="0 0 24 24" aria-hidden="true">
              <path d="M4 6h14v2H4zm0 5h10v2H4zm0 5h16v2H4z" />
            </svg>
          </button>
        </div>
      </div>
      <div class="player-layout" :class="{ expanded: !sidebarOpen }">
        <div class="player-main">
          <PlayerSurface
            :on-set-player-wrap-ref="setPlayerWrapEl"
            :on-set-video-ref="setVideoEl"
            :player-aspect-ratio="playerAspectRatio"
            :current-video-fit-mode="currentVideoFitMode"
            :startup-overlay-visible="startupOverlayVisible"
            :startup-overlay-title="startupOverlayTitle"
            :startup-overlay-text="startupOverlayText"
            :controls-visible="controlsVisible"
            :is-playing="isPlaying"
            :is-muted="isMuted"
            :volume-level="volumeLevel"
            :can-prev-program="canPrevProgram"
            :can-next-program="canNextProgram"
            :can-seek-archive="canSeekArchive"
            :video-fit-mode-label="videoFitModeLabel"
            :selected-channel="selectedChannel"
            :is-in-picture-in-picture="isInPictureInPicture"
            :picture-in-picture-supported="pictureInPictureSupported"
            :archive-supported="archiveSupported"
            :selected-program="selectedProgram"
            :timeshift-max-seconds="timeshiftMaxSeconds"
            :timeshift-offset-seconds="timeshiftOffsetSeconds"
            :timeshift-offset-label="timeshiftOffsetLabel"
            :timeshift-max-label="timeshiftMaxLabel"
            :show-debug-overlay="showDebugOverlay"
            :hls-debug="hlsDebug"
            :on-video-metadata="onVideoMetadata"
            :on-video-ended="onVideoEnded"
            :on-video-play="onVideoPlay"
            :on-video-pause="onVideoPause"
            :on-video-volume-change="onVideoVolumeChange"
            :on-video-time-update="onVideoTimeUpdate"
            :on-video-can-play="onVideoCanPlay"
            :on-video-error="onVideoError"
            :on-video-waiting="onVideoWaiting"
            :on-video-stalled="onVideoStalled"
            :on-reveal-controls="revealControls"
            :on-toggle-play-pause="togglePlayPause"
            :on-toggle-mute="toggleMute"
            :on-volume-input="onVolumeInput"
            :on-select-prev-program="selectPrevProgram"
            :on-select-next-program="selectNextProgram"
            :on-seek-by-seconds="seekBySeconds"
            :on-toggle-fullscreen="toggleFullscreen"
            :on-cycle-video-fit-mode="cycleVideoFitMode"
            :on-detach-player="detachPlayer"
            :on-timeshift-input="onTimeshiftInput"
            :on-apply-timeshift="applyTimeshift"
          />
          <PlayerStatusPrograms
            :t="t"
            :stream-mode-display="streamModeDisplay"
            :stream-mode="streamMode"
            :archive-supported="archiveSupported"
            :selected-program="selectedProgram"
            :selected-program-progress-label="selectedProgramProgressLabel"
            :audio-track-options="audioTrackOptions"
            :selected-audio-track="selectedAudioTrack"
            :deinterlace-enabled="deinterlaceEnabled"
            :selected-channel="selectedChannel"
            :programs="programs"
            :selected-program-key="selectedProgramKey"
            :program-key="programKey"
            :is-current-program="isCurrentProgram"
            :is-archive-playing-program="isArchivePlayingProgram"
            :is-future-program="isFutureProgram"
            :program-description-tooltip="programDescriptionTooltip"
            :format-program-date="formatProgramDate"
            :program-progress-percent="programProgressPercent"
            :on-select-program="selectProgram"
            :on-toggle-deinterlace="toggleDeinterlace"
            :on-select-audio-track="onSelectAudioTrack"
          />
        </div>

        <ChannelSidebar
          :sidebar-open="sidebarOpen"
          :t="t"
          :channel-search-query="channelSearchQuery"
          :favorite-channels="favoriteChannels"
          :selected-channel="selectedChannel"
          :selected-playlist="selectedPlaylist"
          :playlist-name-by-id="playlistNameById"
          :grouped-channels="groupedChannels"
          :now-program-by-channel="nowProgramByChannel"
          :is-group-open="isGroupOpen"
          :show-channel-logo="showChannelLogo"
          :channel-initial="channelInitial"
          :favorite-now-program-title="favoriteNowProgramTitle"
          :is-favorite="isFavorite"
          :on-toggle-group="toggleGroup"
          :on-pick-channel="pickChannel"
          :on-pick-favorite="pickFavorite"
          :on-toggle-favorite="toggleFavorite"
          :on-mark-logo-error="markLogoError"
          :on-mark-favorite-logo-error="markFavoriteLogoError"
          :on-set-channel-search-query="setChannelSearchQuery"
        />

      </div>
    </section>
  </main>
</template>

<script setup>
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import Hls from 'hls.js'
import PlaylistsPanel from './components/PlaylistsPanel.vue'
import ChannelSidebar from './components/ChannelSidebar.vue'
import PlayerStatusPrograms from './components/PlayerStatusPrograms.vue'
import PlayerSurface from './components/PlayerSurface.vue'
import { api, asArray } from './api'
import { useUILanguage } from './composables/useUILanguage'
import { useBrokenLogoCache } from './composables/useBrokenLogoCache'
import {
  channelInitial,
  favoriteKey,
  favoriteStorageKey,
  findChannelByFavorite,
  fuzzyScore
} from './utils/channels'
import {
  buildGroupedChannels,
  buildPlaylistVisibleCountById,
  buildSettingsGroupedChannels
} from './utils/channelGroups'
import { getGroupStateForPlaylist, mergeGroupState, setGroupStateForPlaylist } from './utils/groupState'
import {
  currentOffsetForProgram as programCurrentOffset,
  findClosestProgramToNow,
  findNowProgram,
  findProgramByNowHint,
  formatOffset,
  formatProgramDate,
  formatSyncDateTime,
  isCurrentProgram,
  isFutureProgram,
  isPastProgram,
  programDescriptionTooltip,
  programKey,
  programProgressPercent as calculateProgramProgressPercent
} from './utils/programs'

const tab = ref('player')
const appTitle = ref('WebTV')
const playlists = ref([])
const channels = ref([])
const programs = ref([])
const selectedPlaylist = ref(0)
const selectedChannel = ref(0)
const streamMode = ref('')
const audioTracks = ref([])
const selectedAudioTrack = ref(-1)
const audioTrackByChannel = ref({})
const audioForceTranscode = ref(false)
const pendingAudioTrackApply = ref(-1)
const channelSearchQuery = ref('')
const archiveSupported = ref(false)
const selectedProgramKey = ref('')
const video = ref(null)
const playerWrap = ref(null)
function setVideoEl(el) {
  video.value = el
}
function setPlayerWrapEl(el) {
  playerWrap.value = el
}
const editingId = ref(0)
const isPlaying = ref(false)
const isMuted = ref(false)
const volumeLevel = ref(1)
const playerAspectRatio = ref('16 / 9')
const controlsVisible = ref(true)
const sidebarOpen = ref(true)
const nowProgramByChannel = ref({})
const openGroups = ref({})
const settingsOpenGroups = ref({})
const refreshingPlaylistIds = ref({})
const pictureInPictureSupported = ref(false)
const isInPictureInPicture = ref(false)
const favorites = ref([])
const hiddenGroups = ref([])
const hiddenChannels = ref([])
const visibleCountByPlaylist = ref({})
const favoriteNowProgramByKey = ref({})
const nowTickMs = ref(Date.now())
let nowProgramsRequestToken = 0
let favoriteNowProgramsRequestToken = 0
let channelSelectRequestToken = 0
let playbackTicker = null
let nowTicker = null
let controlsHideTimer = null
const currentVideoSrc = ref('')
let hls = null
let lastForcedReloadAtMs = 0
const forceReloadCooldownMs = 1500
let stalledRecoveryTimer = null
const showDebugOverlay = ref(true)
const startupState = ref('idle')
const deinterlaceEnabled = ref(false)
const themeMode = ref('system')
const defaultAudioLanguage = ref('auto')
const defaultAudioLanguageOptions = [
  { value: 'auto', label: 'Auto' },
  { value: 'ru', label: 'Russian (RU)' },
  { value: 'en', label: 'English (EN)' },
  { value: 'uk', label: 'Ukrainian (UK)' },
  { value: 'de', label: 'German (DE)' },
  { value: 'fr', label: 'French (FR)' },
  { value: 'es', label: 'Spanish (ES)' },
  { value: 'it', label: 'Italian (IT)' },
  { value: 'pt', label: 'Portuguese (PT)' },
  { value: 'pl', label: 'Polish (PL)' },
  { value: 'tr', label: 'Turkish (TR)' },
  { value: 'ar', label: 'Arabic (AR)' },
  { value: 'zh', label: 'Chinese (ZH)' },
  { value: 'ja', label: 'Japanese (JA)' },
  { value: 'ko', label: 'Korean (KO)' },
  { value: 'ro', label: 'Romanian (RO)' }
]
const {
  uiLanguageMode,
  t,
  streamModeDisplay,
  onLanguageModeChange,
  initLanguageMode
} = useUILanguage()
const {
  showChannelLogo,
  markLogoError,
  loadBrokenLogos,
  flushBrokenLogos
} = useBrokenLogoCache()
const hlsDebug = ref({
  engine: 'native',
  serverMode: '-',
  videoPath: '-',
  audioPath: '-',
  audioTracks: 0,
  audioTarget: '-',
  audioResult: '-',
  level: '-',
  buffer: '0.0s',
  latency: '-',
  retries: 0,
  state: 'idle',
  lastError: ''
})
const audioTrackOptions = computed(() => audioTracks.value.map((track) => ({
  value: track.index,
  label: track.label
})))
let hlsRetries = 0
let startupSlowTimer = null
let systemThemeMediaQuery = null
let audioSwitchProbeTimer = null
let audioSwitchProbeToken = 0
const playbackAnchor = ref({
  active: false,
  baseOffsetSeconds: 0,
  startedAtMs: 0
})
const videoFitModes = ['contain', 'cover', 'fill']
const videoFitByChannel = ref({})
const timeshiftDragging = ref(false)
const startupOverlayVisible = computed(() => startupState.value !== 'idle')
const startupOverlayTitle = computed(() => {
  if (startupState.value === 'transcoding') return t('startup_transcoding_title')
  if (startupState.value === 'connecting_slow') return t('startup_slow_title')
  return t('startup_connecting_title')
})
const startupOverlayText = computed(() => {
  if (startupState.value === 'transcoding') return t('startup_transcoding_text')
  if (startupState.value === 'connecting_slow') return t('startup_slow_text')
  return t('startup_connecting_text')
})
const startupBadgeLabel = computed(() => {
  if (startupState.value === 'transcoding') return t('startup_transcoding_badge')
  if (startupState.value === 'connecting_slow') return t('startup_slow_badge')
  return t('startup_connecting_badge')
})

const form = ref(emptyForm())
const timeshiftOffsetSeconds = ref(0)
const selectedProgram = computed(() => {
  if (!selectedProgramKey.value) return null
  return programs.value.find((p) => programKey(p) === selectedProgramKey.value) || null
})
const timeshiftMaxSeconds = computed(() => {
  if (!selectedProgram.value) return 0
  const start = new Date(selectedProgram.value.start_at).getTime()
  const end = new Date(selectedProgram.value.end_at).getTime()
  const duration = Math.floor((end - start) / 1000)
  return duration > 0 ? duration : 0
})
const timeshiftOffsetLabel = computed(() => formatOffset(timeshiftOffsetSeconds.value))
const timeshiftMaxLabel = computed(() => formatOffset(timeshiftMaxSeconds.value))
const selectedProgramIndex = computed(() => {
  if (!selectedProgramKey.value) return -1
  return programs.value.findIndex((p) => programKey(p) === selectedProgramKey.value)
})
const canSeekArchive = computed(() => archiveSupported.value && !!selectedProgram.value)
const canPrevProgram = computed(() => {
  if (archiveSupported.value && selectedProgram.value && isCurrentProgram(selectedProgram.value)) {
    return timeshiftOffsetSeconds.value > 0
  }
  return findPrevProgramIndex() >= 0
})
const canNextProgram = computed(() => {
  if (archiveSupported.value && selectedProgram.value && isCurrentProgram(selectedProgram.value)) {
    const liveEdge = currentLiveEdgeOffset(selectedProgram.value)
    return timeshiftOffsetSeconds.value < liveEdge - 1
  }
  return findNextProgramIndex() >= 0
})
const selectedProgramProgressLabel = computed(() => {
  if (!selectedProgram.value) return ''
  const elapsed = archiveSupported.value
    ? Math.min(Math.max(0, Math.floor(timeshiftOffsetSeconds.value)), timeshiftMaxSeconds.value)
    : Math.min(currentOffsetForProgram(selectedProgram.value), timeshiftMaxSeconds.value)
  return `${formatOffset(elapsed)} / ${timeshiftMaxLabel.value}`
})
const currentVideoFitMode = computed(() => {
  const key = currentVideoFitStorageKey()
  if (!key) return 'contain'
  return videoFitByChannel.value[key] || 'contain'
})
const videoFitModeLabel = computed(() => {
  if (currentVideoFitMode.value === 'cover') return 'Заполнить'
  if (currentVideoFitMode.value === 'fill') return 'Растянуть'
  return 'Вписать'
})
const playlistNameById = computed(() => Object.fromEntries(playlists.value.map((p) => [p.id, p.name])))
const favoriteChannels = computed(() => {
  const unique = new Map()
  for (const fav of favorites.value) {
    unique.set(favoriteStorageKey(fav), { ...fav })
  }
  for (const c of channels.value) {
    if (isFavorite(c.playlist_id, c.id)) {
      unique.set(favoriteStorageKey(c), { ...c })
    }
  }
  return [...unique.values()].filter((c) => {
    if (isHiddenGroup(c.group || 'Без группы')) return false
    if (isHiddenChannel(c.playlist_id, c.id, c.external_id)) return false
    return matchesChannelSearch(c.name || '')
  })
})
const settingsGroupedChannels = computed(() => buildSettingsGroupedChannels(channels.value))
const groupedChannels = computed(() => buildGroupedChannels(channels.value, {
  isHiddenChannel,
  isHiddenGroup,
  matchesChannelSearch
}))
const playlistVisibleCountById = computed(() => buildPlaylistVisibleCountById(
  playlists.value,
  visibleCountByPlaylist.value
))

function matchesChannelSearch(name) {
  return fuzzyScore(name, channelSearchQuery.value) > 0
}

function markFavoriteLogoError(channel) {
  markLogoError(channel)
}

function setChannelSearchQuery(value) {
  channelSearchQuery.value = String(value || '')
}

const OPEN_GROUPS_STORAGE_KEY = 'webtv_open_groups_by_playlist_v1'
const SETTINGS_OPEN_GROUPS_STORAGE_KEY = 'webtv_settings_open_groups_by_playlist_v1'
const AUDIO_TRACK_PREFS_STORAGE_KEY = 'webtv_audio_track_by_channel_v1'
const DEFAULT_AUDIO_LANGUAGE_STORAGE_KEY = 'webtv_default_audio_language_v1'

function saveOpenGroupsForPlaylist() {
  setGroupStateForPlaylist(OPEN_GROUPS_STORAGE_KEY, selectedPlaylist.value, openGroups.value)
}

function saveSettingsOpenGroupsForPlaylist() {
  setGroupStateForPlaylist(SETTINGS_OPEN_GROUPS_STORAGE_KEY, selectedPlaylist.value, settingsOpenGroups.value)
}

function videoFitChannelKey(playlistID, channelID) {
  return `${playlistID}:${channelID}`
}

function audioTrackChannelKey(channelID = selectedChannel.value) {
  if (!channelID) return ''
  const selected = channels.value.find((c) => c.id === channelID)
  if (selected?.playlist_id) {
    if (selected.external_id) {
      return `${selected.playlist_id}:ext:${selected.external_id}`
    }
    return `${selected.playlist_id}:id:${selected.id}`
  }
  if (selectedPlaylist.value) {
    return `${selectedPlaylist.value}:id:${channelID}`
  }
  return `ch:${channelID}`
}

function currentVideoFitStorageKey() {
  if (!selectedChannel.value) return ''
  const selected = channels.value.find((c) => c.id === selectedChannel.value)
  if (selected?.playlist_id) {
    return videoFitChannelKey(selected.playlist_id, selected.id)
  }
  if (selectedPlaylist.value) {
    return videoFitChannelKey(selectedPlaylist.value, selectedChannel.value)
  }
  return `ch:${selectedChannel.value}`
}

function loadFavorites() {
  try {
    const raw = localStorage.getItem('webtv_favorites_v1')
    const parsed = raw ? JSON.parse(raw) : []
    if (!Array.isArray(parsed)) return
    favorites.value = parsed
      .filter((item) => item && item.playlist_id && (item.id || item.external_id))
      .map((item) => ({
        ...item,
        external_id: item.external_id || ''
      }))
  } catch {
    favorites.value = []
  }
}

function saveFavorites() {
  localStorage.setItem('webtv_favorites_v1', JSON.stringify(favorites.value))
}

async function loadHiddenItems() {
  if (!selectedPlaylist.value) {
    hiddenGroups.value = []
    hiddenChannels.value = []
    return
  }
  const data = await api(`/api/playlists/${selectedPlaylist.value}/hidden`)
  hiddenGroups.value = Array.isArray(data?.groups) ? data.groups : []
  hiddenChannels.value = Array.isArray(data?.channels) ? data.channels : []
}

async function saveHiddenItems() {
  if (!selectedPlaylist.value) return
  await api(`/api/playlists/${selectedPlaylist.value}/hidden`, {
    method: 'PUT',
    body: JSON.stringify({
      groups: hiddenGroups.value,
      channels: hiddenChannels.value
    })
  })
}

function loadVideoFitPrefs() {
  try {
    const raw = localStorage.getItem('webtv_video_fit_by_channel_v1')
    const parsed = raw ? JSON.parse(raw) : {}
    videoFitByChannel.value = parsed && typeof parsed === 'object' ? parsed : {}
  } catch {
    videoFitByChannel.value = {}
  }
}

function saveVideoFitPrefs() {
  localStorage.setItem('webtv_video_fit_by_channel_v1', JSON.stringify(videoFitByChannel.value))
}

function loadAudioTrackPrefs() {
  try {
    const raw = localStorage.getItem(AUDIO_TRACK_PREFS_STORAGE_KEY)
    const parsed = raw ? JSON.parse(raw) : {}
    audioTrackByChannel.value = parsed && typeof parsed === 'object' ? parsed : {}
  } catch {
    audioTrackByChannel.value = {}
  }
}

function saveAudioTrackPrefs() {
  localStorage.setItem(AUDIO_TRACK_PREFS_STORAGE_KEY, JSON.stringify(audioTrackByChannel.value))
}

function loadDefaultAudioLanguage() {
  const raw = localStorage.getItem(DEFAULT_AUDIO_LANGUAGE_STORAGE_KEY)
  const allowed = new Set(defaultAudioLanguageOptions.map((opt) => opt.value))
  defaultAudioLanguage.value = allowed.has(raw) ? raw : 'auto'
}

function setDefaultAudioLanguage(lang) {
  const normalized = String(lang || '').trim().toLowerCase()
  const allowed = new Set(defaultAudioLanguageOptions.map((opt) => opt.value))
  const next = allowed.has(normalized) ? normalized : 'auto'
  defaultAudioLanguage.value = next
  localStorage.setItem(DEFAULT_AUDIO_LANGUAGE_STORAGE_KEY, next)
  if (audioTracks.value.length > 0) {
    // Re-evaluate preferred track for current channel immediately.
    setAudioTracks(audioTracks.value)
  }
}

function setAudioTrackPreference(trackIndex, channelID = selectedChannel.value) {
  const key = audioTrackChannelKey(channelID)
  if (!key) return
  const selected = channels.value.find((c) => c.id === channelID)
  const keys = [key]
  if (selected?.playlist_id) {
    if (selected.external_id) {
      keys.push(`${selected.playlist_id}:ext:${selected.external_id}`)
    }
    keys.push(`${selected.playlist_id}:id:${selected.id}`)
  }
  if (!Number.isInteger(trackIndex) || trackIndex < 0) {
    const next = { ...audioTrackByChannel.value }
    let changed = false
    for (const k of keys) {
      if (Object.prototype.hasOwnProperty.call(next, k)) {
        delete next[k]
        changed = true
      }
    }
    if (changed) {
      audioTrackByChannel.value = next
      saveAudioTrackPrefs()
    }
    return
  }
  const next = { ...audioTrackByChannel.value }
  let changed = false
  for (const k of keys) {
    if (next[k] !== trackIndex) {
      next[k] = trackIndex
      changed = true
    }
  }
  if (changed) {
    audioTrackByChannel.value = next
    saveAudioTrackPrefs()
  }
}

function preferredAudioTrackForChannel(channelID = selectedChannel.value) {
  const selected = channels.value.find((c) => c.id === channelID)
  const keys = []
  const key = audioTrackChannelKey(channelID)
  if (key) keys.push(key)
  if (selected?.playlist_id) {
    if (selected.external_id) {
      keys.push(`${selected.playlist_id}:ext:${selected.external_id}`)
    }
    keys.push(`${selected.playlist_id}:id:${selected.id}`)
  }
  for (const k of keys) {
    const val = Number(audioTrackByChannel.value[k])
    if (Number.isInteger(val) && val >= 0) return val
  }
  return -1
}

function normalizeAudioLanguageCode(v) {
  const raw = String(v || '').trim().toLowerCase()
  if (!raw) return ''
  const compact = raw.replace(/[^a-z]/g, '')
  const aliases = {
    ru: 'ru',
    rus: 'ru',
    russian: 'ru',
    en: 'en',
    eng: 'en',
    english: 'en',
    ro: 'ro',
    ron: 'ro',
    rum: 'ro',
    uk: 'uk',
    ukr: 'uk',
    ukrainian: 'uk',
    de: 'de',
    deu: 'de',
    ger: 'de',
    german: 'de',
    fr: 'fr',
    fra: 'fr',
    fre: 'fr',
    french: 'fr',
    es: 'es',
    spa: 'es',
    spanish: 'es',
    it: 'it',
    ita: 'it',
    italian: 'it',
    pt: 'pt',
    por: 'pt',
    portuguese: 'pt',
    pl: 'pl',
    pol: 'pl',
    polish: 'pl',
    tr: 'tr',
    tur: 'tr',
    turkish: 'tr',
    ar: 'ar',
    ara: 'ar',
    arabic: 'ar',
    zh: 'zh',
    zho: 'zh',
    chi: 'zh',
    chinese: 'zh',
    ja: 'ja',
    jpn: 'ja',
    japanese: 'ja',
    ko: 'ko',
    kor: 'ko',
    korean: 'ko'
  }
  if (aliases[compact]) return aliases[compact]
  if (compact.startsWith('ru')) return 'ru'
  if (compact.startsWith('en')) return 'en'
  return raw
}

function preferredAudioTrackByLanguage(tracks) {
  const target = defaultAudioLanguage.value
  if (target === 'auto') return -1
  const match = tracks.find((track) => normalizeAudioLanguageCode(track.language) === target)
  return Number.isInteger(match?.index) ? match.index : -1
}

function pruneVideoFitPrefsByPlaylists() {
  const existingPlaylists = new Set(playlists.value.map((p) => String(p.id)))
  const next = {}
  for (const [key, value] of Object.entries(videoFitByChannel.value)) {
    const [playlistID] = key.split(':')
    if (!existingPlaylists.has(playlistID)) continue
    if (!videoFitModes.includes(value)) continue
    next[key] = value
  }
  if (Object.keys(next).length !== Object.keys(videoFitByChannel.value).length) {
    videoFitByChannel.value = next
    saveVideoFitPrefs()
  }
}

function applyChannelVideoFit() {
  if (!video.value) return
  video.value.style.objectFit = currentVideoFitMode.value
}

function cycleVideoFitMode() {
  const key = currentVideoFitStorageKey()
  if (!key) return
  const current = videoFitByChannel.value[key] || 'contain'
  const idx = videoFitModes.indexOf(current)
  const nextMode = videoFitModes[(idx + 1) % videoFitModes.length]
  videoFitByChannel.value = {
    ...videoFitByChannel.value,
    [key]: nextMode
  }
  saveVideoFitPrefs()
  applyChannelVideoFit()
}

function pruneHiddenItemsByPlaylists() {
  // Hidden state is stored per playlist on server.
}

function pruneFavoritesByPlaylists() {
  const existing = new Set(playlists.value.map((p) => p.id))
  const next = favorites.value.filter((item) => existing.has(item.playlist_id))
  if (next.length !== favorites.value.length) {
    favorites.value = next
    saveFavorites()
  }
}

function isFavorite(playlistID, channelID) {
  const channel = channels.value.find((c) => c.playlist_id === playlistID && c.id === channelID)
  if (channel?.external_id) {
    return favorites.value.some((item) => item.playlist_id === playlistID && item.external_id === channel.external_id)
  }
  const key = favoriteKey(playlistID, channelID)
  return favorites.value.some((item) => favoriteKey(item.playlist_id, item.id) === key)
}

function toggleFavorite(channel) {
  const idx = favorites.value.findIndex((item) => {
    if (channel.external_id && item.external_id) {
      return item.playlist_id === channel.playlist_id && item.external_id === channel.external_id
    }
    return favoriteKey(item.playlist_id, item.id) === favoriteKey(channel.playlist_id, channel.id)
  })
  if (idx >= 0) {
    favorites.value.splice(idx, 1)
    saveFavorites()
    return
  }
  favorites.value.push({
    id: channel.id,
    playlist_id: channel.playlist_id,
    external_id: channel.external_id || '',
    name: channel.name,
    group: channel.group || '',
    logo: channel.logo || '',
    archive_supported: !!channel.archive_supported
  })
  saveFavorites()
  void loadFavoriteNowPrograms()
}

async function pickFavorite(channel) {
  if (selectedPlaylist.value !== channel.playlist_id) {
    selectedPlaylist.value = channel.playlist_id
    localStorage.setItem('webtv_last_playlist', channel.playlist_id)
    await loadChannels()
  }
  const target = findChannelByFavorite(channel, channels.value)
  if (!target) {
    window.alert(t('channel_not_found'))
    favorites.value = favorites.value.filter((item) => favoriteStorageKey(item) !== favoriteStorageKey(channel))
    saveFavorites()
    return
  }
  await pickChannel(target.id)
}

function favoriteNowProgramTitle(channel) {
  const resolved = findChannelByFavorite(channel, channels.value) || channel
  const key = favoriteKey(resolved.playlist_id, resolved.id)
  return favoriteNowProgramByKey.value[key]?.title || nowProgramByChannel.value[resolved.id]?.title || t('no_current_program')
}

function isGroupOpen(name) {
  return openGroups.value[name] !== false
}

function toggleGroup(name) {
  openGroups.value[name] = !isGroupOpen(name)
  saveOpenGroupsForPlaylist()
}

function isSettingsGroupOpen(name) {
  return settingsOpenGroups.value[name] !== false
}

function toggleSettingsGroup(name) {
  settingsOpenGroups.value[name] = !isSettingsGroupOpen(name)
  saveSettingsOpenGroupsForPlaylist()
}

function isHiddenGroup(groupName) {
  return hiddenGroups.value.includes(groupName || 'Без группы')
}

async function toggleHiddenGroup(groupName) {
  const key = groupName || 'Без группы'
  if (isHiddenGroup(key)) {
    hiddenGroups.value = hiddenGroups.value.filter((name) => name !== key)
  } else {
    hiddenGroups.value.push(key)
  }
  await saveHiddenItems()
  updateVisibleCountForSelectedPlaylist()
}

function isHiddenChannel(_playlistID, _channelID, externalID = '') {
  return hiddenChannels.value.includes(externalID || '')
}

async function toggleHiddenChannel(channel) {
  const key = channel.external_id || ''
  if (!key) return
  if (hiddenChannels.value.includes(key)) {
    hiddenChannels.value = hiddenChannels.value.filter((id) => id !== key)
  } else {
    hiddenChannels.value.push(key)
  }
  await saveHiddenItems()
  updateVisibleCountForSelectedPlaylist()
}

function updateVisibleCountForSelectedPlaylist() {
  if (!selectedPlaylist.value) return
  let visible = 0
  for (const channel of channels.value) {
    if (isHiddenChannel(channel.playlist_id, channel.id, channel.external_id)) continue
    if (isHiddenGroup(channel.group || 'Без группы')) continue
    visible += 1
  }
  visibleCountByPlaylist.value = {
    ...visibleCountByPlaylist.value,
    [selectedPlaylist.value]: visible
  }
}

function emptyForm() {
  return {
    name: '',
    m3u_url: '',
    epg_url: '',
    update_interval_minutes: 1440,
    enabled: true
  }
}

async function fetchNowProgramForChannel(playlistID, channelID) {
  if (!playlistID || !channelID) return null
  try {
    const items = await api(`/api/playlists/${playlistID}/now-programs`)
    if (!Array.isArray(items)) return null
    return items.find((item) => item.channel_id === channelID) || null
  } catch {
    return null
  }
}

function isArchivePlayingProgram(program) {
  return archiveSupported.value && selectedProgramKey.value === programKey(program) && isPastProgram(program)
}

function programProgressPercent(program) {
  return calculateProgramProgressPercent(program, nowTickMs.value)
}

function currentOffsetForProgram(program) {
  return programCurrentOffset(program, nowTickMs.value)
}

function currentLiveEdgeOffset(program) {
  if (!program) return 0
  if (!isCurrentProgram(program)) return timeshiftMaxSeconds.value
  return Math.min(currentOffsetForProgram(program), timeshiftMaxSeconds.value)
}

function currentPlaybackOffsetSeconds() {
  if (!selectedProgram.value) return 0
  if (!archiveSupported.value) {
    return currentOffsetForProgram(selectedProgram.value)
  }
  if (timeshiftDragging.value) {
    return Math.max(0, Math.floor(timeshiftOffsetSeconds.value))
  }
  if (playbackAnchor.value.active) {
    const elapsed = Math.floor((nowTickMs.value - playbackAnchor.value.startedAtMs) / 1000)
    const offset = playbackAnchor.value.baseOffsetSeconds + (elapsed > 0 ? elapsed : 0)
    return Math.max(0, offset)
  }
  return Math.max(0, Math.floor(timeshiftOffsetSeconds.value))
}

function withStreamOptions(streamURL) {
  const params = new URLSearchParams()
  if (deinterlaceEnabled.value) {
    params.set('deinterlace', '1')
  }
  const audioParam = selectedAudioTrackParam()
  if (audioParam >= 0 && (streamMode.value === 'transcode' || audioForceTranscode.value)) {
    params.set('audio', String(audioParam))
    if (audioForceTranscode.value) {
      params.set('audio_fallback', '1')
    }
  }
  const query = params.toString()
  if (!query) return streamURL
  const sep = streamURL.includes('?') ? '&' : '?'
  return `${streamURL}${sep}${query}`
}

function defaultAudioTrackIndex() {
  if (!Array.isArray(audioTracks.value) || audioTracks.value.length === 0) return -1
  const preferred = audioTracks.value.find((track) => track.default)
  return Number.isInteger(preferred?.index) ? preferred.index : audioTracks.value[0].index
}

function selectedAudioTrackParam() {
  if (!Number.isInteger(selectedAudioTrack.value) || selectedAudioTrack.value < 0) {
    return -1
  }
  if (selectedAudioTrack.value === 0) {
    return -1
  }
  return selectedAudioTrack.value
}

function isAudioFallbackTrackSelected() {
  return Number.isInteger(selectedAudioTrack.value) && selectedAudioTrack.value > 0
}

function applySelectedAudioTrackToPlayback() {
  if (selectedAudioTrack.value < 0) return
  if (hls && shouldUseManagedHls(currentVideoSrc.value)) {
    if (Array.isArray(hls.audioTracks) && hls.audioTracks.length > 1) {
      if (hls.audioTrack !== selectedAudioTrack.value) {
        hls.audioTrack = selectedAudioTrack.value
      }
      return
    }
    // Managed HLS is active, but alternate tracks are not exposed.
    // Keep direct mode; do not force ffmpeg from audio selection.
    return
  } else if (shouldUseManagedHls(currentVideoSrc.value)) {
    // hls.js instance is not ready yet.
    return
  }
  if (streamMode.value !== 'transcode') return
  if (!video.value || !selectedChannel.value) return
  const targetURL = currentTargetStreamURL()
  if (!targetURL) return
  if (currentVideoSrc.value === targetURL) return
  const reloaded = setVideoSourceIfChanged(targetURL, { force: true })
  if (reloaded) {
    void video.value.play().catch(() => {})
  }
}

function streamURLForProgram(channelID, program, offsetSeconds = 0) {
  const base = `/api/channels/${channelID}/stream`
  if (!archiveSupported.value) {
    return withStreamOptions(base)
  }
  const offset = Math.max(0, Math.floor(offsetSeconds))
  const liveEdge = currentLiveEdgeOffset(program)
  if (isCurrentProgram(program) && offset >= Math.max(0, liveEdge-1)) {
    return withStreamOptions(base)
  }
  const startAt = new Date(program.start_at)
  const now = new Date()
  const endAt = new Date(program.end_at)
  const effectiveEnd = endAt > now ? now : endAt
  const shiftedStart = new Date(startAt.getTime() + offset * 1000)
  if (shiftedStart >= effectiveEnd) {
    return withStreamOptions(base)
  }
  const params = new URLSearchParams({
    start: shiftedStart.toISOString(),
    end: effectiveEnd.toISOString()
  })
  return withStreamOptions(`${base}?${params.toString()}`)
}

function isArchiveStreamURL(src) {
  if (!src) return false
  return src.includes('start=') || src.includes('end=')
}

function shouldUseManagedHls(nextSrc) {
  if (!nextSrc || !video.value) return false
  if (!Hls.isSupported()) return false
  if (audioForceTranscode.value) return false
  if (streamMode.value === 'transcode') return false
  return true
}

function destroyHls() {
  if (!hls) return
  hls.destroy()
  hls = null
  if (audioSwitchProbeTimer) {
    clearTimeout(audioSwitchProbeTimer)
    audioSwitchProbeTimer = null
  }
  hlsDebug.value.engine = 'native'
  hlsDebug.value.level = '-'
  hlsDebug.value.latency = '-'
  hlsDebug.value.state = 'idle'
}

function setAudioTracks(nextTracks) {
  const normalized = Array.isArray(nextTracks) ? nextTracks.filter((item) => Number.isInteger(item?.index)) : []
  audioTracks.value = normalized
  if (normalized.length === 0) {
    selectedAudioTrack.value = -1
    return
  }
  const preferredFromStorage = preferredAudioTrackForChannel()
  if (normalized.some((track) => track.index === preferredFromStorage)) {
    selectedAudioTrack.value = preferredFromStorage
    audioForceTranscode.value = isAudioFallbackTrackSelected()
    applySelectedAudioTrackToPlayback()
    return
  }
  const preferredByLang = preferredAudioTrackByLanguage(normalized)
  if (normalized.some((track) => track.index === preferredByLang)) {
    selectedAudioTrack.value = preferredByLang
    audioForceTranscode.value = isAudioFallbackTrackSelected()
    applySelectedAudioTrackToPlayback()
    return
  }
  if (!normalized.some((track) => track.index === selectedAudioTrack.value)) {
    const preferred = normalized.find((track) => track.default)
    selectedAudioTrack.value = preferred ? preferred.index : normalized[0].index
  }
  audioForceTranscode.value = isAudioFallbackTrackSelected()
  applySelectedAudioTrackToPlayback()
}

function updateAudioTracksFromHls(instance) {
  const tracks = Array.isArray(instance?.audioTracks) ? instance.audioTracks : []
  if (tracks.length === 0) {
    setAudioTracks([])
    return
  }
  const nextTracks = tracks.map((track, index) => {
    const parts = []
    if (typeof track?.name === 'string' && track.name.trim() !== '') {
      parts.push(track.name.trim())
    }
    if (typeof track?.lang === 'string' && track.lang.trim() !== '') {
      parts.push(track.lang.trim().toUpperCase())
    }
    return {
      index,
      default: !!track?.default,
      language: String(track?.lang || ''),
      label: parts.length > 0 ? parts.join(' · ') : `Track ${index + 1}`
    }
  })
  setAudioTracks(nextTracks)
  if (Number.isInteger(instance.audioTrack) && instance.audioTrack >= 0) {
    selectedAudioTrack.value = instance.audioTrack
  }
  applySelectedAudioTrackToPlayback()
}

async function loadAudioTracks() {
  if (!selectedChannel.value) {
    setAudioTracks([])
    return
  }
  if (hls && Array.isArray(hls.audioTracks) && hls.audioTracks.length > 1) {
    updateAudioTracksFromHls(hls)
    return
  }
  try {
    const list = await api(`/api/channels/${selectedChannel.value}/audio-tracks`)
    if (hls && Array.isArray(hls.audioTracks) && hls.audioTracks.length > 1) {
      updateAudioTracksFromHls(hls)
      return
    }
    if (!Array.isArray(list)) {
      setAudioTracks([])
      return
    }
    const nextTracks = list
      .filter((item) => Number.isInteger(item?.index))
      .map((item) => {
        const parts = []
        if (typeof item.title === 'string' && item.title.trim() !== '') {
          parts.push(item.title.trim())
        }
        if (typeof item.language === 'string' && item.language.trim() !== '') {
          parts.push(item.language.trim().toUpperCase())
        }
        if (typeof item.codec === 'string' && item.codec.trim() !== '') {
          parts.push(item.codec.trim().toUpperCase())
        }
        return {
          index: item.index,
          default: !!item.default,
          language: String(item.language || ''),
          label: parts.length > 0 ? parts.join(' · ') : `Track ${item.index + 1}`
        }
      })
    setAudioTracks(nextTracks)
  } catch {
    setAudioTracks([])
  }
}

function clearStartupSlowTimer() {
  if (!startupSlowTimer) return
  clearTimeout(startupSlowTimer)
  startupSlowTimer = null
}

function clearStalledRecoveryTimer() {
  if (!stalledRecoveryTimer) return
  clearTimeout(stalledRecoveryTimer)
  stalledRecoveryTimer = null
}

function beginStartupOverlay() {
  clearStartupSlowTimer()
  startupState.value = streamMode.value === 'transcode' ? 'transcoding' : 'connecting'
  startupSlowTimer = window.setTimeout(() => {
    if (startupState.value === 'connecting') {
      startupState.value = 'connecting_slow'
    }
  }, 3500)
}

function finishStartupOverlay() {
  clearStartupSlowTimer()
  startupState.value = 'idle'
}

function refreshStartupOverlayMode() {
  if (startupState.value === 'idle') return
  if (streamMode.value === 'transcode') {
    startupState.value = 'transcoding'
  } else if (startupState.value === 'transcoding') {
    startupState.value = 'connecting'
  }
}

function updateHlsDebugFromVideo() {
  if (!video.value) return
  const v = video.value
  const ct = Number(v.currentTime || 0)
  const ranges = v.buffered
  let bufferedEnd = ct
  for (let i = 0; i < ranges.length; i += 1) {
    const start = ranges.start(i)
    const end = ranges.end(i)
    if (start <= ct && ct <= end) {
      bufferedEnd = end
      break
    }
  }
  const bufferAhead = Math.max(0, bufferedEnd - ct)
  hlsDebug.value.buffer = `${bufferAhead.toFixed(1)}s`
}

function bindHlsDebugEvents(instance) {
  const tryApplyPendingAudioTrack = () => {
    const target = pendingAudioTrackApply.value
    if (!Number.isInteger(target) || target < 0) return
    if (!Array.isArray(instance.audioTracks) || instance.audioTracks.length <= target) return
    instance.audioTrack = target
    pendingAudioTrackApply.value = -1
    hlsDebug.value.audioResult = `hls_ok:${target}`
  }

  instance.on(Hls.Events.MANIFEST_LOADING, () => {
    hlsDebug.value.state = 'manifest_loading'
  })
  instance.on(Hls.Events.MANIFEST_PARSED, () => {
    hlsDebug.value.state = 'manifest_parsed'
    tryApplyPendingAudioTrack()
  })
  instance.on(Hls.Events.LEVEL_SWITCHED, (_evt, data) => {
    hlsDebug.value.level = String(data?.level ?? '-')
  })
  instance.on(Hls.Events.FRAG_LOADED, () => {
    hlsDebug.value.state = 'frag_loaded'
    updateHlsDebugFromVideo()
    const lat = instance.latency
    hlsDebug.value.latency = Number.isFinite(lat) ? `${lat.toFixed(1)}s` : '-'
  })
  instance.on(Hls.Events.AUDIO_TRACKS_UPDATED, () => {
    updateAudioTracksFromHls(instance)
    hlsDebug.value.audioTracks = Array.isArray(instance.audioTracks) ? instance.audioTracks.length : 0
    tryApplyPendingAudioTrack()
  })
  instance.on(Hls.Events.AUDIO_TRACK_SWITCHED, (_evt, data) => {
    if (Number.isInteger(data?.id) && data.id >= 0) {
      selectedAudioTrack.value = data.id
      if (pendingAudioTrackApply.value === data.id) {
        pendingAudioTrackApply.value = -1
      }
      hlsDebug.value.audioResult = `hls_ok:${data.id}`
    }
  })
  instance.on(Hls.Events.ERROR, (_evt, data) => {
    if (data?.details) {
      hlsDebug.value.lastError = String(data.details)
    }
    if (!data?.fatal) {
      hlsRetries += 1
      hlsDebug.value.retries = hlsRetries
      return
    }
    hlsDebug.value.state = 'fatal_error'
    if (!video.value) return
    // Fallback to native playback if hls.js encounters a fatal error.
    destroyHls()
    video.value.src = currentVideoSrc.value
    video.value.load()
  })
}

function currentTargetStreamURL() {
  if (!selectedChannel.value) return ''
  if (!selectedProgram.value) return withStreamOptions(`/api/channels/${selectedChannel.value}/stream`)
  return streamURLForProgram(selectedChannel.value, selectedProgram.value, currentPlaybackOffsetSeconds())
}

async function refreshStreamModeForCurrentChannel() {
  if (!selectedChannel.value) return
  try {
    const headURL = currentTargetStreamURL() || withStreamOptions(`/api/channels/${selectedChannel.value}/stream`)
    const res = await fetch(headURL, { method: 'HEAD' })
    const nextMode = res.headers.get('x-stream-mode') || ''
    streamMode.value = nextMode
    const videoPath = (res.headers.get('x-video-path') || '').trim()
    const audioPath = (res.headers.get('x-audio-path') || '').trim()
    if (videoPath !== '') hlsDebug.value.videoPath = videoPath
    if (audioPath !== '') hlsDebug.value.audioPath = audioPath
    refreshStartupOverlayMode()
  } catch {
    // Keep current mode on transient network errors.
  }
}

function setVideoSourceIfChanged(nextSrc, { force = false } = {}) {
  if (!video.value) return false
  if (!nextSrc) return false
  if (force) {
    const now = Date.now()
    if (now - lastForcedReloadAtMs < forceReloadCooldownMs) {
      return false
    }
    lastForcedReloadAtMs = now
  }
  if (!force && currentVideoSrc.value === nextSrc) {
    return false
  }
  beginStartupOverlay()
  currentVideoSrc.value = nextSrc
  if (shouldUseManagedHls(nextSrc)) {
    if (!hls) {
      hls = new Hls({
        lowLatencyMode: false,
        maxBufferLength: 20,
        backBufferLength: 30,
        liveSyncDurationCount: 3,
        liveMaxLatencyDurationCount: 8,
        manifestLoadingTimeOut: 8000,
        levelLoadingTimeOut: 8000,
        fragLoadingTimeOut: 15000,
        manifestLoadingMaxRetry: 2,
        levelLoadingMaxRetry: 2,
        fragLoadingMaxRetry: 2
      })
      hlsRetries = 0
      hlsDebug.value.retries = 0
      hlsDebug.value.lastError = ''
      hlsDebug.value.engine = 'hls.js'
      bindHlsDebugEvents(hls)
    }
    hls.attachMedia(video.value)
    hls.loadSource(nextSrc)
  } else {
    destroyHls()
    video.value.src = nextSrc
    video.value.load()
  }
  return true
}

function togglePlayPause() {
  if (!video.value) return
  if (video.value.paused) {
    void video.value.play()
  } else {
    video.value.pause()
  }
}

function onVideoMetadata() {
  if (!video.value) return
  const w = video.value.videoWidth || 0
  const h = video.value.videoHeight || 0
  if (w > 0 && h > 0) {
    playerAspectRatio.value = `${w} / ${h}`
  }
}

function scheduleControlsHide() {
  if (controlsHideTimer) {
    clearTimeout(controlsHideTimer)
    controlsHideTimer = null
  }
  if (!isPlaying.value) {
    controlsVisible.value = true
    return
  }
  controlsHideTimer = window.setTimeout(() => {
    controlsVisible.value = false
    controlsHideTimer = null
  }, 2000)
}

function revealControls() {
  controlsVisible.value = true
  scheduleControlsHide()
}

function onVideoPlay() {
  isPlaying.value = true
  clearStalledRecoveryTimer()
  finishStartupOverlay()
  scheduleControlsHide()
}

function onVideoPause() {
  isPlaying.value = false
  controlsVisible.value = true
  if (controlsHideTimer) {
    clearTimeout(controlsHideTimer)
    controlsHideTimer = null
  }
}

function toggleMute() {
  if (!video.value) return
  video.value.muted = !video.value.muted
  isMuted.value = video.value.muted
}

function onVolumeInput(event) {
  if (!video.value) return
  const level = Number(event.target.value)
  if (!Number.isFinite(level)) return
  const bounded = Math.min(1, Math.max(0, level))
  video.value.volume = bounded
  volumeLevel.value = bounded
  if (bounded > 0 && video.value.muted) {
    video.value.muted = false
  }
}

function onVideoVolumeChange() {
  if (!video.value) return
  isMuted.value = video.value.muted || video.value.volume === 0
  volumeLevel.value = video.value.volume
}

function onVideoTimeUpdate() {
  updateHlsDebugFromVideo()
}

function onVideoCanPlay() {
  hlsDebug.value.state = 'canplay'
  clearStalledRecoveryTimer()
  finishStartupOverlay()
  updateHlsDebugFromVideo()
}

function onVideoError() {
  recoverLiveStream('video_error')
}

function onVideoWaiting() {
  scheduleStalledRecovery()
}

function onVideoStalled() {
  scheduleStalledRecovery()
}

function scheduleStalledRecovery() {
  if (!selectedChannel.value || isPlaying.value === false) return
  if (stalledRecoveryTimer) return
  stalledRecoveryTimer = window.setTimeout(() => {
    stalledRecoveryTimer = null
    recoverLiveStream('stall_timeout')
  }, 5000)
}

function recoverLiveStream(reason) {
  finishStartupOverlay()
  clearStalledRecoveryTimer()
  if (!selectedChannel.value || !video.value) return
  const liveURL = withStreamOptions(`/api/channels/${selectedChannel.value}/stream`)
  const reloaded = setVideoSourceIfChanged(liveURL, { force: true })
  if (!reloaded) return
  hlsDebug.value.state = `recover:${reason}`
  void video.value.play().catch(() => {})
}

function toggleDeinterlace() {
  deinterlaceEnabled.value = !deinterlaceEnabled.value
  localStorage.setItem('webtv_deinterlace', deinterlaceEnabled.value ? '1' : '0')
  if (!selectedChannel.value) return
  const targetURL = currentTargetStreamURL()
  const reloaded = setVideoSourceIfChanged(targetURL, { force: true })
  if (reloaded && video.value) {
    void video.value.play().catch(() => {})
  }
}

async function onSelectAudioTrack(event) {
  const next = Number(event?.target?.value)
  if (!Number.isInteger(next) || next < 0) return
  selectedAudioTrack.value = next
  setAudioTrackPreference(next)
  hlsDebug.value.audioTarget = String(next)
  hlsDebug.value.audioResult = 'pending'

  audioForceTranscode.value = isAudioFallbackTrackSelected()
  await refreshStreamModeForCurrentChannel()

  // Track 0: always prefer direct/hls.js path.
  if (!isAudioFallbackTrackSelected()) {
    audioForceTranscode.value = false
    const targetURL = currentTargetStreamURL()
    pendingAudioTrackApply.value = next
    if (targetURL) {
      if (streamMode.value === 'transcode') {
        streamMode.value = 'direct'
      }
      const reloaded = setVideoSourceIfChanged(targetURL, { force: true })
      if (reloaded && video.value) {
        void video.value.play().catch(() => {})
      }
      window.setTimeout(() => {
        applySelectedAudioTrackToPlayback()
      }, 220)
    }
    hlsDebug.value.audioResult = 'hls_direct_track0'
    return
  }

  // Track > 0: fallback to ffmpeg/transcode.
  audioForceTranscode.value = true
  streamMode.value = 'transcode'
  refreshStartupOverlayMode()
  hlsDebug.value.audioResult = 'ffmpeg_fallback_track>0'
  if (streamMode.value !== 'transcode') return
  if (!video.value) return
  const targetURL = currentTargetStreamURL()
  const reloaded = setVideoSourceIfChanged(targetURL, { force: true })
  if (reloaded) {
    void video.value.play().catch(() => {})
  }
}

function toggleDebugOverlay() {
  showDebugOverlay.value = !showDebugOverlay.value
  localStorage.setItem('webtv_hls_debug_overlay', showDebugOverlay.value ? '1' : '0')
}

function updateMediaPathsDebug() {
  if (streamMode.value === 'direct') {
    hlsDebug.value.videoPath = 'bypass'
    hlsDebug.value.audioPath = 'bypass'
    return
  }
  if (streamMode.value === 'transcode') {
    // Current backend transcode pipeline re-encodes both audio and video.
    hlsDebug.value.videoPath = 'ffmpeg'
    hlsDebug.value.audioPath = 'ffmpeg'
    return
  }
  hlsDebug.value.videoPath = '-'
  hlsDebug.value.audioPath = '-'
}

function debugServerModeLabel(mode) {
  if (mode === 'transcode') return 'transcode (ffmpeg)'
  if (mode === 'direct') return 'direct'
  return '-'
}

function resolveSystemTheme() {
  if (window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches) {
    return 'dark'
  }
  return 'light'
}

function applyTheme(mode) {
  if (mode === 'dark') {
    document.documentElement.setAttribute('data-theme', 'dark')
    return
  }
  if (mode === 'light') {
    document.documentElement.setAttribute('data-theme', 'light')
    return
  }
  document.documentElement.setAttribute('data-theme', resolveSystemTheme())
}

function onSystemThemeChange() {
  if (themeMode.value !== 'system') return
  applyTheme('system')
}

function onThemeModeChange() {
  const nextMode = ['light', 'dark', 'system'].includes(themeMode.value) ? themeMode.value : 'system'
  themeMode.value = nextMode
  localStorage.setItem('webtv_theme_mode', nextMode)
  applyTheme(nextMode)
}

function setThemeMode(mode) {
  themeMode.value = mode
  onThemeModeChange()
}

function setLanguageMode(mode) {
  uiLanguageMode.value = mode
  onLanguageModeChange()
}

function toggleFullscreen() {
  if (!playerWrap.value) return
  if (document.fullscreenElement) {
    void document.exitFullscreen()
    return
  }
  void playerWrap.value.requestFullscreen?.()
}

async function detachPlayer() {
  if (!video.value || !pictureInPictureSupported.value) return
  try {
    if (document.pictureInPictureElement) {
      await document.exitPictureInPicture()
      return
    }
    await video.value.requestPictureInPicture?.()
  } catch {
    // Ignore user-agent errors (for example, blocked by browser policy).
  }
}

function onEnterPictureInPicture() {
  isInPictureInPicture.value = true
}

function onLeavePictureInPicture() {
  isInPictureInPicture.value = false
}

function startPlaybackTracking(offsetSeconds) {
  playbackAnchor.value = {
    active: true,
    baseOffsetSeconds: Math.max(0, Math.floor(offsetSeconds)),
    startedAtMs: Date.now()
  }
}

function stopPlaybackTracking() {
  playbackAnchor.value = {
    active: false,
    baseOffsetSeconds: 0,
    startedAtMs: 0
  }
}

function syncTimeshiftFromPlayback() {
  if (!archiveSupported.value || !selectedProgram.value || timeshiftDragging.value) return
  if (!playbackAnchor.value.active) return

  const elapsed = Math.floor((Date.now() - playbackAnchor.value.startedAtMs) / 1000)
  const offset = playbackAnchor.value.baseOffsetSeconds + (elapsed > 0 ? elapsed : 0)
  timeshiftOffsetSeconds.value = Math.min(offset, currentLiveEdgeOffset(selectedProgram.value))
}

function isSelectableArchiveProgram(program) {
  return archiveSupported.value && !isFutureProgram(program)
}

function findPrevProgramIndex() {
  if (!archiveSupported.value || selectedProgramIndex.value <= 0) return -1
  for (let i = selectedProgramIndex.value - 1; i >= 0; i -= 1) {
    if (isSelectableArchiveProgram(programs.value[i])) return i
  }
  return -1
}

function findNextProgramIndex() {
  if (!archiveSupported.value || selectedProgramIndex.value < 0) return -1
  for (let i = selectedProgramIndex.value + 1; i < programs.value.length; i += 1) {
    if (isSelectableArchiveProgram(programs.value[i])) return i
  }
  return -1
}

async function loadPlaylists() {
  playlists.value = asArray(await api('/api/playlists'))
  const existing = new Set(playlists.value.map((p) => p.id))
  visibleCountByPlaylist.value = Object.fromEntries(
    Object.entries(visibleCountByPlaylist.value).filter(([id]) => existing.has(Number(id)))
  )
  pruneFavoritesByPlaylists()
  pruneHiddenItemsByPlaylists()
  pruneVideoFitPrefsByPlaylists()
  void loadFavoriteNowPrograms()
}

async function loadAppConfig() {
  try {
    const cfg = await api('/api/config')
    const nextTitle = typeof cfg?.app_title === 'string' ? cfg.app_title.trim() : ''
    appTitle.value = nextTitle || 'WebTV'
  } catch {
    appTitle.value = 'WebTV'
  }
}

async function savePlaylist() {
  if (editingId.value) {
    await api(`/api/playlists/${editingId.value}`, { method: 'PUT', body: JSON.stringify(form.value) })
  } else {
    await api('/api/playlists', { method: 'POST', body: JSON.stringify(form.value) })
  }
  cancelEdit()
  await loadPlaylists()
}

function startEdit(playlist) {
  editingId.value = playlist.id
  form.value = {
    name: playlist.name,
    m3u_url: playlist.m3u_url,
    epg_url: playlist.epg_url,
    update_interval_minutes: playlist.update_interval_minutes,
    enabled: playlist.enabled
  }
}

function cancelEdit() {
  editingId.value = 0
  form.value = emptyForm()
}

async function removePlaylist(playlist) {
  const ok = window.confirm(t('confirm_delete_playlist').replace('{name}', playlist.name))
  if (!ok) return

  await api(`/api/playlists/${playlist.id}`, { method: 'DELETE' })
  favorites.value = favorites.value.filter((item) => item.playlist_id !== playlist.id)
  saveFavorites()
  favoriteNowProgramByKey.value = Object.fromEntries(
    Object.entries(favoriteNowProgramByKey.value).filter(([key]) => !key.startsWith(`${playlist.id}:`))
  )

  if (editingId.value === playlist.id) {
    cancelEdit()
  }

  if (selectedPlaylist.value === playlist.id) {
    selectedPlaylist.value = 0
    selectedChannel.value = 0
    channels.value = []
    hiddenGroups.value = []
    hiddenChannels.value = []
    programs.value = []
    nowProgramByChannel.value = {}
    streamMode.value = ''
    archiveSupported.value = false
    if (video.value) {
      currentVideoSrc.value = ''
      video.value.removeAttribute('src')
      video.value.load()
    }
  }

  await loadPlaylists()
}

async function refresh(id) {
  if (refreshingPlaylistIds.value[id]) return
  refreshingPlaylistIds.value[id] = true
  try {
    await api(`/api/playlists/${id}/refresh`, { method: 'POST' })
    await loadPlaylists()
  } finally {
    refreshingPlaylistIds.value[id] = false
  }
}

async function selectPlaylist(id) {
  if (selectedPlaylist.value === id) return
  selectedPlaylist.value = id
  localStorage.setItem('webtv_last_playlist', id)
  await loadChannels()
}

async function loadChannels() {
  nowProgramsRequestToken += 1
  const token = nowProgramsRequestToken
  const pageSize = 250
  const prevSelected = channels.value.find((c) => c.id === selectedChannel.value)
  const prevSelectedExternalID = prevSelected?.external_id || localStorage.getItem('webtv_last_channel_external_id') || ''
  nowProgramByChannel.value = {}

  if (!selectedPlaylist.value) {
    channels.value = []
    openGroups.value = {}
    return
  }

  const firstPage = asArray(await api(`/api/playlists/${selectedPlaylist.value}/channels?limit=${pageSize}&offset=0`))
  if (token !== nowProgramsRequestToken) return
  channels.value = firstPage
  if (selectedChannel.value && !channels.value.some((c) => c.id === selectedChannel.value) && prevSelectedExternalID) {
    const replacement = channels.value.find((c) => c.external_id === prevSelectedExternalID)
    if (replacement) {
      selectedChannel.value = replacement.id
      localStorage.setItem('webtv_last_channel', replacement.id)
    }
  }
  await loadHiddenItems()
  updateVisibleCountForSelectedPlaylist()
  const groupsState = {}
  for (const channel of channels.value) {
    const groupName = channel.group || 'Без группы'
    if (groupsState[groupName] === undefined) {
      groupsState[groupName] = false
    }
  }
  const savedOpenGroups = getGroupStateForPlaylist(OPEN_GROUPS_STORAGE_KEY, selectedPlaylist.value)
  const savedSettingsOpenGroups = getGroupStateForPlaylist(SETTINGS_OPEN_GROUPS_STORAGE_KEY, selectedPlaylist.value)
  openGroups.value = mergeGroupState(groupsState, savedOpenGroups, false)
  settingsOpenGroups.value = mergeGroupState(groupsState, savedSettingsOpenGroups, true)
  saveOpenGroupsForPlaylist()
  saveSettingsOpenGroupsForPlaylist()

  loadNowProgramsInBackground(selectedPlaylist.value, token)
  if (firstPage.length === pageSize) {
    void loadRemainingChannelsInBackground(selectedPlaylist.value, pageSize, pageSize, token)
  }
}

async function loadRemainingChannelsInBackground(playlistID, offset, pageSize, token) {
  let nextOffset = offset
  try {
    while (token === nowProgramsRequestToken) {
      const page = asArray(await api(`/api/playlists/${playlistID}/channels?limit=${pageSize}&offset=${nextOffset}`))
      if (token !== nowProgramsRequestToken) return
      if (!Array.isArray(page) || page.length === 0) return
      channels.value = channels.value.concat(page)
      updateVisibleCountForSelectedPlaylist()
      if (page.length < pageSize) return
      nextOffset += pageSize
    }
  } catch {
    // Ignore background pagination errors to keep the UI responsive.
  }
}

async function loadNowProgramsInBackground(playlistID, token) {
  if (!playlistID) return
  try {
    const items = asArray(await api(`/api/playlists/${playlistID}/now-programs`))
    if (token !== nowProgramsRequestToken) return
    const nextNowByChannel = {}
    const nextFavoriteNowByKey = { ...favoriteNowProgramByKey.value }
    for (const now of items) {
      nextNowByChannel[now.channel_id] = now
      nextFavoriteNowByKey[favoriteKey(playlistID, now.channel_id)] = now
    }
    nowProgramByChannel.value = nextNowByChannel
    favoriteNowProgramByKey.value = nextFavoriteNowByKey
  } catch {
    // Ignore list-view EPG errors to keep channel list usable.
  }
}

async function loadFavoriteNowPrograms() {
  const playlistIDs = [...new Set(favorites.value.map((item) => item.playlist_id))]
    .filter((id) => playlists.value.some((p) => p.id === id))
  favoriteNowProgramsRequestToken += 1
  const token = favoriteNowProgramsRequestToken
  if (playlistIDs.length === 0) {
    favoriteNowProgramByKey.value = {}
    return
  }

  const next = {}
  await Promise.all(playlistIDs.map(async (playlistID) => {
    try {
      const items = asArray(await api(`/api/playlists/${playlistID}/now-programs`))
      if (token !== favoriteNowProgramsRequestToken) return
      for (const now of items) {
        next[favoriteKey(playlistID, now.channel_id)] = now
      }
    } catch {
      // Ignore errors for one playlist; keep favorites list usable.
    }
  }))
  if (token !== favoriteNowProgramsRequestToken) return
  favoriteNowProgramByKey.value = next
}

async function pickChannel(channelID) {
  selectedChannel.value = channelID
  setAudioTracks([])
  selectedAudioTrack.value = preferredAudioTrackForChannel(channelID)
  localStorage.setItem('webtv_last_channel', channelID)
  const selected = channels.value.find((c) => c.id === channelID)
  localStorage.setItem('webtv_last_channel_external_id', selected?.external_id || '')
  applyChannelVideoFit()
  await selectChannel()
}

async function selectChannel() {
  if (!selectedChannel.value) return
  channelSelectRequestToken += 1
  const token = channelSelectRequestToken

  // Clear previous channel EPG selection immediately to avoid stale
  // progress badge/timeshift controls while new EPG is loading.
  programs.value = []
  selectedProgramKey.value = ''
  timeshiftOffsetSeconds.value = 0
  audioForceTranscode.value = isAudioFallbackTrackSelected()
  if (audioForceTranscode.value) {
    streamMode.value = 'transcode'
  }
  pendingAudioTrackApply.value = -1
  hlsDebug.value.audioTracks = 0
  hlsDebug.value.audioTarget = selectedAudioTrack.value >= 0 ? String(selectedAudioTrack.value) : '-'
  hlsDebug.value.audioResult = '-'
  stopPlaybackTracking()

  const c = channels.value.find((x) => x.id === selectedChannel.value)
  archiveSupported.value = !!c?.archive_supported
  const liveURL = withStreamOptions(`/api/channels/${selectedChannel.value}/stream`)
  setVideoSourceIfChanged(liveURL)
  void loadAudioTracks()

  const res = await fetch(liveURL, { method: 'HEAD' })
  if (token !== channelSelectRequestToken) return
  streamMode.value = res.headers.get('x-stream-mode') || ''
  const videoPath = (res.headers.get('x-video-path') || '').trim()
  const audioPath = (res.headers.get('x-audio-path') || '').trim()
  if (videoPath !== '') hlsDebug.value.videoPath = videoPath
  if (audioPath !== '') hlsDebug.value.audioPath = audioPath
  refreshStartupOverlayMode()

  const now = new Date()
  const from = new Date(now.getTime() + (archiveSupported.value ? -24 : 0) * 60 * 60 * 1000)
  const to = new Date(now.getTime() + 24 * 60 * 60 * 1000)
  const params = new URLSearchParams({
    from: from.toISOString(),
    to: to.toISOString()
  })
  programs.value = asArray(await api(`/api/channels/${selectedChannel.value}/epg?${params.toString()}`))
  if (token !== channelSelectRequestToken) return
  const freshNowHint = await fetchNowProgramForChannel(selectedPlaylist.value, selectedChannel.value)
  if (token !== channelSelectRequestToken) return
  if (freshNowHint) {
    nowProgramByChannel.value[selectedChannel.value] = freshNowHint
  }
  const nowHint = freshNowHint || nowProgramByChannel.value[selectedChannel.value]
  const nowProgram = findProgramByNowHint(programs.value, nowHint) || findNowProgram(programs.value) || findClosestProgramToNow(programs.value)
  if (nowProgram) {
    nowProgramByChannel.value[selectedChannel.value] = nowProgram
    selectedProgramKey.value = programKey(nowProgram)
    const liveOffset = currentLiveEdgeOffset(nowProgram)
    timeshiftOffsetSeconds.value = liveOffset
    startPlaybackTracking(liveOffset)
  } else if (programs.value.length > 0) {
    selectedProgramKey.value = programKey(programs.value[0])
    timeshiftOffsetSeconds.value = 0
    startPlaybackTracking(0)
  } else {
    selectedProgramKey.value = ''
    timeshiftOffsetSeconds.value = 0
    stopPlaybackTracking()
  }
}

function selectProgram(program) {
  if (isFutureProgram(program)) return
  selectedProgramKey.value = programKey(program)
  const initialOffset = isCurrentProgram(program)
    ? currentLiveEdgeOffset(program)
    : 0
  timeshiftOffsetSeconds.value = initialOffset
  if (!video.value || !selectedChannel.value) return

  const streamURL = streamURLForProgram(selectedChannel.value, program, timeshiftOffsetSeconds.value)
  setVideoSourceIfChanged(streamURL)
  startPlaybackTracking(initialOffset)
}

function selectPrevProgram() {
  if (archiveSupported.value && selectedProgram.value && isCurrentProgram(selectedProgram.value)) {
    timeshiftOffsetSeconds.value = 0
    applyTimeshift()
    return
  }
  const idx = findPrevProgramIndex()
  if (idx < 0) return
  selectProgram(programs.value[idx])
}

function selectNextProgram() {
  if (archiveSupported.value && selectedProgram.value && isCurrentProgram(selectedProgram.value)) {
    const liveEdge = currentLiveEdgeOffset(selectedProgram.value)
    timeshiftOffsetSeconds.value = liveEdge
    applyTimeshift()
    return
  }
  const idx = findNextProgramIndex()
  if (idx < 0) return
  selectProgram(programs.value[idx])
}

function onTimeshiftInput(event) {
  const offset = Number(event.target.value)
  if (!Number.isFinite(offset)) return
  timeshiftDragging.value = true
  const cap = currentLiveEdgeOffset(selectedProgram.value)
  timeshiftOffsetSeconds.value = Math.min(Math.max(0, Math.floor(offset)), cap)
}

function applyTimeshift() {
  timeshiftDragging.value = false
  if (!selectedProgram.value || !video.value || !selectedChannel.value) return
  const streamURL = streamURLForProgram(selectedChannel.value, selectedProgram.value, timeshiftOffsetSeconds.value)
  setVideoSourceIfChanged(streamURL)
  startPlaybackTracking(timeshiftOffsetSeconds.value)
}

function seekBySeconds(delta) {
  if (!selectedProgram.value || !canSeekArchive.value) return
  const cap = currentLiveEdgeOffset(selectedProgram.value)
  const nextOffset = Math.min(
    cap,
    Math.max(0, timeshiftOffsetSeconds.value + delta)
  )
  timeshiftOffsetSeconds.value = nextOffset
  applyTimeshift()
}

function onVideoEnded() {
  if (!selectedChannel.value || !video.value) return

  const liveURL = withStreamOptions(`/api/channels/${selectedChannel.value}/stream`)
  const src = currentVideoSrc.value || video.value.currentSrc || ''
  const archiveStream = isArchiveStreamURL(src)

  // If browser thinks stream has ended while playing live URL, force reconnect.
  if (!archiveStream) {
    const reloaded = setVideoSourceIfChanged(liveURL, { force: true })
    if (reloaded) {
      void video.value.play().catch(() => {})
    }
    return
  }

  if (!selectedProgram.value || !archiveSupported.value) return

  // If archived playback reached the end of a finished program,
  // jump to the next available non-future program automatically.
  if (!isCurrentProgram(selectedProgram.value)) {
    const nextIdx = findNextProgramIndex()
    if (nextIdx >= 0) {
      selectProgram(programs.value[nextIdx])
      void video.value.play().catch(() => {})
      return
    }
    const reloaded = setVideoSourceIfChanged(liveURL, { force: true })
    if (reloaded) {
      void video.value.play().catch(() => {})
    }
    return
  }

  const liveEdge = currentLiveEdgeOffset(selectedProgram.value)
  if (timeshiftOffsetSeconds.value < liveEdge-2) return
  const reloaded = setVideoSourceIfChanged(liveURL, { force: true })
  if (reloaded) {
    startPlaybackTracking(liveEdge)
    void video.value.play().catch(() => {})
  }
}

watch(
  () => streamMode.value,
  (mode) => {
    hlsDebug.value.serverMode = debugServerModeLabel(mode)
    if (hlsDebug.value.videoPath === '-' || hlsDebug.value.audioPath === '-') {
      updateMediaPathsDebug()
    }
  },
  { immediate: true }
)

watch(
  () => selectedAudioTrack.value,
  () => {
    updateMediaPathsDebug()
  }
)

onMounted(async () => {
  showDebugOverlay.value = localStorage.getItem('webtv_hls_debug_overlay') === '1'
  deinterlaceEnabled.value = localStorage.getItem('webtv_deinterlace') === '1'
  const storedThemeMode = localStorage.getItem('webtv_theme_mode')
  themeMode.value = ['light', 'dark', 'system'].includes(storedThemeMode) ? storedThemeMode : 'system'
  applyTheme(themeMode.value)
  initLanguageMode()
  loadBrokenLogos()
  if (window.matchMedia) {
    systemThemeMediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
    systemThemeMediaQuery.addEventListener?.('change', onSystemThemeChange)
  }
  loadFavorites()
  loadVideoFitPrefs()
  loadAudioTrackPrefs()
  loadDefaultAudioLanguage()
  await loadAppConfig()
  await loadPlaylists()

  const lastPlaylistID = parseInt(localStorage.getItem('webtv_last_playlist') || '0', 10)
  const lastChannelID = parseInt(localStorage.getItem('webtv_last_channel') || '0', 10)
  const lastChannelExternalID = localStorage.getItem('webtv_last_channel_external_id') || ''

  if (lastPlaylistID && playlists.value.some((p) => p.id === lastPlaylistID)) {
    selectedPlaylist.value = lastPlaylistID
    await loadChannels()

    let channelToRestore = 0
    if (lastChannelID && channels.value.some((c) => c.id === lastChannelID)) {
      channelToRestore = lastChannelID
    } else if (lastChannelExternalID) {
      const byExternal = channels.value.find((c) => c.external_id === lastChannelExternalID)
      if (byExternal) channelToRestore = byExternal.id
    }
    if (channelToRestore) {
      selectedChannel.value = channelToRestore
      applyChannelVideoFit()
      await selectChannel()
      void loadAudioTracks()
    }
  }

  playbackTicker = window.setInterval(syncTimeshiftFromPlayback, 1000)
  nowTicker = window.setInterval(() => {
    nowTickMs.value = Date.now()
  }, 1000)
  revealControls()
  pictureInPictureSupported.value = !!(
    document.pictureInPictureEnabled &&
    video.value &&
    !video.value.disablePictureInPicture &&
    typeof video.value.requestPictureInPicture === 'function'
  )
  video.value?.addEventListener('enterpictureinpicture', onEnterPictureInPicture)
  video.value?.addEventListener('leavepictureinpicture', onLeavePictureInPicture)
})

onUnmounted(() => {
  if (playbackTicker) {
    clearInterval(playbackTicker)
    playbackTicker = null
  }
  if (nowTicker) {
    clearInterval(nowTicker)
    nowTicker = null
  }
  if (controlsHideTimer) {
    clearTimeout(controlsHideTimer)
    controlsHideTimer = null
  }
  flushBrokenLogos()
  clearStalledRecoveryTimer()
  clearStartupSlowTimer()
  destroyHls()
  systemThemeMediaQuery?.removeEventListener?.('change', onSystemThemeChange)
  systemThemeMediaQuery = null
  video.value?.removeEventListener('enterpictureinpicture', onEnterPictureInPicture)
  video.value?.removeEventListener('leavepictureinpicture', onLeavePictureInPicture)
})
</script>

<style>
:root {
  font-family: 'Manrope', sans-serif;
  color-scheme: light;
  --bg-page-from: #f2f7fb;
  --bg-page-to: #dfeaf5;
  --panel-bg: #ffffff;
  --text-main: #0f1720;
  --muted-text: #4d5f73;
  --border-soft: #d7e0ea;
  --card-soft: #f9fbfe;
  --card-strong: #ffffff;
  --btn-bg: #ffffff;
  --btn-text: #243445;
  --btn-border: #cfd9e4;
  --btn-hover: #f1f6fb;
  --text-on-card: #1c2b3a;
  --text-subtle-on-card: #59697a;
  --form-label: #34465a;
  --form-muted: #2f4358;
  --table-head-bg: #f3f8fd;
  --table-head-text: #2a3b4f;
  --table-row-even: #fbfdff;
  --table-cell-text: #243445;
}
:root[data-theme='dark'] {
  color-scheme: dark;
  --bg-page-from: #0e1520;
  --bg-page-to: #111c29;
  --panel-bg: #162231;
  --text-main: #e8eff8;
  --muted-text: #b1bfd0;
  --border-soft: #2a3a4c;
  --card-soft: #1a2a3b;
  --card-strong: #233447;
  --btn-bg: #1c2d3f;
  --btn-text: #e7eef8;
  --btn-border: #2f4357;
  --btn-hover: #24374c;
  --text-on-card: #ecf3fb;
  --text-subtle-on-card: #c6d4e4;
  --form-label: #c6d6e8;
  --form-muted: #d4e1ef;
  --table-head-bg: #2a3d51;
  --table-head-text: #e9f1fb;
  --table-row-even: #1a2b3d;
  --table-cell-text: #e6eef8;
}
html, body, #app { min-height: 100%; }
body { margin: 0; min-height: 100vh; background: linear-gradient(180deg, var(--bg-page-from), var(--bg-page-to)); color: var(--text-main); }
.layout { max-width: 1200px; margin: 0 auto; padding: 16px; }
.topbar { display: flex; align-items: center; justify-content: space-between; margin-bottom: 8px; }
.settings-btn {
  width: 38px;
  height: 38px;
  border: 1px solid var(--btn-border);
  border-radius: 999px;
  background: linear-gradient(180deg, var(--btn-bg), var(--card-soft));
  color: #2e6daa;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
}
.settings-btn:hover { background: var(--btn-hover); }
.settings-btn svg { width: 20px; height: 20px; fill: currentColor; }
.panel { background: var(--panel-bg); padding: 16px; border-radius: 12px; }
.panel-head { display: flex; align-items: center; justify-content: space-between; margin-bottom: 8px; }
.panel-head h2 { margin: 0; }
.panel-head-actions { display: flex; gap: 8px; align-items: center; }
.settings-tabs { display: inline-flex; gap: 8px; margin-bottom: 12px; }
.settings-tab-btn { border-radius: 999px; }
.settings-tab-btn.active {
  color: #fff;
  border-color: #2e6daa;
  background: linear-gradient(180deg, #3d86ca, #2e6daa);
}
.playlist-switcher { display: grid; grid-template-columns: minmax(0, 1fr) auto; align-items: center; column-gap: 8px; margin-bottom: 10px; width: 100%; }
.playlist-switcher-list { display: flex; flex-wrap: wrap; gap: 8px; min-width: 0; }
.playlist-switcher-actions { display: inline-flex; align-items: center; gap: 8px; }
.playlist-btn {
  border-radius: 999px;
  display: inline-flex;
  align-items: center;
  gap: 8px;
}
.playlist-btn.active {
  color: #fff;
  border-color: #2e6daa;
  background: linear-gradient(180deg, #3d86ca, #2e6daa);
}
.playlist-btn-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 20px;
  height: 20px;
  padding: 0 6px;
  border-radius: 999px;
  font-size: 12px;
  line-height: 1;
  color: var(--text-main);
  background: var(--card-soft);
}
.playlist-btn.active .playlist-btn-badge {
  color: #fff;
  background: rgba(255, 255, 255, 0.22);
}
.grid { display: grid; gap: 8px; margin-bottom: 12px; }
.field-label {
  display: grid;
  gap: 6px;
  color: var(--form-label);
  font-size: 13px;
  font-weight: 600;
}
.form-input {
  border: 1px solid #cfd9e4;
  background: #fff;
  color: #243445;
  border-radius: 8px;
  padding: 8px 10px;
  font: inherit;
}
.form-input:focus {
  outline: none;
  border-color: #3d86ca;
  box-shadow: 0 0 0 3px rgba(61, 134, 202, 0.2);
}
.checkbox { display: flex; gap: 8px; align-items: center; }
.actions { display: flex; gap: 8px; }
.row-actions { display: flex; gap: 6px; flex-wrap: nowrap; white-space: nowrap; }
.checkbox,
.debug-checkbox,
.theme-option {
  color: var(--form-muted);
}
.debug-checkbox { margin-top: 12px; font-weight: 600; }
.settings-grid { display: grid; gap: 12px; }
.interface-settings-grid {
  display: grid;
  gap: 12px;
  grid-template-columns: 1fr;
  align-items: start;
}
.settings-block {
  border: 1px solid var(--border-soft);
  border-radius: 10px;
  background: var(--card-soft);
  padding: 12px;
}
.playlists-theme-settings { margin-top: 12px; max-width: 360px; }
.settings-block h3 { margin: 0 0 10px; }
.theme-options { display: grid; gap: 8px; }
.theme-option {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
  color: var(--muted-text);
}
.playlists-table {
  width: 100%;
  border-collapse: separate;
  border-spacing: 0;
  border: 1px solid var(--border-soft);
  border-radius: 10px;
  overflow: hidden;
  background: var(--panel-bg);
}
.playlists-table-wrap {
  width: 100%;
}
.playlists-table th,
.playlists-table td {
  padding: 7px 10px;
  border-bottom: 1px solid #e6edf4;
  text-align: left;
  vertical-align: middle;
  color: var(--table-cell-text);
  line-height: 1.2;
}
.playlists-table thead th {
  background: var(--table-head-bg);
  color: var(--table-head-text);
  font-size: 12px;
  padding-top: 8px;
  padding-bottom: 8px;
}
.playlists-table tbody td {
  font-size: 14px;
}
.playlists-table td:nth-child(7),
.playlists-table td:nth-child(9) {
  white-space: nowrap;
}
.playlists-table tbody tr:nth-child(even) {
  background: var(--table-row-even);
}
.playlists-table tbody tr:last-child td {
  border-bottom: none;
}
.status-chip {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  padding: 1px 7px;
  font-size: 11px;
  font-weight: 700;
}
.status-chip.on {
  color: #1f6a2a;
  background: #e8f7eb;
  border: 1px solid #b6e0bd;
}
.status-chip.off {
  color: #6a737d;
  background: #f1f3f5;
  border: 1px solid #d4dbe2;
}
.last-error-cell {
  max-width: 220px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.playlists-table .btn {
  padding: 4px 8px;
  border-radius: 7px;
  font-size: 13px;
}
@media (max-width: 980px) {
  .playlists-table-wrap {
    overflow-x: auto;
    -webkit-overflow-scrolling: touch;
  }
  .playlists-table {
    min-width: 920px;
  }
}
@media (min-width: 981px) {
  .interface-settings-grid {
    grid-template-columns: repeat(2, minmax(260px, 360px));
  }
}
.btn {
  border: 1px solid var(--btn-border);
  background: var(--btn-bg);
  color: var(--btn-text);
  border-radius: 8px;
  padding: 6px 10px;
  font: inherit;
  font-weight: 600;
  cursor: pointer;
}
.btn.subtle:hover { background: var(--btn-hover); }
.btn.danger { border-color: #e5b9bf; color: #8f2434; background: #fff5f6; }
.btn.danger:hover { background: #ffe9ec; }
.channel-toggle-btn {
  width: 36px;
  height: 36px;
  padding: 0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}
.channel-toggle-btn svg {
  width: 18px;
  height: 18px;
  fill: currentColor;
}
.player-layout { display: grid; grid-template-columns: minmax(0, 1fr) 360px; gap: 12px; align-items: stretch; height: 75vh; min-height: 0; }
.player-layout.expanded { grid-template-columns: minmax(0, 1fr); }
.player-main { min-width: 0; min-height: 0; display: grid; grid-template-rows: auto auto minmax(0, 1fr); gap: 10px; }
.player-wrap {
  position: relative;
  border-radius: 12px;
  overflow: hidden;
  background: #000;
  width: 100%;
}
.player { width: 100%; height: 100%; object-fit: contain; background: #000; display: block; }
.player-hitbox { position: absolute; inset: 0; z-index: 1; }
.player-controls-overlay {
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 2;
  padding: 10px;
  background: linear-gradient(180deg, rgba(0, 0, 0, 0) 0%, rgba(0, 0, 0, 0.78) 100%);
  transition: opacity 0.2s ease;
}
.player-controls-overlay.hidden {
  opacity: 0;
  pointer-events: none;
}
.hls-debug-overlay {
  position: absolute;
  top: 8px;
  right: 8px;
  z-index: 3;
  display: grid;
  gap: 2px;
  min-width: 180px;
  padding: 8px 10px;
  border-radius: 8px;
  background: rgba(7, 12, 18, 0.78);
  color: #d8e6f5;
  font-size: 11px;
  line-height: 1.2;
}
.hls-debug-overlay strong {
  color: #ffffff;
  font-size: 12px;
}
.hls-debug-overlay .err {
  color: #ffc0c0;
}
.startup-overlay {
  position: absolute;
  inset: 0;
  z-index: 3;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 6px;
  background: rgba(3, 8, 14, 0.62);
  color: #eaf2fc;
  text-align: center;
  pointer-events: none;
}
.startup-overlay strong {
  font-size: 17px;
  font-weight: 700;
}
.startup-overlay span {
  font-size: 13px;
}
.startup-spinner {
  width: 36px;
  height: 36px;
  border-radius: 999px;
  border: 3px solid rgba(255, 255, 255, 0.3);
  border-top-color: #ffffff;
  animation: startup-spin 0.8s linear infinite;
}
@keyframes startup-spin {
  to { transform: rotate(360deg); }
}
.player-controls-row { display: flex; gap: 8px; margin-bottom: 8px; }
.icon-btn {
  width: 38px;
  height: 34px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0;
}
.icon-btn svg {
  width: 18px;
  height: 18px;
  fill: currentColor;
}
.volume-slider {
  width: 110px;
  accent-color: #ffffff;
}
.player-controls-overlay .btn {
  border-color: rgba(255, 255, 255, 0.35);
  color: #fff;
  background: rgba(0, 0, 0, 0.35);
}
.player-controls-overlay .btn.subtle:hover { background: rgba(255, 255, 255, 0.2); }
.channel-sidebar { border: 1px solid var(--border-soft); border-radius: 10px; padding: 10px; background: var(--card-soft); max-height: 75vh; overflow: hidden; display: grid; gap: 10px; grid-template-rows: auto auto auto 1fr; margin-top: 0; }
.sidebar-head { display: flex; justify-content: space-between; align-items: center; }
.channel-search { position: relative; display: block; }
.channel-search-input {
  width: 100%;
  box-sizing: border-box;
  border: 1px solid var(--border-soft);
  background: var(--card-strong);
  color: var(--text-on-card);
  border-radius: 8px;
  padding: 8px 32px 8px 10px;
  font: inherit;
  font-size: 13px;
}
.channel-search-input:focus {
  outline: none;
  border-color: #4d84c4;
  box-shadow: 0 0 0 3px rgba(77, 132, 196, 0.2);
}
.channel-search-clear {
  position: absolute;
  top: 50%;
  right: 6px;
  transform: translateY(-50%);
  width: 22px;
  height: 22px;
  border: none;
  border-radius: 999px;
  background: transparent;
  color: #6b7f94;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
}
.channel-search-clear:hover { background: rgba(77, 132, 196, 0.12); }
.channel-search-clear svg { width: 14px; height: 14px; fill: currentColor; }
.channel-groups { overflow: auto; padding-right: 4px; }
.favorites-block h4 { margin: 0 0 8px; }
.playlists-hidden-settings { margin-top: 12px; max-width: none; width: 100%; }
.muted { color: var(--muted-text); margin: 0 0 8px; }
.settings-channel-groups { margin-top: 10px; max-height: 55vh; overflow: auto; padding-right: 4px; }
.hidden-setting-row { display: flex; gap: 8px; align-items: center; margin-bottom: 6px; font-size: 13px; }
.group-block h4 { margin: 12px 0 8px; }
.group-head { display: flex; align-items: center; justify-content: space-between; gap: 8px; }
.group-toggle { width: 100%; display: flex; gap: 8px; align-items: center; border: none; background: transparent; padding: 0; text-align: left; cursor: pointer; }
.group-toggle h4 { margin: 12px 0 8px; }
.group-channels .hidden-setting-row { break-inside: avoid; page-break-inside: avoid; }
.channel-row { width: 100%; border: 1px solid var(--border-soft); background: var(--card-strong); color: var(--text-on-card); border-radius: 8px; padding: 6px; display: grid; grid-template-columns: 44px 1fr auto; gap: 8px; align-items: center; text-align: left; margin-bottom: 6px; }
.channel-row { content-visibility: auto; contain-intrinsic-size: 58px; }
.channel-row.selected { border-color: #4d84c4; background: #f1f7ff; }
.channel-logo { width: 44px; height: 44px; object-fit: contain; border-radius: 6px; background: #fff; }
.channel-logo-fallback { display: flex; align-items: center; justify-content: center; color: #3f4a56; font-size: 16px; font-weight: 700; border: 1px dashed #c7d2de; text-transform: uppercase; }
.channel-meta strong { display: block; font-size: 14px; }
.channel-meta span { display: block; color: var(--text-subtle-on-card); font-size: 12px; }
.channel-actions {
  justify-self: end;
  display: inline-flex;
  align-items: center;
  gap: 6px;
}
.archive-indicator {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  color: var(--text-subtle-on-card);
  font-size: 14px;
  line-height: 1;
}
.favorite-toggle { display: inline-flex; align-items: center; justify-content: center; font-size: 16px; line-height: 1; color: #b9c3cf; user-select: none; cursor: pointer; }
.favorite-toggle.active { color: #d49a00; }
.programs-under-player { border: 1px solid var(--border-soft); border-radius: 10px; background: var(--card-soft); padding: 10px; min-height: 0; display: grid; grid-template-rows: auto 1fr; }
.player-status-row { display: flex; gap: 8px; flex-wrap: wrap; }
.status-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  border: 1px solid #d7e0ea;
  border-radius: 999px;
  background: #f9fbfe;
  color: #3f4a56;
  font-size: 12px;
  line-height: 1;
  font-weight: 700;
  height: 28px;
  padding: 0 10px;
  box-sizing: border-box;
  white-space: nowrap;
}
.status-badge svg { width: 14px; height: 14px; fill: currentColor; }
.status-badge.ok { color: #2b6f2f; border-color: #b9debc; background: #eef9ef; }
.status-badge.muted { color: #6b7580; }
.status-badge.pending { color: #8a5a00; border-color: #f0d39b; background: #fff6e5; }
.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 999px;
  background: currentColor;
  animation: status-pulse 1s ease-in-out infinite;
}
.deinterlace-toggle {
  border-radius: 999px;
  padding: 4px 10px;
  min-height: 28px;
  box-sizing: border-box;
  font-size: 12px;
  line-height: 1;
  font-weight: 700;
  color: #6b7580;
}
.deinterlace-toggle.active {
  color: #1f6a2a;
  border-color: #b6e0bd;
  background: #e8f7eb;
}
@keyframes status-pulse {
  0%, 100% { opacity: 0.35; transform: scale(0.92); }
  50% { opacity: 1; transform: scale(1); }
}
.programs-under-player h3 { margin: 0 0 8px; }
.programs-custom-scroll { min-height: 0; display: grid; grid-template-columns: minmax(0, 1fr) 10px; gap: 8px; }
.programs-viewport { min-height: 0; overflow: hidden; outline: none; }
.programs-list { margin: 0; padding: 0; list-style: none; will-change: transform; }
.programs-scrollbar {
  min-height: 0;
  position: relative;
  border-radius: 999px;
  background: color-mix(in srgb, var(--card-strong) 82%, #5e7389 18%);
  user-select: none;
  touch-action: none;
}
.programs-scrollbar-thumb {
  position: absolute;
  left: 1px;
  right: 1px;
  top: 0;
  border-radius: 999px;
  background: color-mix(in srgb, #4d84c4 70%, #a8c2dd 30%);
  cursor: grab;
}
.programs-scrollbar-thumb:active { cursor: grabbing; }
.program-item { margin-bottom: 6px; }
.program-item button { width: 100%; text-align: left; border: 1px solid var(--border-soft); border-radius: 8px; background: var(--card-strong); color: var(--text-on-card); padding: 8px; cursor: pointer; position: relative; overflow: hidden; }
.program-item button:disabled { opacity: 0.75; cursor: not-allowed; background: color-mix(in srgb, var(--card-strong) 70%, #6d7d90 30%); color: var(--text-subtle-on-card); }
.program-item.current button { border-color: #2e6daa; background: #eef6ff; font-weight: 700; }
.program-item.selected button { box-shadow: inset 0 0 0 1px #2e6daa; }
.program-item.archive-playing button { border-color: #bf6f00; background: #fff3dc; box-shadow: inset 0 0 0 1px #bf6f00; }
.program-progress {
  display: block;
  margin-top: 6px;
  height: 4px;
  border-radius: 999px;
  background: rgba(46, 109, 170, 0.2);
}
.program-progress-fill {
  display: block;
  height: 100%;
  border-radius: inherit;
  background: linear-gradient(90deg, #3d86ca, #2e6daa);
}
.timeshift { margin-top: 10px; padding: 10px; border: 1px solid #d7e0ea; border-radius: 10px; background: #f9fbfe; }
.timeshift-head { display: flex; justify-content: space-between; gap: 10px; margin-bottom: 8px; color: #4a5868; font-size: 13px; }
.timeshift input[type="range"] { width: 100%; }
.timeshift.in-player {
  margin-top: 0;
  border-color: rgba(255, 255, 255, 0.25);
  background: rgba(0, 0, 0, 0.35);
}
.timeshift.in-player .timeshift-head { color: #e8f0ff; }
@media (max-width: 980px) {
  .player-layout { grid-template-columns: 1fr; height: auto; min-height: 0; }
  .player-main { grid-template-rows: auto auto auto; }
  .programs-viewport { max-height: 220px; }
  .channel-sidebar { max-height: none; margin-top: 0; }
  .playlist-switcher { align-items: flex-start; }
  .settings-channel-groups { max-height: 50vh; }
}
@media (min-width: 981px) {
  .group-channels {
    column-width: 240px;
    column-gap: 20px;
  }
}
@media (max-height: 760px) {
  .player-layout { height: auto; min-height: 0; }
  .player-main { grid-template-rows: auto auto auto; }
  .programs-viewport { max-height: 180px; }
}
:root[data-theme='dark'] .channel-row.selected { background: #21364d; }
:root[data-theme='dark'] .program-item.current button { background: #21364d; color: #f0f6ff; }
:root[data-theme='dark'] .program-item.archive-playing button { background: #4f3a1f; color: #fff3dd; }
:root[data-theme='dark'] .channel-logo { background: #203247; }
:root[data-theme='dark'] .channel-logo-fallback { color: #d6e3f2; border-color: #42586d; }
</style>
