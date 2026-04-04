<template>
  <div class="h-full flex flex-col items-center justify-center px-8 text-center max-w-lg mx-auto">
    <div class="w-16 h-16 rounded-2xl bg-primary/15 flex items-center justify-center mb-6">
      <span class="text-2xl font-bold text-primary">S</span>
    </div>
    <h1 class="text-xl font-bold text-white mb-3">{{ t('skills.title') }}</h1>
    <p class="text-sm text-white/60 leading-relaxed">
      {{ t('skills.webNote') }}
    </p>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
</script>
