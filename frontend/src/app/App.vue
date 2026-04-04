<template>
  <div class="h-screen w-screen bg-background text-white overflow-hidden">
    <router-view />
    <ContextMenuHost />
    <ToastStack />
  </div>
</template>

<script setup lang="ts">
import ContextMenuHost from '@/shared/context-menu/ContextMenuHost.vue'
import ToastStack from '@/shared/components/ToastStack.vue'
</script>