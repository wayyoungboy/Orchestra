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

// Always apply dark theme
document.documentElement.classList.add('dark-theme')

app.use(i18n)

const settingsStore = useSettingsStore()
watch(
  () => settingsStore.locale,
  (loc) => {
    i18n.global.locale.value = settingsLocaleToI18n(loc)
  },
  { immediate: true }
)

app.use(router)
app.mount('#app')
