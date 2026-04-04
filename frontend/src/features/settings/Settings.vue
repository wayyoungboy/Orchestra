<template>
  <div class="flex h-full w-full">
    <!-- Settings Sidebar -->
    <aside class="w-64 bg-panel/80 border-r border-white/5 flex flex-col">
      <div class="p-4 border-b border-white/5">
        <h2 class="text-lg font-bold text-white">{{ t('settings.title') }}</h2>
      </div>
      <nav class="flex-1 p-2">
        <button
          v-for="section in sections"
          :key="section.id"
          type="button"
          @click="activeSection = section.id"
          :class="[
            'w-full flex items-center gap-3 px-4 py-3 rounded-xl transition-colors',
            activeSection === section.id
              ? 'bg-primary/20 text-white'
              : 'text-white/60 hover:bg-white/5 hover:text-white'
          ]"
        >
          <component :is="section.icon" class="w-5 h-5" />
          <span>{{ section.label }}</span>
        </button>
      </nav>
    </aside>

    <!-- Settings Content -->
    <main class="flex-1 overflow-y-auto p-6">
      <div class="max-w-2xl">
        <!-- General Settings -->
        <div v-if="activeSection === 'general'">
          <h3 class="text-xl font-bold text-white mb-6">{{ t('settings.general') }}</h3>

          <div class="space-y-6">
            <div class="bg-panel/40 rounded-xl p-4 border border-white/5">
              <h4 class="text-sm font-bold text-white mb-4">Account</h4>
              <div class="space-y-4">
                <div>
                  <label class="block text-xs text-white/60 mb-2">Display Name</label>
                  <input
                    v-model="settings.displayName"
                    type="text"
                    class="w-full bg-surface text-white rounded-lg px-4 py-2 border border-white/10 focus:border-primary/50 focus:outline-none"
                  />
                </div>
              </div>
            </div>

            <div class="bg-panel/40 rounded-xl p-4 border border-white/5">
              <h4 class="text-sm font-bold text-white mb-4">Theme</h4>
              <div class="flex gap-3">
                <button
                  v-for="theme in themes"
                  :key="theme.id"
                  type="button"
                  @click="settings.theme = theme.id"
                  :class="[
                    'flex-1 p-3 rounded-lg border transition-colors',
                    settings.theme === theme.id
                      ? 'bg-primary/20 border-primary/50 text-white'
                      : 'bg-white/5 border-white/10 text-white/60 hover:border-white/20'
                  ]"
                >
                  {{ theme.label }}
                </button>
              </div>
            </div>

            <div class="bg-panel/40 rounded-xl p-4 border border-white/5">
              <h4 class="text-sm font-bold text-white mb-4">{{ t('settings.language') }}</h4>
              <select
                :value="settingsStore.locale"
                class="w-full bg-surface text-white rounded-lg px-4 py-2 border border-white/10 focus:border-primary/50 focus:outline-none"
                @change="onLocaleChange"
              >
                <option value="en-US">English</option>
                <option value="zh-CN">中文</option>
              </select>
            </div>
          </div>
        </div>

        <!-- Terminal Settings -->
        <div v-if="activeSection === 'terminal'">
          <h3 class="text-xl font-bold text-white mb-6">Terminal</h3>

          <div class="space-y-6">
            <div class="bg-panel/40 rounded-xl p-4 border border-white/5">
              <h4 class="text-sm font-bold text-white mb-4">Appearance</h4>
              <div class="space-y-4">
                <div>
                  <label class="block text-xs text-white/60 mb-2">Font Size</label>
                  <input
                    v-model.number="terminalSettings.fontSize"
                    type="number"
                    min="10"
                    max="24"
                    class="w-full bg-surface text-white rounded-lg px-4 py-2 border border-white/10 focus:border-primary/50 focus:outline-none"
                  />
                </div>
                <div>
                  <label class="block text-xs text-white/60 mb-2">Font Family</label>
                  <input
                    v-model="terminalSettings.fontFamily"
                    type="text"
                    class="w-full bg-surface text-white rounded-lg px-4 py-2 border border-white/10 focus:border-primary/50 focus:outline-none"
                  />
                </div>
              </div>
            </div>

            <div class="bg-panel/40 rounded-xl p-4 border border-white/5">
              <h4 class="text-sm font-bold text-white mb-4">Shell</h4>
              <div>
                <label class="block text-xs text-white/60 mb-2">Default Shell</label>
                <select
                  v-model="terminalSettings.shell"
                  class="w-full bg-surface text-white rounded-lg px-4 py-2 border border-white/10 focus:border-primary/50 focus:outline-none"
                >
                  <option value="/bin/bash">/bin/bash</option>
                  <option value="/bin/zsh">/bin/zsh</option>
                  <option value="/bin/sh">/bin/sh</option>
                </select>
              </div>
            </div>
          </div>
        </div>

        <!-- Workspace Settings -->
        <div v-if="activeSection === 'workspace'">
          <h3 class="text-xl font-bold text-white mb-6">Workspace</h3>

          <div class="space-y-6">
            <div class="bg-panel/40 rounded-xl p-4 border border-white/5">
              <h4 class="text-sm font-bold text-white mb-4">Current Workspace</h4>
              <div class="space-y-2">
                <div class="flex justify-between text-sm">
                  <span class="text-white/60">Name</span>
                  <span class="text-white">{{ workspaceStore.currentWorkspace?.name || '-' }}</span>
                </div>
                <div class="flex justify-between text-sm">
                  <span class="text-white/60">Path</span>
                  <span class="text-white/80 font-mono text-xs">{{ workspaceStore.currentWorkspace?.path || '-' }}</span>
                </div>
              </div>
            </div>

            <div class="bg-panel/40 rounded-xl p-4 border border-white/5">
              <h4 class="text-sm font-bold text-white mb-4">Members</h4>
              <div class="space-y-2">
                <div
                  v-for="member in projectStore.members"
                  :key="member.id"
                  class="flex items-center justify-between p-2 rounded-lg bg-white/5"
                >
                  <div class="flex items-center gap-3">
                    <div class="w-8 h-8 rounded-full bg-primary/20 flex items-center justify-center">
                      <span class="text-xs font-bold text-white">{{ member.name.charAt(0).toUpperCase() }}</span>
                    </div>
                    <div>
                      <p class="text-sm text-white">{{ member.name }}</p>
                      <p class="text-[10px] text-white/40">{{ member.roleType }}</p>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Diagnostics (REQ-404) -->
        <div v-if="activeSection === 'diagnostics'">
          <h3 class="text-xl font-bold text-white mb-2">{{ t('settings.diagnostics') }}</h3>
          <p class="text-sm text-white/50 mb-6">{{ t('settings.diagnosticsHint') }}</p>
          <div class="bg-panel/40 rounded-xl p-4 border border-white/5">
            <div class="flex justify-end mb-2">
              <button
                type="button"
                class="text-xs font-medium text-primary hover:text-primary/80"
                @click="diagnosticsStore.clear()"
              >
                {{ t('settings.clearLog') }}
              </button>
            </div>
            <pre
              class="text-[11px] font-mono text-white/70 whitespace-pre-wrap break-all max-h-80 overflow-y-auto bg-black/30 rounded-lg p-3 border border-white/5"
              >{{ diagnosticsText }}</pre
            >
          </div>
        </div>

        <!-- About -->
        <div v-if="activeSection === 'about'">
          <h3 class="text-xl font-bold text-white mb-6">About</h3>

          <div class="bg-panel/40 rounded-xl p-6 border border-white/5">
            <div class="text-center">
              <div class="w-16 h-16 rounded-2xl bg-primary/20 flex items-center justify-center mx-auto mb-4">
                <span class="text-2xl font-bold text-primary">O</span>
              </div>
              <h4 class="text-lg font-bold text-white mb-2">Orchestra</h4>
              <p class="text-sm text-white/60 mb-4">Multi-Agent Development Platform</p>
              <p class="text-xs text-white/40">Version 0.1.0</p>
            </div>
          </div>
        </div>

        <!-- Account Security -->
        <div v-if="activeSection === 'security'">
          <h3 class="text-xl font-bold text-white mb-6">Account Security</h3>

          <div class="bg-panel/40 rounded-xl p-4 border border-white/5">
            <div class="space-y-4">
              <div>
                <label class="block text-xs text-white/60 mb-2">New Username</label>
                <input
                  v-model="newUsername"
                  type="text"
                  class="w-full bg-surface text-white rounded-lg px-4 py-2 border border-white/10 focus:border-primary/50 focus:outline-none"
                />
              </div>
              <div>
                <label class="block text-xs text-white/60 mb-2">New Password</label>
                <input
                  v-model="newPassword"
                  type="password"
                  class="w-full bg-surface text-white rounded-lg px-4 py-2 border border-white/10 focus:border-primary/50 focus:outline-none"
                />
              </div>
              <button
                type="button"
                @click="saveCredentials"
                class="w-full py-2 bg-primary text-on-primary font-medium rounded-lg text-sm hover:bg-primary/80 transition-colors"
              >
                Save Changes
              </button>
              <p v-if="credentialsSaved" class="text-xs text-green-400 text-center">Credentials updated successfully</p>
            </div>
          </div>
        </div>

        <!-- Keyboard Shortcuts -->
        <div v-if="activeSection === 'keyboard'">
          <h3 class="text-xl font-bold text-white mb-6">Keyboard Shortcuts</h3>

          <div class="space-y-6">
            <div class="bg-panel/40 rounded-xl p-4 border border-white/5">
              <h4 class="text-sm font-bold text-white mb-4">General</h4>
              <div class="space-y-4">
                <div class="flex items-center justify-between">
                  <span class="text-sm text-white/60">Enable keyboard shortcuts</span>
                  <button
                    type="button"
                    @click="keyboardSettings.enabled = !keyboardSettings.enabled"
                    :class="[
                      'relative w-10 h-6 rounded-full transition-colors',
                      keyboardSettings.enabled ? 'bg-primary' : 'bg-white/20'
                    ]"
                  >
                    <span
                      :class="[
                        'absolute top-1 w-4 h-4 rounded-full bg-white transition-transform',
                        keyboardSettings.enabled ? 'left-5' : 'left-1'
                      ]"
                    />
                  </button>
                </div>
                <div class="flex items-center justify-between">
                  <span class="text-sm text-white/60">Show shortcut hints</span>
                  <button
                    type="button"
                    @click="keyboardSettings.showHints = !keyboardSettings.showHints"
                    :class="[
                      'relative w-10 h-6 rounded-full transition-colors',
                      keyboardSettings.showHints ? 'bg-primary' : 'bg-white/20'
                    ]"
                  >
                    <span
                      :class="[
                        'absolute top-1 w-4 h-4 rounded-full bg-white transition-transform',
                        keyboardSettings.showHints ? 'left-5' : 'left-1'
                      ]"
                    />
                  </button>
                </div>
              </div>
            </div>

            <div class="bg-panel/40 rounded-xl p-4 border border-white/5">
              <h4 class="text-sm font-bold text-white mb-4">Available Shortcuts</h4>
              <div class="space-y-3">
                <div
                  v-for="shortcut in keyboardShortcutsMerged"
                  :key="shortcut.id"
                  class="flex items-center justify-between py-2 px-3 rounded-lg bg-white/5"
                >
                  <span class="text-sm text-white/80">{{ shortcut.description }}</span>
                  <kbd class="px-2 py-1 rounded bg-surface border border-white/10 text-xs font-mono text-white/60">
                    {{ shortcut.displayKey }}
                  </kbd>
                </div>
              </div>
            </div>

            <div class="bg-panel/40 rounded-xl p-4 border border-white/5">
              <h4 class="text-sm font-bold text-white mb-3">Tip</h4>
              <p class="text-sm text-white/60">
                Press <kbd class="px-1.5 py-0.5 rounded bg-surface border border-white/10 text-xs font-mono">Ctrl+/</kbd>
                anywhere in the app to toggle the shortcuts help panel.
              </p>
            </div>
          </div>
        </div>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useWorkspaceStore } from '@/features/workspace/workspaceStore'
import { useProjectStore } from '@/features/workspace/projectStore'
import { useAuthStore } from '@/features/auth/authStore'
import { useSettingsStore } from '@/features/settings/settingsStore'
import { useDiagnosticsStore } from '@/stores/diagnosticsStore'
import { getAppShortcuts } from '@/shared/composables'
import {
  getRegisteredShortcutsSnapshot,
  formatShortcutDefinitionForDisplay
} from '@/shared/composables/useKeyboard'
import type { AppLocale } from '@/shared/types/settings'

const { t } = useI18n()
const workspaceStore = useWorkspaceStore()
const projectStore = useProjectStore()
const authStore = useAuthStore()
const settingsStore = useSettingsStore()
const diagnosticsStore = useDiagnosticsStore()

const activeSection = ref('general')

const sections = computed(() => [
  { id: 'general', label: t('settings.general'), icon: 'GeneralIcon' },
  { id: 'terminal', label: t('settings.terminal'), icon: 'TerminalIcon' },
  { id: 'workspace', label: t('settings.workspace'), icon: 'WorkspaceIcon' },
  { id: 'keyboard', label: t('settings.keyboard'), icon: 'KeyboardIcon' },
  { id: 'diagnostics', label: t('settings.diagnostics'), icon: 'DiagnosticsIcon' },
  { id: 'security', label: t('settings.security'), icon: 'SecurityIcon' },
  { id: 'about', label: t('settings.about'), icon: 'AboutIcon' },
])

const themes = [
  { id: 'dark', label: 'Dark' },
  { id: 'light', label: 'Light' },
  { id: 'system', label: 'System' },
]

const settings = reactive({
  displayName: 'User',
  theme: 'dark',
})

function onLocaleChange(e: Event) {
  const v = (e.target as HTMLSelectElement).value as AppLocale
  settingsStore.setLocale(v)
}

const diagnosticsText = computed(() =>
  diagnosticsStore.recentLines.length ? diagnosticsStore.recentLines.join('\n') : '—'
)

const keyboardShortcutsMerged = computed(() => {
  const base = getAppShortcuts().map((s) => ({
    id: s.id,
    description: s.description,
    displayKey: s.displayKey
  }))
  const seen = new Set(base.map((b) => `${b.description}\0${b.displayKey}`))
  const snap = getRegisteredShortcutsSnapshot()
  const extra = snap
    .filter((d) => d.options.description)
    .map((d) => ({
      id: `dyn-${d.options.scope}-${d.key}-${formatShortcutDefinitionForDisplay(d)}`,
      description: d.options.description,
      displayKey: formatShortcutDefinitionForDisplay(d)
    }))
    .filter((r) => !seen.has(`${r.description}\0${r.displayKey}`))
  return [...base, ...extra]
})

const terminalSettings = reactive({
  fontSize: 14,
  fontFamily: 'JetBrains Mono',
  shell: '/bin/bash',
})

// Keyboard shortcuts settings
const keyboardSettings = reactive({
  enabled: true,
  showHints: true,
})

// Security settings
const currentCreds = authStore.getCredentials()
const newUsername = ref(currentCreds.username)
const newPassword = ref(currentCreds.password)
const credentialsSaved = ref(false)

function saveCredentials() {
  authStore.updateCredentials(newUsername.value, newPassword.value)
  credentialsSaved.value = true
  setTimeout(() => {
    credentialsSaved.value = false
  }, 2000)
}
</script>