<template>
  <div class="player-status-row">
    <div v-if="audioTrackOptions.length > 1" class="status-badge audio-track-badge">
      <span class="audio-track-icon" aria-hidden="true">
        <svg viewBox="0 0 24 24">
          <path d="M3 10v4h4l5 4V6L7 10H3zm13.5 2a3.5 3.5 0 0 0-1.5-2.87v5.74A3.5 3.5 0 0 0 16.5 12zm0-7a10 10 0 0 1 0 14l-1.41-1.41a8 8 0 0 0 0-11.18L16.5 5z" />
        </svg>
      </span>
      <select
        id="audio-track-select"
        class="audio-track-select"
        :value="selectedAudioTrack"
        @change="onSelectAudioTrack"
      >
        <option v-for="track in audioTrackOptions" :key="track.value" :value="track.value">
          {{ track.label }}
        </option>
      </select>
    </div>
    <div class="status-badge" :title="`${t('stream_mode')}: ${streamModeDisplay(streamMode)}`">
      <svg viewBox="0 0 24 24" aria-hidden="true">
        <path d="M4 6h16v10H4zM2 4v14h20V4zm8 16h4v-2h-4z" />
      </svg>
      <span>{{ streamModeDisplay(streamMode) }}</span>
    </div>
    <div
      class="status-badge"
      :class="{ ok: archiveSupported, muted: !archiveSupported }"
      :title="archiveSupported ? t('archive_supported') : t('archive_not_supported')"
    >
      <svg v-if="archiveSupported" viewBox="0 0 24 24" aria-hidden="true">
        <path d="M12 2a10 10 0 1 0 10 10h-2a8 8 0 1 1-8-8V2zm1 5h-2v6l5 3 1-1.73-4-2.27z" />
      </svg>
      <svg v-else viewBox="0 0 24 24" aria-hidden="true">
        <path d="M12 2a10 10 0 0 0-7.07 17.07l1.41-1.41A8 8 0 1 1 20 12h2A10 10 0 0 0 12 2zm5.66 17.24-4.79-4.79V7h-2v6.62l-1.58-1.58-1.41 1.41 8.37 8.37z" />
      </svg>
      <span>{{ archiveSupported ? t('archive') : t('no_archive') }}</span>
    </div>
    <button
      v-if="streamMode === 'transcode'"
      type="button"
      class="btn deinterlace-toggle"
      :class="{ active: deinterlaceEnabled }"
      :title="deinterlaceEnabled ? t('deinterlace_on') : t('deinterlace_off')"
      @click="onToggleDeinterlace"
    >
      {{ deinterlaceEnabled ? t('deinterlace_on_short') : t('deinterlace_off_short') }}
    </button>
  </div>

  <div v-if="selectedChannel" class="programs-under-player">
    <h3>{{ t('channel_programs') }}</h3>

    <div class="programs-custom-scroll">
      <div
        ref="viewportEl"
        class="programs-viewport"
        tabindex="0"
        @keydown="onViewportKeyDown"
        @scroll="onViewportScroll"
      >
        <ul
          ref="contentEl"
          class="programs-list"
        >
          <li
            v-for="p in programs"
            :key="programKey(p)"
            class="program-item"
            :data-program-key="programKey(p)"
            :class="{ current: isCurrentProgram(p), selected: selectedProgramKey === programKey(p), 'archive-playing': isArchivePlayingProgram(p) }"
          >
            <button
              type="button"
              :disabled="isFutureProgram(p)"
              :title="programDescriptionTooltip(p)"
              @click="onSelectProgram(p)"
            >
              <span class="program-row">
                <span class="program-title">[{{ formatProgramDate(p.start_at) }}] {{ p.title }}</span>
                <span v-if="isTimeshiftPlayingProgram(p)" class="program-time archive">
                  <span class="program-mode-pill">{{ t('archive') }}</span>
                  {{ selectedProgramProgressLabel }}
                </span>
                <span v-else-if="isCurrentProgram(p)" class="program-time">{{ programProgressLabel(p) }}</span>
                <span v-else-if="isArchivePlayingProgram(p)" class="program-time archive">
                  <span class="program-mode-pill">{{ t('archive') }}</span>
                  {{ selectedProgramProgressLabel }}
                </span>
              </span>
              <span v-if="isCurrentProgram(p) || isArchivePlayingProgram(p)" class="program-progress">
                <span
                  v-if="isTimeshiftPlayingProgram(p)"
                  class="program-progress-live"
                  :style="{ width: `${programProgressPercent(p)}%` }"
                ></span>
                <span
                  class="program-progress-fill"
                  :class="{ archive: isArchivePlayingProgram(p) || isTimeshiftPlayingProgram(p) }"
                  :style="{ width: `${isArchivePlayingProgram(p) || isTimeshiftPlayingProgram(p) ? selectedProgramProgressPercent : programProgressPercent(p)}%` }"
                ></span>
              </span>
            </button>
          </li>
        </ul>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'

const props = defineProps({
  t: { type: Function, required: true },
  streamModeDisplay: { type: Function, required: true },
  streamMode: { type: String, required: true },
  archiveSupported: { type: Boolean, required: true },
  selectedProgramProgressLabel: { type: String, required: true },
  selectedProgramProgressPercent: { type: Number, required: true },
  audioTrackOptions: { type: Array, required: true },
  selectedAudioTrack: { type: Number, required: true },
  deinterlaceEnabled: { type: Boolean, required: true },
  selectedChannel: { type: Number, required: true },
  programs: { type: Array, required: true },
  selectedProgramKey: { type: String, required: true },
  programKey: { type: Function, required: true },
  isCurrentProgram: { type: Function, required: true },
  isArchivePlayingProgram: { type: Function, required: true },
  isTimeshiftPlayingProgram: { type: Function, required: true },
  isFutureProgram: { type: Function, required: true },
  programDescriptionTooltip: { type: Function, required: true },
  formatProgramDate: { type: Function, required: true },
  programProgressPercent: { type: Function, required: true },
  programProgressLabel: { type: Function, required: true },
  onSelectProgram: { type: Function, required: true },
  onToggleDeinterlace: { type: Function, required: true },
  onSelectAudioTrack: { type: Function, required: true }
})

const viewportEl = ref(null)
const contentEl = ref(null)

const scrollTopPx = ref(0)
const viewportHeightPx = ref(0)
const contentHeightPx = ref(0)

let resizeObserver = null
let focusRequestId = 0
const LINE_SCROLL_STEP_PX = 40
const MIN_PAGE_SCROLL_STEP_PX = 40

const maxScrollPx = computed(() => Math.max(0, contentHeightPx.value - viewportHeightPx.value))

function clampScroll(next) {
  if (!Number.isFinite(next)) return 0
  return Math.max(0, Math.min(maxScrollPx.value, next))
}

function setScrollTop(next) {
  const nextTop = clampScroll(next)
  scrollTopPx.value = nextTop
  if (viewportEl.value && Math.abs(viewportEl.value.scrollTop - nextTop) > 1) {
    viewportEl.value.scrollTop = nextTop
  }
}

function measureScroller() {
  viewportHeightPx.value = viewportEl.value?.clientHeight || 0
  contentHeightPx.value = viewportEl.value?.scrollHeight || contentEl.value?.scrollHeight || 0
  setScrollTop(scrollTopPx.value)
}

function findFocusTargetElement() {
  if (!contentEl.value) return null
  const currentEl = contentEl.value.querySelector('.program-item.current')
  if (currentEl) return currentEl
  if (!props.selectedProgramKey) return null
  return contentEl.value.querySelector(`[data-program-key="${props.selectedProgramKey}"]`)
}

function focusCurrentProgramInList() {
  const targetEl = findFocusTargetElement()
  if (!targetEl || !viewportEl.value) return
  const viewportRect = viewportEl.value.getBoundingClientRect()
  const targetRect = targetEl.getBoundingClientRect()
  const targetTop = viewportEl.value.scrollTop + targetRect.top - viewportRect.top
  const next = targetTop - Math.max(0, Math.floor((viewportEl.value.clientHeight - targetRect.height) / 2))
  setScrollTop(next)
}

function onViewportScroll(event) {
  scrollTopPx.value = clampScroll(event.currentTarget.scrollTop)
}

function onViewportKeyDown(event) {
  const pageStep = Math.max(MIN_PAGE_SCROLL_STEP_PX, Math.floor(viewportHeightPx.value * 0.9))
  const actionByKey = {
    ArrowDown: () => setScrollTop(scrollTopPx.value + LINE_SCROLL_STEP_PX),
    ArrowUp: () => setScrollTop(scrollTopPx.value - LINE_SCROLL_STEP_PX),
    PageDown: () => setScrollTop(scrollTopPx.value + pageStep),
    PageUp: () => setScrollTop(scrollTopPx.value - pageStep),
    Home: () => setScrollTop(0),
    End: () => setScrollTop(maxScrollPx.value)
  }
  const action = actionByKey[event.key]
  if (!action) return
  event.preventDefault()
  action()
}

async function measureAndFocusCurrent() {
  const requestId = ++focusRequestId
  await nextTick()
  for (let frame = 0; frame < 3; frame += 1) {
    if (typeof requestAnimationFrame === 'function') {
      await new Promise((resolve) => requestAnimationFrame(resolve))
    }
    if (requestId !== focusRequestId) return
    measureScroller()
    focusCurrentProgramInList()
  }
}

watch(
  () => [props.selectedChannel, props.programs.length, props.selectedProgramKey],
  async () => { await measureAndFocusCurrent() },
  { flush: 'post' }
)

onMounted(async () => {
  await measureAndFocusCurrent()
  if (typeof ResizeObserver !== 'undefined') {
    resizeObserver = new ResizeObserver(() => {
      measureScroller()
    })
    if (viewportEl.value) resizeObserver.observe(viewportEl.value)
    if (contentEl.value) resizeObserver.observe(contentEl.value)
  }
})

onUnmounted(() => {
  if (resizeObserver) {
    resizeObserver.disconnect()
    resizeObserver = null
  }
})
</script>

<style scoped>
.audio-track-badge {
  gap: 4px;
  padding-right: 8px;
  padding-left: 6px;
}

.audio-track-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 14px;
  height: 14px;
}

.audio-track-icon svg {
  width: 14px;
  height: 14px;
  fill: currentColor;
}

.audio-track-select {
  border: none;
  background: transparent;
  color: inherit;
  font-size: 12px;
  font-weight: 700;
  line-height: 1;
  height: 22px;
  min-width: 92px;
  max-width: 130px;
  padding: 0 15px 0 2px;
  appearance: none;
  -webkit-appearance: none;
  -moz-appearance: none;
  background-image:
    linear-gradient(45deg, transparent 50%, currentColor 50%),
    linear-gradient(135deg, currentColor 50%, transparent 50%);
  background-position:
    calc(100% - 9px) 9px,
    calc(100% - 5px) 9px;
  background-size: 4px 4px, 4px 4px;
  background-repeat: no-repeat;
}

.audio-track-select:focus {
  outline: none;
}
</style>
