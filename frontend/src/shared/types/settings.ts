export type AppTheme = 'dark' | 'light' | 'system'
export type AccountStatus = 'online' | 'working' | 'dnd' | 'offline'
export type AppLocale = 'en-US' | 'zh-CN'

export interface AccountSettings {
  displayName: string
  email: string
  title: string
  avatar: string
  timezone: string
  status: AccountStatus
  statusMessage: string
}

export interface NotificationSettings {
  desktop: boolean
  sound: boolean
  mentionsOnly: boolean
  previews: boolean
  quietHoursEnabled: boolean
  quietHoursStart: string
  quietHoursEnd: string
}

export interface KeybindSettings {
  enabled: boolean
  showHints: boolean
  profile: 'default' | 'vim' | 'emacs'
}

export interface ChatSettings {
  streamOutput: boolean
}

export interface AppearanceSettings {
  theme: AppTheme
}

export interface TerminalSettings {
  fontSize: number
  fontFamily: string
  cursorStyle: 'block' | 'underline' | 'bar'
  cursorBlink: boolean
  scrollback: number
  bellSound: boolean
}

export interface Settings {
  appearance: AppearanceSettings
  locale: AppLocale
  account: AccountSettings
  notifications: NotificationSettings
  keybinds: KeybindSettings
  chat: ChatSettings
  terminal: TerminalSettings
}

export const DEFAULT_SETTINGS: Settings = {
  appearance: {
    theme: 'dark'
  },
  locale: 'en-US',
  account: {
    displayName: '',
    email: '',
    title: '',
    avatar: '',
    timezone: 'utc',
    status: 'online',
    statusMessage: ''
  },
  notifications: {
    desktop: true,
    sound: false,
    mentionsOnly: false,
    previews: true,
    quietHoursEnabled: false,
    quietHoursStart: '22:00',
    quietHoursEnd: '07:00'
  },
  keybinds: {
    enabled: true,
    showHints: true,
    profile: 'default'
  },
  chat: {
    streamOutput: true
  },
  terminal: {
    fontSize: 14,
    fontFamily: 'monospace',
    cursorStyle: 'block',
    cursorBlink: true,
    scrollback: 10000,
    bellSound: false
  }
}