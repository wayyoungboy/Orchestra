<template>
  <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm">
    <div class="w-full max-w-sm bg-panel/90 border border-white/10 rounded-2xl shadow-2xl p-6">
      <div class="flex justify-between items-center mb-6">
        <h3 class="text-white font-bold text-lg">Create Conversation</h3>
        <button type="button" @click="$emit('close')" class="text-white/40 hover:text-white transition-colors">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <div class="space-y-4">
        <!-- Conversation Type Selection -->
        <div class="space-y-2">
          <label class="text-xs font-bold text-white/50 uppercase tracking-wider">Type</label>
          <div class="flex gap-2">
            <button
              type="button"
              @click="conversationType = 'channel'"
              :class="[
                'flex-1 py-2.5 rounded-xl font-medium transition-all flex items-center justify-center gap-2',
                conversationType === 'channel'
                  ? 'bg-primary/20 text-primary border border-primary/30'
                  : 'bg-white/5 text-white/60 border border-white/10 hover:bg-white/10'
              ]"
            >
              <span class="text-lg">#</span>
              <span class="text-sm">Channel</span>
            </button>
            <button
              type="button"
              @click="conversationType = 'dm'"
              :class="[
                'flex-1 py-2.5 rounded-xl font-medium transition-all flex items-center justify-center gap-2',
                conversationType === 'dm'
                  ? 'bg-primary/20 text-primary border border-primary/30'
                  : 'bg-white/5 text-white/60 border border-white/10 hover:bg-white/10'
              ]"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
              </svg>
              <span class="text-sm">Direct Message</span>
            </button>
          </div>
        </div>

        <!-- Name Input (for channels) -->
        <div v-if="conversationType === 'channel'" class="space-y-2">
          <label class="text-xs font-bold text-white/50 uppercase tracking-wider">Channel Name</label>
          <div class="relative">
            <span class="absolute left-3 top-1/2 -translate-y-1/2 text-white/30">#</span>
            <input
              v-model="conversationName"
              class="w-full bg-surface/80 border border-white/10 rounded-xl px-4 py-2.5 pl-8 text-white placeholder-white/30 focus:border-primary/50 outline-none transition-all"
              placeholder="channel-name"
            />
          </div>
        </div>

        <!-- Member Selection (for DMs) -->
        <div v-if="conversationType === 'dm'" class="space-y-2">
          <label class="text-xs font-bold text-white/50 uppercase tracking-wider">Select Member</label>
          <div class="space-y-2 max-h-48 overflow-y-auto">
            <button
              v-for="member in availableMembers"
              :key="member.id"
              type="button"
              @click="selectedMemberId = member.id"
              :class="[
                'w-full flex items-center gap-3 p-3 rounded-xl border transition-all',
                selectedMemberId === member.id
                  ? 'bg-primary/10 border-primary/20'
                  : 'border-transparent hover:bg-white/5'
              ]"
            >
              <div class="w-8 h-8 rounded-full bg-white/10 flex items-center justify-center">
                <span class="text-xs font-semibold text-white/60">{{ getInitials(member.name) }}</span>
              </div>
              <div class="flex flex-col min-w-0">
                <span :class="['text-sm font-medium truncate', selectedMemberId === member.id ? 'text-white' : 'text-white/70']">
                  {{ member.name }}
                </span>
                <span class="text-[10px] text-white/30 capitalize">{{ member.roleType }}</span>
              </div>
              <svg v-if="selectedMemberId === member.id" class="w-4 h-4 ml-auto text-primary" fill="currentColor" viewBox="0 0 20 20">
                <path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd" />
              </svg>
            </button>
          </div>
          <p v-if="availableMembers.length === 0" class="text-sm text-white/40 text-center py-4">
            No members available for direct message
          </p>
        </div>

        <!-- Action Buttons -->
        <div class="flex gap-3 pt-4">
          <button
            type="button"
            @click="$emit('close')"
            class="flex-1 py-2.5 bg-white/10 text-white font-medium rounded-xl hover:bg-white/15 transition-colors"
          >
            Cancel
          </button>
          <button
            type="button"
            @click="handleCreate"
            :disabled="!isValid"
            :class="[
              'flex-1 py-2.5 font-bold rounded-xl transition-colors',
              isValid
                ? 'bg-primary text-on-primary hover:bg-primary-hover'
                : 'bg-white/10 text-white/40 cursor-not-allowed'
            ]"
          >
            Create
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import type { Member } from '@/shared/types/member'

const props = defineProps<{
  members: Member[]
  currentUserId: string
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'create', data: { type: 'channel' | 'dm'; name?: string; memberId?: string }): void
}>()

const conversationType = ref<'channel' | 'dm'>('channel')
const conversationName = ref('')
const selectedMemberId = ref<string | null>(null)

const availableMembers = computed(() =>
  props.members.filter(m => m.id !== props.currentUserId)
)

const isValid = computed(() => {
  if (conversationType.value === 'channel') {
    return conversationName.value.trim().length > 0
  }
  return selectedMemberId.value !== null
})

function getInitials(name: string): string {
  return name.charAt(0).toUpperCase()
}

function handleCreate() {
  if (conversationType.value === 'channel') {
    emit('create', {
      type: 'channel',
      name: conversationName.value.trim()
    })
  } else if (selectedMemberId.value) {
    emit('create', {
      type: 'dm',
      memberId: selectedMemberId.value
    })
  }
}
</script>