<template>
  <div class="min-h-screen flex items-center justify-center bg-gradient-to-br from-slate-900 via-slate-800 to-slate-900">
    <div class="w-full max-w-sm p-8">
      <div class="text-center mb-8">
        <div class="w-16 h-16 rounded-2xl bg-primary/20 flex items-center justify-center mx-auto mb-4">
          <span class="text-3xl font-bold text-primary">O</span>
        </div>
        <h1 class="text-2xl font-bold text-white">{{ t('app.name') }}</h1>
        <p class="text-white/40 text-sm mt-1">{{ t('auth.subtitle') }}</p>
      </div>

      <form @submit.prevent="handleLogin" class="space-y-4">
        <div>
          <label class="text-xs font-bold text-white/50 uppercase tracking-wider mb-1.5 block" for="orchestra-login-user">{{ t('auth.username') }}</label>
          <input id="orchestra-login-user" v-model="username" type="text" autocomplete="username" class="w-full bg-white/5 border border-white/10 rounded-xl px-4 py-3 text-white placeholder-white/30 focus:border-primary/50 outline-none transition-all" :placeholder="t('auth.usernamePlaceholder')" />
        </div>
        <div>
          <label class="text-xs font-bold text-white/50 uppercase tracking-wider mb-1.5 block" for="orchestra-login-pass">{{ t('auth.password') }}</label>
          <input id="orchestra-login-pass" v-model="password" type="password" autocomplete="current-password" class="w-full bg-white/5 border border-white/10 rounded-xl px-4 py-3 text-white placeholder-white/30 focus:border-primary/50 outline-none transition-all" :placeholder="t('auth.passwordPlaceholder')" />
        </div>
        <p v-if="error" class="text-red-400 text-sm text-center">{{ error }}</p>
        <button type="submit" class="w-full py-3 bg-primary text-on-primary font-bold rounded-xl hover:bg-primary-hover transition-colors">{{ t('auth.signIn') }}</button>
      </form>
      <p class="text-center text-white/30 text-xs mt-6">{{ t('auth.defaultCredentialsHint') }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from './authStore'

const { t } = useI18n()
const router = useRouter()
const authStore = useAuthStore()
const username = ref('')
const password = ref('')
const error = ref('')

function handleLogin() {
  error.value = ''
  if (authStore.login(username.value, password.value)) {
    router.push('/workspaces')
  } else {
    error.value = t('auth.invalidCredentials')
  }
}
</script>
