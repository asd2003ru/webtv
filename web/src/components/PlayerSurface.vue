<template>
  <div>
    <div :ref="onSetPlayerWrapRef" class="player-wrap" :style="{ aspectRatio: playerAspectRatio }">
      <video
        :ref="onSetVideoRef"
        autoplay
        playsinline
        class="player"
        :style="{ objectFit: currentVideoFitMode }"
        @loadedmetadata="onVideoMetadata"
        @ended="onVideoEnded"
        @play="onVideoPlay"
        @pause="onVideoPause"
        @volumechange="onVideoVolumeChange"
        @timeupdate="onVideoTimeUpdate"
        @canplay="onVideoCanPlay"
        @error="onVideoError"
        @waiting="onVideoWaiting"
        @stalled="onVideoStalled"
      ></video>
      <div v-if="startupOverlayVisible" class="startup-overlay" role="status" aria-live="polite">
        <span class="startup-spinner" aria-hidden="true"></span>
        <strong>{{ startupOverlayTitle }}</strong>
        <span>{{ startupOverlayText }}</span>
      </div>
      <div
        class="player-hitbox"
        tabindex="0"
        @mousemove="onRevealControls"
        @click="onRevealControls"
        @touchstart="onRevealControls"
        @keydown="onRevealControls"
      ></div>
      <div class="player-controls-overlay" :class="{ hidden: !controlsVisible }">
        <div class="player-controls-row">
          <button
            type="button"
            class="btn subtle icon-btn"
            :aria-label="isPlaying ? t('player_pause') : t('player_play')"
            :title="isPlaying ? t('player_pause') : t('player_play')"
            @click="onTogglePlayPause"
          >
            <svg v-if="!isPlaying" viewBox="0 0 24 24" aria-hidden="true">
              <path d="M8 5v14l11-7z" />
            </svg>
            <svg v-else viewBox="0 0 24 24" aria-hidden="true">
              <path d="M7 5h4v14H7zm6 0h4v14h-4z" />
            </svg>
          </button>
          <button
            type="button"
            class="btn subtle icon-btn"
            :aria-label="isMuted ? t('player_unmute') : t('player_mute')"
            :title="isMuted ? t('player_unmute') : t('player_mute')"
            @click="onToggleMute"
          >
            <svg v-if="!isMuted" viewBox="0 0 24 24" aria-hidden="true">
              <path d="M3 10v4h4l5 4V6L7 10H3zm13.5 2a3.5 3.5 0 0 0-1.5-2.87v5.74A3.5 3.5 0 0 0 16.5 12zm0-7a10 10 0 0 1 0 14l-1.41-1.41a8 8 0 0 0 0-11.18L16.5 5z" />
            </svg>
            <svg v-else viewBox="0 0 24 24" aria-hidden="true">
              <path d="M3 10v4h4l5 4V6L7 10H3zm10.59 2 2.7 2.7-1.3 1.3-2.7-2.7-2.7 2.7-1.3-1.3 2.7-2.7-2.7-2.7 1.3-1.3 2.7 2.7 2.7-2.7 1.3 1.3-2.7 2.7z" />
            </svg>
          </button>
          <input
            class="volume-slider"
            type="range"
            min="0"
            max="1"
            step="0.01"
            :value="volumeLevel"
            aria-label="Volume"
            @input="onVolumeInput"
          />
          <button
            type="button"
            class="btn subtle icon-btn"
            :aria-label="t('player_previous_program')"
            :title="t('player_previous_program')"
            :disabled="!canPrevProgram"
            @click="onSelectPrevProgram"
          >
            <svg viewBox="0 0 24 24" aria-hidden="true">
              <path d="M6 6h2v12H6zm3 6 9-6v12z" />
            </svg>
          </button>
          <button
            type="button"
            class="btn subtle icon-btn"
            :aria-label="t('player_next_program')"
            :title="t('player_next_program')"
            :disabled="!canNextProgram"
            @click="onSelectNextProgram"
          >
            <svg viewBox="0 0 24 24" aria-hidden="true">
              <path d="M16 6h2v12h-2zM7 18V6l9 6z" />
            </svg>
          </button>
          <button
            type="button"
            class="btn subtle icon-btn"
            :aria-label="t('player_back_10')"
            :title="t('player_back_10')"
            :disabled="!canSeekArchive"
            @click="onSeekBySeconds(-10)"
          >
            <svg viewBox="0 0 24 24" aria-hidden="true">
              <path d="M12 5a7 7 0 1 1-6.32 10H3l3.5-3.5L10 15H7.7A5 5 0 1 0 12 7v2l-3-3 3-3v2zm1 4h-2v3l2.5 1.5 1-1.73-1.5-.87z" />
            </svg>
          </button>
          <button
            type="button"
            class="btn subtle icon-btn"
            :aria-label="t('player_forward_10')"
            :title="t('player_forward_10')"
            :disabled="!canSeekArchive"
            @click="onSeekBySeconds(10)"
          >
            <svg viewBox="0 0 24 24" aria-hidden="true">
              <path d="M12 5V3l3 3-3 3V7a5 5 0 1 1-4.3 7H5.68A7 7 0 1 0 12 5zm-1 4h2v3l2.5 1.5-1 1.73L11 13z" />
            </svg>
          </button>
          <button
            type="button"
            class="btn subtle icon-btn"
            :aria-label="t('player_fullscreen')"
            :title="t('player_fullscreen')"
            @click="onToggleFullscreen"
          >
            <svg viewBox="0 0 24 24" aria-hidden="true">
              <path d="M7 3H3v4h2V5h2V3zm14 0h-4v2h2v2h2V3zM5 17H3v4h4v-2H5v-2zm16 0h-2v2h-2v2h4v-4z" />
            </svg>
          </button>
          <button
            type="button"
            class="btn subtle icon-btn"
            :aria-label="t('player_video_fit').replace('{mode}', videoFitModeLabel)"
            :title="t('player_video_fit').replace('{mode}', videoFitModeLabel)"
            :disabled="!selectedChannel"
            @click="onCycleVideoFitMode"
          >
            <svg v-if="currentVideoFitMode === 'contain'" viewBox="0 0 24 24" aria-hidden="true">
              <path d="M3 5h18v14H3zm2 2v10h14V7zm2 2h3v2H7zm7 0h3v2h-3zm-7 4h3v2H7zm7 0h3v2h-3z" />
            </svg>
            <svg v-else-if="currentVideoFitMode === 'cover'" viewBox="0 0 24 24" aria-hidden="true">
              <path d="M3 5h18v14H3zm2 2v10h14V7zm2 2h10v6H7z" />
            </svg>
            <svg v-else viewBox="0 0 24 24" aria-hidden="true">
              <path d="M3 5h18v14H3zm2 2v10h14V7zM7 9h1v6H7zm9 0h1v6h-1zM11 9h2v6h-2z" />
            </svg>
          </button>
          <button
            type="button"
            class="btn subtle icon-btn"
            :aria-label="isInPictureInPicture ? t('player_pip_close') : t('player_pip_open')"
            :title="isInPictureInPicture ? t('player_pip_close') : t('player_pip_open')"
            :disabled="!selectedChannel || !pictureInPictureSupported"
            @click="onDetachPlayer"
          >
            <svg viewBox="0 0 24 24" aria-hidden="true">
              <path d="M5 4h8v2H5v12h12v-8h2v10H3V4zm8 0h8v8h-2V7.41l-7.29 7.3-1.42-1.42L17.59 6H13V4z" />
            </svg>
          </button>
        </div>
        <div v-if="archiveSupported && selectedProgram && timeshiftMaxSeconds > 0" class="timeshift in-player">
          <div class="timeshift-head">
            <strong>{{ selectedProgram.title }}</strong>
            <span>{{ timeshiftOffsetLabel }} / {{ timeshiftMaxLabel }}</span>
          </div>
          <input
            type="range"
            min="0"
            :max="timeshiftMaxSeconds"
            :value="timeshiftOffsetSeconds"
            @input="onTimeshiftInput"
            @change="onApplyTimeshift"
          />
        </div>
      </div>
      <div v-if="showDebugOverlay" class="hls-debug-overlay">
        <strong>HLS Debug</strong>
        <span>engine: {{ hlsDebug.engine }}</span>
        <span>server-mode: {{ hlsDebug.serverMode }}</span>
        <span>video-path: {{ hlsDebug.videoPath }}</span>
        <span>audio-path: {{ hlsDebug.audioPath }}</span>
        <span>audio-tracks: {{ hlsDebug.audioTracks }}</span>
        <span>audio-target: {{ hlsDebug.audioTarget }}</span>
        <span>audio-result: {{ hlsDebug.audioResult }}</span>
        <span>level: {{ hlsDebug.level }}</span>
        <span>buffer: {{ hlsDebug.buffer }}</span>
        <span>latency: {{ hlsDebug.latency }}</span>
        <span>retries: {{ hlsDebug.retries }}</span>
        <span>state: {{ hlsDebug.state }}</span>
        <span v-if="hlsDebug.lastError" class="err">{{ hlsDebug.lastError }}</span>
      </div>
    </div>
  </div>
</template>

<script setup>
defineProps({
  onSetPlayerWrapRef: { type: Function, required: true },
  onSetVideoRef: { type: Function, required: true },
  playerAspectRatio: { type: String, required: true },
  currentVideoFitMode: { type: String, required: true },
  startupOverlayVisible: { type: Boolean, required: true },
  startupOverlayTitle: { type: String, required: true },
  startupOverlayText: { type: String, required: true },
  controlsVisible: { type: Boolean, required: true },
  isPlaying: { type: Boolean, required: true },
  isMuted: { type: Boolean, required: true },
  volumeLevel: { type: Number, required: true },
  canPrevProgram: { type: Boolean, required: true },
  canNextProgram: { type: Boolean, required: true },
  canSeekArchive: { type: Boolean, required: true },
  videoFitModeLabel: { type: String, required: true },
  selectedChannel: { type: Number, required: true },
  isInPictureInPicture: { type: Boolean, required: true },
  pictureInPictureSupported: { type: Boolean, required: true },
  archiveSupported: { type: Boolean, required: true },
  selectedProgram: { type: Object, default: null },
  timeshiftMaxSeconds: { type: Number, required: true },
  timeshiftOffsetSeconds: { type: Number, required: true },
  timeshiftOffsetLabel: { type: String, required: true },
  timeshiftMaxLabel: { type: String, required: true },
  showDebugOverlay: { type: Boolean, required: true },
  hlsDebug: { type: Object, required: true },
  t: { type: Function, required: true },
  onVideoMetadata: { type: Function, required: true },
  onVideoEnded: { type: Function, required: true },
  onVideoPlay: { type: Function, required: true },
  onVideoPause: { type: Function, required: true },
  onVideoVolumeChange: { type: Function, required: true },
  onVideoTimeUpdate: { type: Function, required: true },
  onVideoCanPlay: { type: Function, required: true },
  onVideoError: { type: Function, required: true },
  onVideoWaiting: { type: Function, required: true },
  onVideoStalled: { type: Function, required: true },
  onRevealControls: { type: Function, required: true },
  onTogglePlayPause: { type: Function, required: true },
  onToggleMute: { type: Function, required: true },
  onVolumeInput: { type: Function, required: true },
  onSelectPrevProgram: { type: Function, required: true },
  onSelectNextProgram: { type: Function, required: true },
  onSeekBySeconds: { type: Function, required: true },
  onToggleFullscreen: { type: Function, required: true },
  onCycleVideoFitMode: { type: Function, required: true },
  onDetachPlayer: { type: Function, required: true },
  onTimeshiftInput: { type: Function, required: true },
  onApplyTimeshift: { type: Function, required: true }
})
</script>
