import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { useSettingsStore } from '@/features/settings/settingsStore'
import { workspaceApi } from '@/shared/api/workspace'
import { getApiErrorMessage } from '@/shared/api/errors'
import type { Workspace } from '@/shared/types/workspace'

export const useWorkspaceStore = defineStore('workspace', () => {
  const workspaces = ref<Workspace[]>([])
  const currentWorkspace = ref<Workspace | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  const recentWorkspaces = computed(() =>
    [...workspaces.value].sort((a, b) => {
      const timeA = a.lastOpenedAt ? new Date(a.lastOpenedAt).getTime() : 0
      const timeB = b.lastOpenedAt ? new Date(b.lastOpenedAt).getTime() : 0
      return timeB - timeA
    })
  )

  async function loadWorkspaces() {
    loading.value = true
    error.value = null
    try {
      const response = await workspaceApi.list()
      workspaces.value = response.data
    } catch (e) {
      error.value = getApiErrorMessage(e)
    } finally {
      loading.value = false
    }
  }

  async function openWorkspace(id: string) {
    try {
      const response = await workspaceApi.get(id)
      currentWorkspace.value = response.data
    } catch (e) {
      error.value = getApiErrorMessage(e)
    }
  }

  async function createWorkspace(name: string, path: string) {
    loading.value = true
    error.value = null
    try {
      const settingsStore = useSettingsStore()
      const ownerDisplayName = settingsStore.settings.account.displayName?.trim() || undefined
      const response = await workspaceApi.create({ name, path, ownerDisplayName })
      workspaces.value.push(response.data)
      return response.data
    } catch (e) {
      error.value = getApiErrorMessage(e)
      return null
    } finally {
      loading.value = false
    }
  }

  async function deleteWorkspace(id: string) {
    try {
      await workspaceApi.delete(id)
      workspaces.value = workspaces.value.filter((w) => w.id !== id)
      if (currentWorkspace.value?.id === id) {
        currentWorkspace.value = null
      }
    } catch (e) {
      error.value = getApiErrorMessage(e)
    }
  }

  async function browseWorkspace(workspaceId: string, path?: string) {
    try {
      const response = await workspaceApi.browse(workspaceId, path)
      return response.data
    } catch (e) {
      error.value = getApiErrorMessage(e)
      return null
    }
  }

  function closeWorkspace() {
    currentWorkspace.value = null
  }

  return {
    workspaces,
    currentWorkspace,
    loading,
    error,
    recentWorkspaces,
    loadWorkspaces,
    openWorkspace,
    createWorkspace,
    deleteWorkspace,
    browseWorkspace,
    closeWorkspace,
  }
})