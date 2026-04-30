<template>
  <section v-show="tab === 'playlists'" class="panel">
    <div class="panel-head">
      <h2>{{ t('settings') }}</h2>
      <button
        type="button"
        class="settings-btn"
        :aria-label="t('close')"
        :title="t('close')"
        @click="onClose"
      >
        <svg viewBox="0 0 24 24" aria-hidden="true">
          <path d="M18.3 5.71a1 1 0 0 0-1.41 0L12 10.59 7.11 5.7A1 1 0 0 0 5.7 7.11L10.59 12 5.7 16.89a1 1 0 1 0 1.41 1.41L12 13.41l4.89 4.89a1 1 0 0 0 1.41-1.41L13.41 12l4.89-4.89a1 1 0 0 0 0-1.4z" />
        </svg>
      </button>
    </div>
    <div class="settings-tabs" role="tablist" :aria-label="t('settings_tabs')">
      <button
        type="button"
        class="btn settings-tab-btn"
        :class="{ active: settingsTab === 'playlists' }"
        :aria-selected="settingsTab === 'playlists'"
        @click="settingsTab = 'playlists'"
      >
        {{ t('playlists') }}
      </button>
      <button
        type="button"
        class="btn settings-tab-btn"
        :class="{ active: settingsTab === 'interface' }"
        :aria-selected="settingsTab === 'interface'"
        @click="settingsTab = 'interface'"
      >
        {{ t('interface') }}
      </button>
    </div>

    <template v-if="settingsTab === 'playlists'">
      <form class="grid" @submit.prevent="onSavePlaylist">
        <label class="field-label">
          <span>{{ t('playlist_name') }}</span>
          <input v-model="form.name" class="form-input" :placeholder="t('playlist_name_placeholder')" required />
        </label>
        <label class="field-label">
          <span>M3U URL</span>
          <input v-model="form.m3u_url" class="form-input" placeholder="https://..." required />
        </label>
        <label class="field-label">
          <span>EPG XML.GZ URL</span>
          <input v-model="form.epg_url" class="form-input" placeholder="https://..." required />
        </label>
        <label class="field-label">
          <span>{{ t('autoupdate_interval') }}</span>
          <input v-model.number="form.update_interval_minutes" class="form-input" type="number" min="1" :placeholder="t('autoupdate_interval_placeholder')" required />
        </label>
        <label class="checkbox">
          <input v-model="form.enabled" type="checkbox" />
          {{ t('autoupdate_enabled') }}
        </label>
        <div class="actions">
          <button type="submit" class="btn">{{ editingId ? t('save_changes') : t('save') }}</button>
          <button v-if="editingId" type="button" class="btn subtle" @click="onCancelEdit">{{ t('cancel') }}</button>
        </div>
      </form>
      <div class="playlists-table-wrap">
        <table class="playlists-table">
          <thead>
            <tr>
              <th>ID</th>
              <th>{{ t('name') }}</th>
              <th>{{ t('channels') }}</th>
              <th>{{ t('epg_channels') }}</th>
              <th>{{ t('autoupdate') }}</th>
              <th>{{ t('interval_min') }}</th>
              <th>{{ t('last_sync') }}</th>
              <th>{{ t('last_error') }}</th>
              <th>{{ t('actions') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="p in playlists" :key="p.id">
              <td>{{ p.id }}</td>
              <td>{{ p.name }}</td>
              <td>{{ p.channel_count ?? 0 }}</td>
              <td>{{ p.epg_channel_count ?? 0 }}</td>
              <td>
                <span class="status-chip" :class="{ on: p.enabled, off: !p.enabled }">
                  {{ p.enabled ? t('enabled') : t('disabled') }}
                </span>
              </td>
              <td>{{ p.update_interval_minutes }}</td>
              <td>{{ formatSyncDateTime(p.last_sync_at) }}</td>
              <td class="last-error-cell" :title="p.last_error || '-'">{{ p.last_error || '-' }}</td>
              <td class="row-actions">
                <button class="btn subtle" :disabled="refreshingPlaylistIds[p.id]" @click="onRefresh(p.id)">
                  {{ refreshingPlaylistIds[p.id] ? t('updating') : t('refresh') }}
                </button>
                <button class="btn subtle" @click="onStartEdit(p)">{{ t('edit') }}</button>
                <button class="btn danger" @click="onRemovePlaylist(p)">{{ t('delete') }}</button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <div v-if="playlists.length > 0" class="settings-block playlists-hidden-settings">
        <h3>{{ t('hide_channels_groups') }}</h3>
        <div class="playlist-switcher-list">
          <button
            v-for="p in playlists"
            :key="`settings-hidden-playlist-${p.id}`"
            type="button"
            class="btn playlist-btn"
            :class="{ active: selectedPlaylist === p.id }"
            @click="onSelectPlaylist(p.id)"
          >
            {{ p.name }}
          </button>
        </div>
        <div class="settings-channel-groups">
          <section v-for="group in settingsGroupedChannels" :key="`settings-group-${group.name}`" class="group-block">
            <div class="group-head">
              <button type="button" class="group-toggle" @click="onToggleSettingsGroup(group.name)">
                <span>{{ isSettingsGroupOpen(group.name) ? '▾' : '▸' }}</span>
                <h4>{{ group.name }} ({{ group.channels.length }})</h4>
              </button>
              <label class="hidden-setting-row">
                <input type="checkbox" :checked="isHiddenGroup(group.name)" @change="onToggleHiddenGroup(group.name)" />
                <span>{{ t('hide_group') }}</span>
              </label>
            </div>
            <div v-show="isSettingsGroupOpen(group.name)" class="group-channels">
              <label v-for="c in group.channels" :key="`settings-channel-${c.id}`" class="hidden-setting-row">
                <input type="checkbox" :checked="isHiddenChannel(c.playlist_id, c.id, c.external_id)" @change="onToggleHiddenChannel(c)" />
                <span>{{ c.name }}</span>
              </label>
            </div>
          </section>
          <div v-if="settingsGroupedChannels.length === 0" class="muted">
            {{ t('select_playlist_first') }}
          </div>
        </div>
      </div>
    </template>

    <template v-else>
      <label class="checkbox debug-checkbox">
        <input :checked="showDebugOverlay" type="checkbox" @change="onToggleDebugOverlay" />
        {{ t('show_debug_overlay') }}
      </label>
      <div class="interface-settings-grid">
        <div class="settings-block playlists-theme-settings">
          <h3>{{ t('interface_language') }}</h3>
          <div class="theme-options" role="radiogroup" :aria-label="t('interface_language')">
            <label class="theme-option">
              <input :checked="uiLanguageMode === 'system'" type="radio" value="system" @change="onLanguageModeChange('system')" />
              <span>{{ t('lang_system') }}</span>
            </label>
            <label class="theme-option">
              <input :checked="uiLanguageMode === 'en'" type="radio" value="en" @change="onLanguageModeChange('en')" />
              <span>{{ t('lang_english') }}</span>
            </label>
            <label class="theme-option">
              <input :checked="uiLanguageMode === 'ru'" type="radio" value="ru" @change="onLanguageModeChange('ru')" />
              <span>{{ t('lang_russian') }}</span>
            </label>
          </div>
        </div>
        <div class="settings-block playlists-theme-settings">
          <h3>{{ t('interface_theme') }}</h3>
          <div class="theme-options" role="radiogroup" :aria-label="t('interface_theme')">
            <label class="theme-option">
              <input :checked="themeMode === 'light'" type="radio" value="light" @change="onThemeModeChange('light')" />
              <span>{{ t('theme_light') }}</span>
            </label>
            <label class="theme-option">
              <input :checked="themeMode === 'dark'" type="radio" value="dark" @change="onThemeModeChange('dark')" />
              <span>{{ t('theme_dark') }}</span>
            </label>
            <label class="theme-option">
              <input :checked="themeMode === 'system'" type="radio" value="system" @change="onThemeModeChange('system')" />
              <span>{{ t('theme_system') }}</span>
            </label>
          </div>
        </div>
        <div class="settings-block playlists-theme-settings">
          <h3>{{ t('default_audio_language') }}</h3>
          <label class="field-label">
            <span>{{ t('default_audio_language') }}</span>
            <select class="form-input" :value="defaultAudioLanguage" @change="onDefaultAudioLanguageChange($event.target.value)">
              <option v-for="item in defaultAudioLanguageOptions" :key="item.value" :value="item.value">
                {{ item.label }}
              </option>
            </select>
          </label>
        </div>
      </div>
    </template>
  </section>
</template>

<script setup>
import { ref } from 'vue'

const settingsTab = ref('playlists')

defineProps({
  tab: { type: String, required: true },
  t: { type: Function, required: true },
  form: { type: Object, required: true },
  editingId: { type: Number, required: true },
  playlists: { type: Array, required: true },
  refreshingPlaylistIds: { type: Object, required: true },
  showDebugOverlay: { type: Boolean, required: true },
  uiLanguageMode: { type: String, required: true },
  themeMode: { type: String, required: true },
  defaultAudioLanguage: { type: String, required: true },
  defaultAudioLanguageOptions: { type: Array, required: true },
  selectedPlaylist: { type: Number, required: true },
  settingsGroupedChannels: { type: Array, required: true },
  formatSyncDateTime: { type: Function, required: true },
  isSettingsGroupOpen: { type: Function, required: true },
  isHiddenGroup: { type: Function, required: true },
  isHiddenChannel: { type: Function, required: true },
  onClose: { type: Function, required: true },
  onSavePlaylist: { type: Function, required: true },
  onCancelEdit: { type: Function, required: true },
  onRefresh: { type: Function, required: true },
  onStartEdit: { type: Function, required: true },
  onRemovePlaylist: { type: Function, required: true },
  onToggleDebugOverlay: { type: Function, required: true },
  onLanguageModeChange: { type: Function, required: true },
  onThemeModeChange: { type: Function, required: true },
  onDefaultAudioLanguageChange: { type: Function, required: true },
  onSelectPlaylist: { type: Function, required: true },
  onToggleSettingsGroup: { type: Function, required: true },
  onToggleHiddenGroup: { type: Function, required: true },
  onToggleHiddenChannel: { type: Function, required: true }
})
</script>
