<template>
  <div class="h-screen w-screen overflow-hidden">
    <router-view />
    <ContextMenuHost />
    <ToastStack />
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useAuthStore } from '@/features/auth/authStore'
import ContextMenuHost from '@/shared/context-menu/ContextMenuHost.vue'
import ToastStack from '@/shared/components/ToastStack.vue'

const authStore = useAuthStore()

onMounted(async () => {
  await authStore.fetchConfig()
})
</script>