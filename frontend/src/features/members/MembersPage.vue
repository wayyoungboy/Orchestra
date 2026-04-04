<template>
  <div class="members-page-root animate-in fade-in zoom-in-95 duration-500">
    <div class="page-header">
      <div class="header-info">
        <h1 class="page-title">团队成员</h1>
        <p class="page-subtitle">{{ workspaceStore.currentWorkspace?.name }} · {{ members.length }} 位协作成员</p>
      </div>
      <button @click="showAddModal = true" class="add-member-btn">
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M12 4v16m8-8H4" />
        </svg>
        <span>邀请成员</span>
      </button>
    </div>

    <div class="members-content">
      <div class="members-grid">
        <div v-for="member in members" :key="member.id" class="member-card">
          <div class="card-top">
            <div :class="['member-avatar', member.roleType === 'owner' ? 'is-owner' : '']">
              {{ member.name.charAt(0).toUpperCase() }}
            </div>
            <div class="member-badge" :class="member.roleType">
              {{ roleLabel(member.roleType) }}
            </div>
          </div>
          
          <div class="member-info">
            <h3 class="member-name">{{ member.name }}</h3>
            <p class="member-id">ID: {{ member.id.slice(0, 8) }}...</p>
          </div>

          <div class="member-actions">
            <button @click="handleEdit(member)" class="action-btn">编辑</button>
            <button v-if="member.roleType !== 'owner'" @click="handleDelete(member.id)" class="action-btn is-danger">移除</button>
          </div>
        </div>
      </div>
    </div>

    <!-- Modals -->
    <AddMemberModal v-if="showAddModal" @close="showAddModal = false" @add="loadMembers" />
    <EditMemberModal v-if="editingMember" :member="editingMember" @close="editingMember = null" @update="loadMembers" />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useProjectStore } from '@/features/workspace/projectStore'
import { useWorkspaceStore } from '@/features/workspace/workspaceStore'
import AddMemberModal from './AddMemberModal.vue'
import EditMemberModal from './EditMemberModal.vue'

const projectStore = useProjectStore()
const workspaceStore = useWorkspaceStore()
const members = ref<any[]>([])
const showAddModal = ref(false)
const editingMember = ref<any>(null)

async function loadMembers() {
  if (workspaceStore.currentWorkspace) {
    const res = await projectStore.loadMembers(workspaceStore.currentWorkspace.id)
    members.value = res || []
  }
}

function roleLabel(role: string) {
  const map: any = { owner: '所有者', admin: '管理员', assistant: 'AI 助手', member: '成员' }
  return map[role] || role
}

function handleEdit(member: any) { editingMember.value = member }
async function handleDelete(id: string) {
  if (confirm('确定要移除该成员吗？')) {
    await projectStore.deleteMember(workspaceStore.currentWorkspace!.id, id)
    await loadMembers()
  }
}

onMounted(loadMembers)
</script>

<style scoped>
.members-page-root {
  height: 100%;
  display: flex;
  flex-direction: column;
  gap: 32px;
  padding: 24px;
}

.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.page-title { font-size: 28px; font-weight: 900; color: #0f172a; }
.page-subtitle { font-size: 14px; font-weight: 600; color: #64748b; margin-top: 4px; }

.add-member-btn {
  display: flex; align-items: center; gap: 8px; padding: 10px 20px;
  background: #4f46e5; color: white; border-radius: 14px;
  font-size: 14px; font-weight: 800; border: none; cursor: pointer;
  box-shadow: 0 10px 25px -5px rgba(79, 70, 229, 0.4);
  transition: all 0.3s;
}
.add-member-btn:hover { background: #4338ca; transform: translateY(-2px); }

.members-content { flex: 1; overflow-y: auto; }

.members-grid {
  display: grid;
  grid-template-cols: repeat(auto-fill, minmax(240px, 1fr));
  gap: 20px;
}

.member-card {
  background: rgba(255, 255, 255, 0.6);
  backdrop-filter: blur(24px);
  border-radius: 24px;
  padding: 24px;
  border: 1px solid white;
  display: flex;
  flex-direction: column;
  gap: 20px;
  transition: all 0.3s;
}
.member-card:hover { transform: translateY(-4px); background: white; box-shadow: 0 20px 40px rgba(0,0,0,0.04); }

.card-top { display: flex; align-items: flex-start; justify-content: space-between; }

.member-avatar {
  width: 48px; height: 48px; border-radius: 14px;
  background: #f1f5f9; color: #64748b;
  display: flex; align-items: center; justify-content: center;
  font-size: 18px; font-weight: 900;
}
.member-avatar.is-owner { background: rgba(99, 102, 241, 0.1); color: #4f46e5; }

.member-badge {
  padding: 4px 10px; border-radius: 100px; font-size: 10px; font-weight: 900;
  text-transform: uppercase; letter-spacing: 0.05em;
}
.member-badge.owner { background: #fee2e2; color: #ef4444; }
.member-badge.assistant { background: #dcfce7; color: #10b981; }
.member-badge.member { background: #f1f5f9; color: #64748b; }

.member-info h3 { font-size: 16px; font-weight: 800; color: #0f172a; }
.member-id { font-size: 11px; font-weight: 600; color: #94a3b8; font-family: monospace; }

.member-actions { display: flex; gap: 8px; }
.action-btn {
  flex: 1; padding: 8px; border-radius: 10px; border: 1px solid #e2e8f0;
  background: white; color: #64748b; font-size: 12px; font-weight: 700;
  cursor: pointer; transition: all 0.2s;
}
.action-btn:hover { background: #f8fafc; color: #0f172a; border-color: #cbd5e1; }
.action-btn.is-danger:hover { background: #fef2f2; color: #ef4444; border-color: #fecaca; }
</style>
