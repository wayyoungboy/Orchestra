import { defineStore } from 'pinia'
import { ref } from 'vue'
import client from '@/shared/api/client'
import { notifyUserError } from '@/shared/notifyError'
import type { Workspace } from '@/shared/types/workspace'

export const useWorkspaceStore = defineStore('workspace', () => {
  const workspaces = ref<Workspace[]>([])
  const recentWorkspaces = ref<Workspace[]>([])
  const currentWorkspace = ref<Workspace | null>(null)
  const loading = ref(false)
  const searchResults = ref<any[]>([])

  async function loadWorkspaces() {
    loading.value = true
    try {
      const response = await client.get('/workspaces')
      workspaces.value = response.data || []
      recentWorkspaces.value = workspaces.value.slice(0, 6)
    } catch (e) {
      notifyUserError('Failed to load workspaces', e)
    } finally {
      loading.value = false
    }
  }

  async function openWorkspace(id: string) {
    try {
      const response = await client.get(`/workspaces/${id}`)
      currentWorkspace.value = response.data
      return response.data
    } catch (e) {
      notifyUserError('Failed to open workspace', e)
    }
  }

  async function createWorkspace(name: string, path: string) {
    try {
      const response = await client.post('/workspaces', { name, path })
      await loadWorkspaces()
      return response.data
    } catch (e) {
      notifyUserError('Failed to create workspace', e)
    }
  }

  /**
   * New: Update workspace properties (PUT /api/workspaces/{id})
   */
  async function updateWorkspace(id: string, data: { name?: string, path?: string }) {
    try {
      const response = await client.put(`/workspaces/${id}`, data)
      if (currentWorkspace.value?.id === id) {
        currentWorkspace.value = response.data
      }
      await loadWorkspaces()
      return response.data
    } catch (e) {
      notifyUserError('Failed to update workspace', e)
    }
  }

  /**
   * New: Search messages in workspace (GET /api/workspaces/{id}/search)
   */
  async function searchWorkspace(query: string) {
    if (!currentWorkspace.value || !query.trim()) {
      searchResults.value = []
      return
    }
    
    try {
      const response = await client.get(`/workspaces/${currentWorkspace.value.id}/search`, {
        params: { q: query, limit: 50 }
      })
      searchResults.value = response.data || []
      return searchResults.value
    } catch (e) {
      notifyUserError('Search failed', e)
    }
  }

  async function deleteWorkspace(id: string) {
    try {
      await client.delete(`/workspaces/${id}`)
      if (currentWorkspace.value?.id === id) currentWorkspace.value = null
      await loadWorkspaces()
    } catch (e) {
      notifyUserError('Failed to delete workspace', e)
    }
  }

  return {
    workspaces,
    recentWorkspaces,
    currentWorkspace,
    searchResults,
    loading,
    loadWorkspaces,
    openWorkspace,
    createWorkspace,
    updateWorkspace,
    searchWorkspace,
    deleteWorkspace
  }
})
