<template>
  <header class="chat-header-root">
    <div class="header-left">
      <h1 class="header-title">{{ title }}</h1>
      <p v-if="description" class="header-sub">{{ description }}</p>
    </div>

    <div class="header-right">
      <button
        v-if="memberCount !== undefined"
        type="button"
        class="members-btn"
        @click="$emit('open-members')"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
        </svg>
        <span class="member-count">{{ memberCount }}</span>
      </button>
    </div>
  </header>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

defineProps<{
  title: string
  description?: string
  memberCount?: number
}>()

defineEmits<{
  (e: 'open-members'): void
}>()
</script>

<style scoped>
.chat-header-root {
  width: 100%;
  height: 72px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  background: rgba(255, 255, 255, 0.5);
  backdrop-filter: blur(32px);
  -webkit-backdrop-filter: blur(32px);
  border-radius: 24px;
  border: 1px solid white;
  box-shadow: 0 20px 50px rgba(0, 0, 0, 0.02);
  margin-bottom: 16px;
}

.header-left {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.header-title {
  font-size: 18px;
  font-weight: 900;
  color: #0f172a;
}

.header-sub {
  font-size: 12px;
  font-weight: 500;
  color: #64748b;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 400px;
}

.header-right {
  display: flex;
  align-items: center;
}

.members-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 12px;
  border-radius: 12px;
  border: 1px solid #e2e8f0;
  background: white;
  color: #64748b;
  transition: all 0.2s;
}

.members-btn:hover {
  background: #f8fafc;
  color: #0f172a;
  border-color: #cbd5e1;
}

.member-count {
  font-size: 12px;
  font-weight: 700;
  color: #475569;
}
</style>
