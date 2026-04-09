import { defineStore } from 'pinia'
import { ref, computed, nextTick } from 'vue'
import { memberApi } from '@/shared/api/member'
import { getApiErrorMessage } from '@/shared/api/errors'
import { notifyUserError } from '@/shared/notifyError'
import { useToastStore } from '@/stores/toastStore'
import { i18n } from '@/i18n'
import type { Member, MemberCreate } from '@/shared/types/member'
import { useWorkspaceStore } from './workspaceStore'

function workspaceIdOrCurrent(
  explicit: string | undefined,
  workspaceStore: ReturnType<typeof useWorkspaceStore>
): string | undefined {
  if (explicit && explicit.trim() !== '') return explicit
  return workspaceStore.currentWorkspace?.id
}

export const useProjectStore = defineStore('project', () => {
  const members = ref<Member[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  const workspaceStore = useWorkspaceStore()

  /** Bumps when local member mutations must win over in-flight GET /members (avoids stale list wiping a just-added assistant). */
  let membersListGeneration = 0

  /** Member ids recently added locally; kept until a list response includes them (avoids empty/stale GET wiping the row). */
  const optimisticMemberIds = ref<Set<string>>(new Set())

  const sortedMembers = computed(() =>
    [...members.value].sort((a, b) => {
      const roleOrder: Record<string, number> = { owner: 0, admin: 1, secretary: 2, assistant: 3, member: 4 }
      const ra = roleOrder[a.roleType] ?? 99
      const rb = roleOrder[b.roleType] ?? 99
      return ra - rb
    })
  )

  async function loadMembers(explicitWorkspaceId?: string, options?: { silent?: boolean }) {
    const workspaceId = workspaceIdOrCurrent(explicitWorkspaceId, workspaceStore)
    if (!workspaceId) return

    const silent = options?.silent === true
    const gen = ++membersListGeneration
    if (!silent) {
      loading.value = true
      error.value = null
    }
    try {
      const response = await memberApi.list(workspaceId)
      if (gen !== membersListGeneration) return
      const raw = response.data
      let nextList: Member[] = Array.isArray(raw) ? raw : []
      let opt = optimisticMemberIds.value
      if (opt.size > 0) {
        const seen = new Set(nextList.map((m) => m.id))
        for (const id of opt) {
          if (seen.has(id)) continue
          // Match by id only: workspaceId on the created row can be missing in edge cases; strict match dropped merges.
          const hold = members.value.find((m) => m.id === id)
          if (hold) {
            nextList = [...nextList, hold]
          }
        }
        opt = new Set(opt)
        for (const m of nextList) {
          opt.delete(m.id)
        }
        optimisticMemberIds.value = opt
      }
      members.value = nextList
    } catch (e) {
      if (gen !== membersListGeneration) return
      if (!silent) {
        error.value = getApiErrorMessage(e)
      }
    } finally {
      if (!silent) {
        loading.value = false
      }
    }
  }

  async function addMember(data: MemberCreate, explicitWorkspaceId?: string) {
    const workspaceId = workspaceIdOrCurrent(explicitWorkspaceId, workspaceStore)
    if (!workspaceId) {
      const msg = 'No workspace selected (URL or switcher out of sync).'
      error.value = msg
      notifyUserError('Add member', new Error(msg))
      return null
    }

    try {
      const response = await memberApi.create(workspaceId, data)
      const created = response.data
      if (!created?.id) {
        error.value = 'Invalid member response from server'
        return null
      }
      {
        const next = new Set(optimisticMemberIds.value)
        next.add(created.id)
        optimisticMemberIds.value = next
      }
      members.value.push(created)
      await nextTick()
      try {
        useToastStore().success(i18n.global.t('members.memberAdded'), 3500)
      } catch {
        /* Pinia not ready */
      }
      const wid = workspaceId
      void loadMembers(wid, { silent: true })
      return created
    } catch (e) {
      const msg = getApiErrorMessage(e)
      error.value = msg
      return null
    }
  }

  async function updateMember(memberId: string, data: Partial<Member>, explicitWorkspaceId?: string) {
    const workspaceId = workspaceIdOrCurrent(explicitWorkspaceId, workspaceStore)
    if (!workspaceId) return

    try {
      const response = await memberApi.update(workspaceId, memberId, data)
      const index = members.value.findIndex((m) => m.id === memberId)
      if (index !== -1) {
        members.value[index] = response.data
      }
    } catch (e) {
      error.value = getApiErrorMessage(e)
    }
  }

  async function removeMember(memberId: string, explicitWorkspaceId?: string) {
    const workspaceId = workspaceIdOrCurrent(explicitWorkspaceId, workspaceStore)
    if (!workspaceId) return

    try {
      await memberApi.delete(workspaceId, memberId)
      membersListGeneration++
      {
        const next = new Set(optimisticMemberIds.value)
        next.delete(memberId)
        optimisticMemberIds.value = next
      }
      members.value = members.value.filter((m) => m.id !== memberId)
    } catch (e) {
      error.value = getApiErrorMessage(e)
    }
  }

  function reset() {
    membersListGeneration++
    optimisticMemberIds.value = new Set()
    members.value = []
    error.value = null
  }

  return {
    members,
    loading,
    error,
    sortedMembers,
    loadMembers,
    addMember,
    updateMember,
    removeMember,
    reset,
  }
})