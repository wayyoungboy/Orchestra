<template>
  <div :class="[positionClass, baseClass]" data-invite-menu>
    <div class="px-3 py-2 mb-1">
      <h3 class="text-white font-semibold text-[13px] tracking-tight">{{ t('invite.menu.title') }}</h3>
      <p class="text-white/40 text-[11px] font-medium mt-0.5">{{ t('invite.menu.subtitle') }}</p>
    </div>
    <div class="h-px bg-white/5 mx-2 mb-1.5"></div>
    <button
      type="button"
      class="flex items-center gap-3 px-3 py-2 rounded-xl hover:bg-white/10 group transition-all duration-200 text-left w-full"
      @click="emit('select', 'admin')"
    >
      <div
        class="w-8 h-8 rounded-[10px] bg-red-500/10 text-red-400 flex items-center justify-center group-hover:bg-red-500 group-hover:text-on-primary group-hover:shadow-lg group-hover:shadow-red-500/20 transition-all duration-200"
      >
        <svg class="w-[18px] h-[18px]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z"
          />
        </svg>
      </div>
      <div class="flex flex-col min-w-0">
        <span class="text-white/90 text-[13px] font-medium group-hover:text-white tracking-tight">{{ t('invite.menu.admin') }}</span>
        <span class="text-white/40 text-[10px] group-hover:text-white/60">{{ t('invite.menu.adminDesc') }}</span>
      </div>
    </button>
    <button
      type="button"
      class="flex items-center gap-3 px-3 py-2 rounded-xl hover:bg-white/10 group transition-all duration-200 text-left w-full"
      @click="emit('select', 'secretary')"
    >
      <div
        class="w-8 h-8 rounded-[10px] bg-sky-500/10 text-sky-400 flex items-center justify-center group-hover:bg-sky-500 group-hover:text-on-primary group-hover:shadow-lg group-hover:shadow-sky-500/20 transition-all duration-200"
      >
        <svg class="w-[18px] h-[18px]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M15 12a3 3 0 11-6 0 3 3 0 016 0zM2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"
          />
        </svg>
      </div>
      <div class="flex flex-col min-w-0">
        <span class="text-white/90 text-[13px] font-medium group-hover:text-white tracking-tight">{{ t('invite.menu.secretary') }}</span>
        <span class="text-white/40 text-[10px] group-hover:text-white/60">{{ t('invite.menu.secretaryDesc') }}</span>
      </div>
    </button>
    <button
      type="button"
      class="flex items-center gap-3 px-3 py-2 rounded-xl hover:bg-white/10 group transition-all duration-200 text-left w-full"
      @click="emit('select', 'assistant')"
    >
      <div
        class="w-8 h-8 rounded-[10px] bg-yellow-500/10 text-yellow-400 flex items-center justify-center group-hover:bg-yellow-500 group-hover:text-on-primary group-hover:shadow-lg group-hover:shadow-yellow-500/20 transition-all duration-200"
      >
        <svg class="w-[18px] h-[18px]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"
          />
        </svg>
      </div>
      <div class="flex flex-col min-w-0">
        <span class="text-white/90 text-[13px] font-medium group-hover:text-white tracking-tight">{{ t('invite.menu.assistant') }}</span>
        <span class="text-white/40 text-[10px] group-hover:text-white/60">{{ t('invite.menu.assistantDesc') }}</span>
      </div>
    </button>
    <button
      type="button"
      class="flex items-center gap-3 px-3 py-2 rounded-xl hover:bg-white/10 group transition-all duration-200 text-left w-full"
      @click="emit('select', 'member')"
    >
      <div
        class="w-8 h-8 rounded-[10px] bg-primary/10 text-primary flex items-center justify-center group-hover:bg-primary group-hover:text-on-primary group-hover:shadow-lg group-hover:shadow-primary/20 transition-all duration-200"
      >
        <svg class="w-[18px] h-[18px]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"
          />
        </svg>
      </div>
      <div class="flex flex-col min-w-0">
        <span class="text-white/90 text-[13px] font-medium group-hover:text-white tracking-tight">{{ t('invite.menu.member') }}</span>
        <span class="text-white/40 text-[10px] group-hover:text-white/60">{{ t('invite.menu.memberDesc') }}</span>
      </div>
    </button>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

const props = defineProps<{ positionClass?: string }>()
const emit = defineEmits<{ (e: 'select', type: 'admin' | 'secretary' | 'assistant' | 'member'): void }>()

const { t } = useI18n()

const baseClass =
  'z-50 w-[280px] bg-panel/80 backdrop-blur-2xl border border-white/10 rounded-2xl shadow-2xl flex flex-col p-1.5 ring-1 ring-white/5'
const positionClass = computed(() => props.positionClass ?? 'fixed top-[74px] right-6')
</script>
