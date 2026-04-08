import { defineStore } from 'pinia'
import { ref } from 'vue'
import client from '@/shared/api/client'
import type { Member } from '@/shared/types/member'

export const useMemberStore = defineStore('member', () => {
  const members = ref<Member[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function fetchMembers(workspaceId: string) {
    loading.value = true
    error.value = null

    try {
      const response = await client.get(`/workspaces/${workspaceId}/members`)
      members.value = response.data
    } catch (e: any) {
      error.value = e.response?.data?.error || 'Failed to fetch members'
    } finally {
      loading.value = false
    }
  }

  async function createMember(workspaceId: string, data: Partial<Member>) {
    const response = await client.post(`/workspaces/${workspaceId}/members`, data)
    const member = response.data
    members.value.push(member)
    return member
  }

  async function updateMember(workspaceId: string, memberId: string, data: Partial<Member>) {
    const response = await client.put(`/workspaces/${workspaceId}/members/${memberId}`, data)
    const updatedMember = response.data
    const index = members.value.findIndex(m => m.id === memberId)
    if (index !== -1) {
      members.value[index] = updatedMember
    }
    return updatedMember
  }

  async function deleteMember(workspaceId: string, memberId: string) {
    await client.delete(`/workspaces/${workspaceId}/members/${memberId}`)
    members.value = members.value.filter(m => m.id !== memberId)
  }

  function getMemberById(id: string) {
    return members.value.find(m => m.id === id)
  }

  return {
    members,
    loading,
    error,
    fetchMembers,
    createMember,
    updateMember,
    deleteMember,
    getMemberById
  }
})