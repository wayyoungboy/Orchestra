import { watch } from 'vue'
import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import { i18n, settingsLocaleToI18n } from '@/i18n'
import { useSettingsStore } from '@/features/settings/settingsStore'
import '../assets/main.css'

const app = createApp(App)
const pinia = createPinia()
app.use(pinia)

// Apply theme immediately from stored settings before mounting
const applyThemeOnStartup = () => {
  try {
    const stored = localStorage.getItem('orchestra-settings')
    if (stored) {
      const settings = JSON.parse(stored)
      const theme = settings?.appearance?.theme
      const root = document.documentElement
      root.classList.remove('dark-theme')
      if (theme === 'dark') {
        root.classList.add('dark-theme')
      } else if (theme === 'system') {
        const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches
        if (prefersDark) root.classList.add('dark-theme')
      }
    }
  } catch {
    // ignore
  }
}
applyThemeOnStartup()

app.use(i18n)

const settingsStore = useSettingsStore()
watch(
  () => settingsStore.locale,
  (loc) => {
    i18n.global.locale.value = settingsLocaleToI18n(loc)
  },
  { immediate: true }
)

// Keep theme class in sync when store updates
watch(
  () => settingsStore.theme,
  () => {}, // store's applyTheme handles DOM updates
  { immediate: true }
)

app.use(router)
app.mount('#app')