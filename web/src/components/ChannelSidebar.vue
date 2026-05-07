<template>
  <aside v-show="sidebarOpen" class="channel-sidebar">
    <div class="sidebar-head">
      <h3>{{ t('channels') }}</h3>
    </div>
    <label class="channel-search">
      <input
        :value="channelSearchQuery"
        type="text"
        class="channel-search-input"
        :placeholder="t('channel_search_placeholder')"
        :aria-label="t('channel_search')"
        @input="onSetChannelSearchQuery($event.target.value.trim())"
      />
      <button
        v-if="channelSearchQuery"
        type="button"
        class="channel-search-clear"
        :title="t('reset_search')"
        :aria-label="t('reset_search')"
        @click="onSetChannelSearchQuery('')"
      >
        <svg viewBox="0 0 24 24" aria-hidden="true">
          <path d="M18.3 5.71 12 12l6.3 6.29-1.41 1.42L10.59 13.4 4.29 19.7 2.88 18.29 9.17 12 2.88 5.71 4.29 4.29l6.3 6.3 6.29-6.3z" />
        </svg>
      </button>
    </label>

    <div class="channel-groups">
      <section v-if="favoriteChannels.length > 0" class="group-block favorites-block">
        <div class="group-head">
          <button type="button" class="group-toggle" @click="onToggleGroup(FAVORITES_GROUP)">
            <span>{{ isGroupOpen(FAVORITES_GROUP) ? '▾' : '▸' }}</span>
            <h4>{{ t('favorites') }}</h4>
          </button>
        </div>
        <div v-if="isGroupOpen(FAVORITES_GROUP)">
          <button
            v-for="c in favoriteChannels"
            :key="`fav-${c.playlist_id}-${c.id}`"
            type="button"
            class="channel-row"
            :class="{ selected: selectedChannel === c.id && selectedPlaylist === c.playlist_id }"
            v-memo="[selectedChannel === c.id && selectedPlaylist === c.playlist_id, favoriteNowProgramForChannel(c)?.title || '', favoriteNowProgramForChannel(c)?.start_at || '', favoriteNowProgramForChannel(c)?.end_at || '']"
            @click="onPickFavorite(c)"
          >
            <img
              v-if="showChannelLogo(c)"
              :src="normalizeLogoUrl(c.logo)"
              :alt="c.name"
              class="channel-logo"
              width="44"
              height="44"
              loading="lazy"
              decoding="async"
              fetchpriority="low"
              @error="onMarkFavoriteLogoError(c)"
            />
            <div v-else class="channel-logo channel-logo-fallback">{{ channelInitial(c.name) }}</div>
            <div class="channel-meta">
              <strong>{{ c.name }}</strong>
              <span class="channel-program-title">{{ (playlistNameById[c.playlist_id] || `Playlist #${c.playlist_id}`) }} · {{ favoriteNowProgramTitle(c) }}</span>
              <span
                v-if="hasTimedProgram(favoriteNowProgramForChannel(c))"
                class="channel-program-progress"
                aria-hidden="true"
              >
                <span
                  class="channel-program-progress-fill"
                  :style="programProgressAnimationStyle(favoriteNowProgramForChannel(c))"
                ></span>
              </span>
            </div>
            <div class="channel-actions">
              <span v-if="c.archive_supported" class="archive-indicator" :title="t('has_archive')">⏱</span>
              <span class="favorite-toggle active" :title="t('remove_favorite')" @click.stop="onToggleFavorite(c)">
                ★
              </span>
            </div>
          </button>
        </div>
      </section>

      <section v-for="group in visibleGroupedChannels" :key="group.name" class="group-block">
        <div class="group-head">
          <button type="button" class="group-toggle" @click="onToggleGroup(group.name)">
            <span>{{ isGroupOpen(group.name) ? '▾' : '▸' }}</span>
            <h4>{{ group.name || t('no_group') }}</h4>
          </button>
        </div>
        <div v-if="isGroupOpen(group.name)">
          <button
            v-for="c in group.channels"
            :key="c.id"
            type="button"
            class="channel-row"
            :class="{ selected: selectedChannel === c.id }"
            v-memo="[selectedChannel === c.id, nowProgramByChannel[c.id]?.title || '', nowProgramByChannel[c.id]?.start_at || '', nowProgramByChannel[c.id]?.end_at || '', isFavorite(c.playlist_id, c.id)]"
            @click="onPickChannel(c.id)"
          >
            <img
              v-if="showChannelLogo(c)"
              :src="normalizeLogoUrl(c.logo)"
              :alt="c.name"
              class="channel-logo"
              width="44"
              height="44"
              loading="lazy"
              decoding="async"
              fetchpriority="low"
              @error="onMarkLogoError(c)"
            />
            <div v-else class="channel-logo channel-logo-fallback">{{ channelInitial(c.name) }}</div>
            <div class="channel-meta">
              <strong>{{ c.name }}</strong>
              <span class="channel-program-title">{{ nowProgramByChannel[c.id]?.title || t('no_current_program') }}</span>
              <span
                v-if="hasTimedProgram(nowProgramByChannel[c.id])"
                class="channel-program-progress"
                aria-hidden="true"
              >
                <span
                  class="channel-program-progress-fill"
                  :style="programProgressAnimationStyle(nowProgramByChannel[c.id])"
                ></span>
              </span>
            </div>
            <div class="channel-actions">
              <span v-if="c.archive_supported" class="archive-indicator" :title="t('has_archive')">⏱</span>
              <span
                class="favorite-toggle"
                :class="{ active: isFavorite(c.playlist_id, c.id) }"
                :title="isFavorite(c.playlist_id, c.id) ? t('remove_favorite') : t('add_favorite')"
                @click.stop="onToggleFavorite(c)"
              >
                ★
              </span>
            </div>
          </button>
        </div>
      </section>
    </div>
  </aside>
</template>

<script setup>
import { computed, onUnmounted, ref, watch } from 'vue'
import { normalizeLogoUrl } from '../utils/channels'

const props = defineProps({
  sidebarOpen: { type: Boolean, required: true },
  t: { type: Function, required: true },
  channelSearchQuery: { type: String, required: true },
  favoriteChannels: { type: Array, required: true },
  selectedChannel: { type: Number, required: true },
  selectedPlaylist: { type: Number, required: true },
  playlistNameById: { type: Object, required: true },
  groupedChannels: { type: Array, required: true },
  nowProgramByChannel: { type: Object, required: true },
  isGroupOpen: { type: Function, required: true },
  showChannelLogo: { type: Function, required: true },
  channelInitial: { type: Function, required: true },
  favoriteNowProgramTitle: { type: Function, required: true },
  favoriteNowProgramForChannel: { type: Function, required: true },
  isFavorite: { type: Function, required: true },
  onToggleGroup: { type: Function, required: true },
  onPickChannel: { type: Function, required: true },
  onPickFavorite: { type: Function, required: true },
  onToggleFavorite: { type: Function, required: true },
  onMarkLogoError: { type: Function, required: true },
  onMarkFavoriteLogoError: { type: Function, required: true },
  onSetChannelSearchQuery: { type: Function, required: true }
})

const GROUP_INITIAL_RENDER = 120
const GROUP_RENDER_STEP = 120
const FAVORITES_GROUP = '__favorites__'
const groupLimits = ref({})
let progressiveRenderTimer = null

function timedProgramBounds(program) {
  const start = new Date(program?.start_at).getTime()
  const end = new Date(program?.end_at).getTime()
  if (!Number.isFinite(start) || !Number.isFinite(end) || end <= start) return null
  return { start, end }
}

function hasTimedProgram(program) {
  return Boolean(timedProgramBounds(program))
}

function programProgressAnimationStyle(program) {
  const bounds = timedProgramBounds(program)
  if (!bounds) return {}
  const duration = bounds.end - bounds.start
  const elapsed = Math.max(0, Math.min(Date.now() - bounds.start, duration))
  return {
    '--channel-program-duration': `${duration}ms`,
    '--channel-program-delay': `${-elapsed}ms`
  }
}

watch(
  () => props.groupedChannels,
  (groups) => {
    const next = {}
    for (const group of groups) {
      next[group.name] = groupLimits.value[group.name] || GROUP_INITIAL_RENDER
    }
    groupLimits.value = next
  },
  { immediate: true }
)

const visibleGroupedChannels = computed(() => {
  return props.groupedChannels.map((group) => {
    const limit = groupLimits.value[group.name] || GROUP_INITIAL_RENDER
    return {
      ...group,
      channels: group.channels.slice(0, limit)
    }
  })
})

function scheduleProgressiveRender() {
  if (progressiveRenderTimer) {
    clearTimeout(progressiveRenderTimer)
    progressiveRenderTimer = null
  }
  progressiveRenderTimer = setTimeout(() => {
    let updated = false
    const next = { ...groupLimits.value }
    for (const group of props.groupedChannels) {
      const current = next[group.name] || GROUP_INITIAL_RENDER
      const total = group.channels.length
      if (current < total) {
        next[group.name] = Math.min(current + GROUP_RENDER_STEP, total)
        updated = true
      }
    }
    if (updated) {
      groupLimits.value = next
      scheduleProgressiveRender()
    } else {
      progressiveRenderTimer = null
    }
  }, 16)
}

watch(
  () => visibleGroupedChannels.value.map((g) => g.channels.length).join(','),
  () => {
    scheduleProgressiveRender()
  }
)

onUnmounted(() => {
  if (progressiveRenderTimer) {
    clearTimeout(progressiveRenderTimer)
    progressiveRenderTimer = null
  }
})
</script>
