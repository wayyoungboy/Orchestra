// 全局设置管理：负责账户/通知/主题/终端偏好与应用级持久化
import { computed, ref, watch } from 'vue'
import { acceptHMRUpdate, defineStore } from 'pinia'
import type {
  Settings,
  AppTheme,
  AppLocale,
  AccountStatus
} from '@/shared/types/settings'
import { DEFAULT_SETTINGS } from '@/shared/types/settings'
import { getApiErrorMessage } from '@/shared/api/errors'
import { notifyUserError } from '@/shared/notifyError'

const SETTINGS_STORAGE_KEY = 'orchestra-settings'

const ALLOWED_THEMES = new Set<AppTheme>(['dark', 'light', 'system'])
const ALLOWED_LOCALES = new Set<AppLocale>(['en-US', 'zh-CN'])
const ALLOWED_STATUSES = new Set<AccountStatus>(['online', 'working', 'dnd', 'offline'])

const cloneSettings = (value: Settings): Settings => JSON.parse(JSON.stringify(value)) as Settings

const normalizeTheme = (value: unknown): AppTheme => {
  if (typeof value === 'string' && ALLOWED_THEMES.has(value as AppTheme)) {
    return value as AppTheme
  }
  return DEFAULT_SETTINGS.appearance.theme
}

const normalizeLocale = (value: unknown): AppLocale => {
  if (typeof value === 'string' && ALLOWED_LOCALES.has(value as AppLocale)) {
    return value as AppLocale
  }
  return DEFAULT_SETTINGS.locale
}

const normalizeAccountStatus = (value: unknown): AccountStatus => {
  if (ALLOWED_STATUSES.has(value as AccountStatus)) {
    return value as AccountStatus
  }
  return DEFAULT_SETTINGS.account.status
}

const normalizeSettings = (candidate: Partial<Settings>): Settings => {
  return {
    appearance: {
      theme: normalizeTheme(candidate.appearance?.theme)
    },
    locale: normalizeLocale(candidate.locale),
    account: {
      displayName: candidate.account?.displayName?.trim() ?? '',
      email: candidate.account?.email?.trim().toLowerCase() ?? '',
      title: candidate.account?.title?.trim() ?? '',
      avatar: candidate.account?.avatar ?? '',
      timezone: candidate.account?.timezone?.trim() ?? 'utc',
      status: normalizeAccountStatus(candidate.account?.status),
      statusMessage: candidate.account?.statusMessage?.trim() ?? ''
    },
    notifications: {
      desktop: Boolean(candidate.notifications?.desktop ?? DEFAULT_SETTINGS.notifications.desktop),
      sound: Boolean(candidate.notifications?.sound ?? DEFAULT_SETTINGS.notifications.sound),
      mentionsOnly: Boolean(candidate.notifications?.mentionsOnly ?? DEFAULT_SETTINGS.notifications.mentionsOnly),
      previews: Boolean(candidate.notifications?.previews ?? DEFAULT_SETTINGS.notifications.previews),
      quietHoursEnabled: Boolean(candidate.notifications?.quietHoursEnabled ?? DEFAULT_SETTINGS.notifications.quietHoursEnabled),
      quietHoursStart: candidate.notifications?.quietHoursStart ?? DEFAULT_SETTINGS.notifications.quietHoursStart,
      quietHoursEnd: candidate.notifications?.quietHoursEnd ?? DEFAULT_SETTINGS.notifications.quietHoursEnd
    },
    keybinds: {
      enabled: Boolean(candidate.keybinds?.enabled ?? DEFAULT_SETTINGS.keybinds.enabled),
      showHints: Boolean(candidate.keybinds?.showHints ?? DEFAULT_SETTINGS.keybinds.showHints),
      profile: candidate.keybinds?.profile ?? DEFAULT_SETTINGS.keybinds.profile
    },
    chat: {
      streamOutput: Boolean(candidate.chat?.streamOutput ?? DEFAULT_SETTINGS.chat.streamOutput)
    }
  }
}

/**
 * 保存设置到 localStorage
 */
const persistSettings = (settings: Settings) => {
  try {
    localStorage.setItem(SETTINGS_STORAGE_KEY, JSON.stringify(settings))
  } catch (error) {
    notifyUserError('Save settings to localStorage', error)
  }
}

/**
 * 从 localStorage 读取设置
 */
const loadStoredSettings = (): Partial<Settings> | null => {
  try {
    const stored = localStorage.getItem(SETTINGS_STORAGE_KEY)
    if (stored) {
      return JSON.parse(stored) as Partial<Settings>
    }
  } catch (error) {
    notifyUserError('Load settings from localStorage', error)
  }
  return null
}

/**
 * 应用主题到 DOM
 */
const applyTheme = (theme: AppTheme) => {
  const root = document.documentElement
  if (theme === 'system') {
    const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches
    root.classList.toggle('dark', prefersDark)
  } else {
    root.classList.toggle('dark', theme === 'dark')
  }
}

/**
 * 设置状态存储
 * 输入：保存/重置/修改局部设置
 * 输出：当前设置与操作函数
 */
export const useSettingsStore = defineStore('settings', () => {
  const settingsRef = ref<Settings>(cloneSettings(DEFAULT_SETTINGS))
  const loadingSettings = ref(false)
  const loadedSettings = ref(false)
  const settingsError = ref<string | null>(null)

  const settings = computed(() => settingsRef.value)
  const theme = computed(() => settingsRef.value.appearance.theme)
  const locale = computed(() => settingsRef.value.locale)

  /**
   * 读取并初始化设置，仅执行一次
   */
  const hydrate = async () => {
    if (loadingSettings.value || loadedSettings.value) return
    loadingSettings.value = true
    settingsError.value = null

    try {
      const stored = loadStoredSettings()
      const normalized = normalizeSettings(stored ?? {})
      const previous = settingsRef.value
      settingsRef.value = normalized

      // 应用主题
      if (previous.appearance.theme !== normalized.appearance.theme) {
        applyTheme(normalized.appearance.theme)
      }

      // 如果没有存储过，保存默认设置
      if (!stored) {
        persistSettings(normalized)
      }

      loadedSettings.value = true
    } catch (error) {
      settingsError.value = getApiErrorMessage(error)
      notifyUserError('Hydrate settings', error)
    } finally {
      loadingSettings.value = false
    }
  }

  /**
   * 保存完整设置并持久化
   */
  const saveSettings = (next: Settings) => {
    const previous = settingsRef.value
    const normalized = normalizeSettings(next)
    settingsRef.value = normalized
    persistSettings(normalized)

    // 应用主题变化
    if (previous.appearance.theme !== normalized.appearance.theme) {
      applyTheme(normalized.appearance.theme)
    }

    return normalized
  }

  /**
   * 恢复默认设置并持久化
   */
  const resetSettings = () => {
    const previous = settingsRef.value
    const next = cloneSettings(DEFAULT_SETTINGS)
    settingsRef.value = next
    persistSettings(next)

    // 应用主题
    if (previous.appearance.theme !== next.appearance.theme) {
      applyTheme(next.appearance.theme)
    }

    return next
  }

  /**
   * 切换主题并同步到 DOM
   */
  const setTheme = (next: AppTheme) => {
    const normalized = normalizeTheme(next)
    if (settingsRef.value.appearance.theme === normalized) {
      applyTheme(normalized)
      return
    }
    const updated: Settings = {
      ...settingsRef.value,
      appearance: {
        ...settingsRef.value.appearance,
        theme: normalized
      }
    }
    saveSettings(updated)
  }

  /**
   * 切换语言
   */
  const setLocale = (next: AppLocale) => {
    const normalized = normalizeLocale(next)
    if (settingsRef.value.locale === normalized) {
      return
    }
    const updated: Settings = {
      ...settingsRef.value,
      locale: normalized
    }
    saveSettings(updated)
  }

  /**
   * 设置账户显示名称
   */
  const setAccountDisplayName = (displayName: string) => {
    const nextName = displayName.trim()
    if (settingsRef.value.account.displayName === nextName) {
      return settingsRef.value
    }
    const next: Settings = {
      ...settingsRef.value,
      account: {
        ...settingsRef.value.account,
        displayName: nextName
      }
    }
    return saveSettings(next)
  }

  /**
   * 设置账户状态
   */
  const setAccountStatus = (status: AccountStatus) => {
    if (settingsRef.value.account.status === status) {
      return settingsRef.value
    }
    const next: Settings = {
      ...settingsRef.value,
      account: {
        ...settingsRef.value.account,
        status
      }
    }
    return saveSettings(next)
  }

  /**
   * 设置通知偏好
   */
  const setNotifications = (notifications: Partial<typeof DEFAULT_SETTINGS.notifications>) => {
    const next: Settings = {
      ...settingsRef.value,
      notifications: {
        ...settingsRef.value.notifications,
        ...notifications
      }
    }
    return saveSettings(next)
  }

  /**
   * 设置聊天配置
   */
  const setChatConfig = (config: Partial<typeof DEFAULT_SETTINGS.chat>) => {
    const next: Settings = {
      ...settingsRef.value,
      chat: {
        ...settingsRef.value.chat,
        ...config
      }
    }
    return saveSettings(next)
  }

  // 监听系统主题变化
  watch(theme, (newTheme) => {
    if (newTheme === 'system') {
      const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
      const handler = (e: MediaQueryListEvent) => {
        document.documentElement.classList.toggle('dark', e.matches)
      }
      mediaQuery.addEventListener('change', handler)
    }
  })

  // 初始化时自动加载
  void hydrate()

  return {
    settings,
    loadingSettings,
    loadedSettings,
    settingsError,
    theme,
    locale,
    hydrate,
    saveSettings,
    resetSettings,
    setTheme,
    setLocale,
    setAccountDisplayName,
    setAccountStatus,
    setNotifications,
    setChatConfig
  }
})

if (import.meta.hot) {
  import.meta.hot.accept(acceptHMRUpdate(useSettingsStore, import.meta.hot))
}