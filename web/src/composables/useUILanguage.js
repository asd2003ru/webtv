import { computed, ref } from 'vue'
import { messages } from '../i18n/messages'

const STORAGE_KEY = 'webtv_ui_language_mode'

function systemLanguageCode() {
  const locale = (navigator.language || 'en').toLowerCase()
  return locale.startsWith('ru') ? 'ru' : 'en'
}

export function useUILanguage() {
  const uiLanguageMode = ref('system')

  const currentLanguage = computed(() => {
    if (uiLanguageMode.value === 'system') {
      return systemLanguageCode()
    }
    return uiLanguageMode.value === 'ru' ? 'ru' : 'en'
  })

  function t(key) {
    return messages[currentLanguage.value][key] || messages.ru[key] || key
  }

  function streamModeDisplay(mode) {
    if (mode === 'direct') return t('stream_mode_direct')
    if (mode === 'transcode') return t('stream_mode_transcode')
    return '-'
  }

  function onLanguageModeChange() {
    const nextMode = ['system', 'en', 'ru'].includes(uiLanguageMode.value) ? uiLanguageMode.value : 'system'
    uiLanguageMode.value = nextMode
    localStorage.setItem(STORAGE_KEY, nextMode)
  }

  function initLanguageMode() {
    const storedLanguageMode = localStorage.getItem(STORAGE_KEY)
    uiLanguageMode.value = ['system', 'en', 'ru'].includes(storedLanguageMode) ? storedLanguageMode : 'system'
  }

  return {
    uiLanguageMode,
    currentLanguage,
    t,
    streamModeDisplay,
    onLanguageModeChange,
    initLanguageMode
  }
}
