// Terminal Member Session Management: handles terminal creation for AI assistants
import axios from 'axios'
import { ref, watch } from 'vue'
import { acceptHMRUpdate, defineStore } from 'pinia'
import { notifyUserError } from '@/shared/notifyError'
import { i18n } from '@/i18n'
import { useToastStore } from '@/stores/toastStore'
import { useTerminalStore } from '@/features/terminal/terminalStore'
import { useProjectStore } from '@/features/workspace/projectStore'
import { useWorkspaceStore } from '@/features/workspace/workspaceStore'
import { terminalApi } from '@/shared/api/terminal'
import type { Member } from '@/shared/types/member'

/**
 * Checks if member has terminal configuration
 */
export function hasTerminalConfig(terminalType?: string, terminalCommand?: string): boolean {
  if (!terminalType) return false
  if (terminalType === 'shell' && !terminalCommand?.trim()) return false
  return true
}

/**
 * Splits command string into command and args for API request
 * e.g. "claude --dangerously-skip-permissions" -> { command: "claude", args: ["--dangerously-skip-permissions"] }
 */
function splitCommand(cmd: string): { command: string; args?: string[] } {
  const trimmed = cmd.trim()
  if (!trimmed) return { command: '' }

  // Split by whitespace, but handle edge cases
  const parts = trimmed.split(/\s+/)
  if (parts.length === 1) {
    return { command: parts[0] }
  }
  return { command: parts[0], args: parts.slice(1) }
}

type MemberSession = {
  memberId: string
  terminalId: string
  title: string
  workspaceId: string
  status: 'pending' | 'connecting' | 'connected' | 'disconnected'
}

/**
 * Terminal Member Store
 * Manages terminal sessions for AI assistants
 */
export const useTerminalMemberStore = defineStore('terminal-member', () => {
  const terminalStore = useTerminalStore()
  const projectStore = useProjectStore()
  const workspaceStore = useWorkspaceStore()

  const memberSessions = ref<Record<string, MemberSession>>({})
  /** 服务端池中仍有 PTY 的成员（REQ-303 轮询，与本地 WebSocket 状态互补） */
  const serverPTYByMember = ref<Record<string, { sessionId: string; pid: number }>>({})
  let pollTimer: ReturnType<typeof setInterval> | null = null

  const stopServerPoll = () => {
    if (pollTimer) {
      clearInterval(pollTimer)
      pollTimer = null
    }
  }

  const refreshServerTerminalStatus = async () => {
    const wid = workspaceStore.currentWorkspace?.id
    if (!wid) {
      serverPTYByMember.value = {}
      return
    }
    try {
      const sessions = await terminalApi.listWorkspaceTerminalSessions(wid)
      const next: Record<string, { sessionId: string; pid: number }> = {}
      for (const s of sessions) {
        next[s.memberId] = { sessionId: s.sessionId, pid: s.pid }
      }
      serverPTYByMember.value = next
    } catch {
      /* GET terminal-sessions: axios interceptor reports API errors */
    }
  }

  const getServerTerminalForMember = (memberId: string) => serverPTYByMember.value[memberId] ?? null

  watch(
    () => workspaceStore.currentWorkspace?.id,
    (wid) => {
      stopServerPoll()
      serverPTYByMember.value = {}
      if (!wid) return
      void refreshServerTerminalStatus()
      pollTimer = setInterval(() => void refreshServerTerminalStatus(), 8000)
    },
    { immediate: true }
  )

  /**
   * Build member key for session lookup
   */
  const buildMemberKey = (memberId: string, workspaceId?: string) =>
    workspaceId ? `${workspaceId}:${memberId}` : memberId

  /**
   * Get session for a member
   */
  const getSession = (memberId: string, workspaceId?: string) => {
    const key = buildMemberKey(memberId, workspaceId)
    return memberSessions.value[key] || null
  }

  /** Merged flags for concurrent startMemberSession calls (same workspace+member). */
  type StartMerge = { wantOpenTab: boolean; quietAutostart: boolean }
  const startInflight = new Map<string, Promise<MemberSession | null>>()
  const startMerge = new Map<string, StartMerge>()

  /**
   * Start terminal session for a member
   */
  const startMemberSession = async (
    member: Member,
    options?: { openTab?: boolean; quietAutostart?: boolean }
  ): Promise<MemberSession | null> => {
    const workspace = workspaceStore.currentWorkspace
    if (!workspace) return null

    if (!hasTerminalConfig(member.terminalType, member.terminalCommand)) {
      return null
    }

    const key = buildMemberKey(member.id, workspace.id)
    const wantOpenTab = options?.openTab ?? true
    const quietAutostart = options?.quietAutostart ?? false

    const existingP = startInflight.get(key)
    if (existingP) {
      const m = startMerge.get(key)!
      m.wantOpenTab = m.wantOpenTab || wantOpenTab
      m.quietAutostart = m.quietAutostart && quietAutostart
      return existingP
    }

    const merge: StartMerge = { wantOpenTab, quietAutostart }
    startMerge.set(key, merge)

    const promise = (async (): Promise<MemberSession | null> => {
      try {
        const attached = await terminalApi.getSessionForMember(workspace.id, member.id)
        const skipToast = merge.quietAutostart
        const { command, args } = member.terminalCommand ? splitCommand(member.terminalCommand) : { command: undefined, args: undefined }
        const response =
          attached ??
          (await terminalApi.createSession(
            {
              workspaceId: workspace.id,
              terminalType: member.terminalType || 'native',
              memberId: member.id,
              command,
              args,
              memberName: member.name // For introduction prompt to AI assistant
            },
            { skipErrorToast: skipToast }
          ))

        const terminalId = response.sessionId
        const title = member.name

        const session: MemberSession = {
          memberId: member.id,
          terminalId,
          title,
          workspaceId: workspace.id,
          status: 'connecting'
        }

        memberSessions.value[key] = session

        // Always open tab for terminal management (controls keepAlive)
        terminalStore.openTab(terminalId, {
          title,
          memberId: member.id,
          terminalType: member.terminalType as 'native' | 'web' | 'claude' | 'custom',
          keepAlive: true,
          activate: merge.wantOpenTab
        })

        // Always establish WebSocket connection so @mentions can reach the assistant
        try {
          await terminalStore.createConnection(terminalId)
        } catch {
          /* createConnection already reports errors */
        }

        projectStore.updateMember(member.id, { status: 'online' })

        return session
      } catch (error) {
        if (merge.quietAutostart) {
          notifyUserError('Assistant terminal', error, { skipToast: true })
          try {
            useToastStore().pushToast(i18n.global.t('members.terminalAutostartFailed'), {
              tone: 'warning',
              duration: 7000
            })
          } catch {
            /* Pinia not ready */
          }
        } else if (!axios.isAxiosError(error)) {
          notifyUserError('Assistant terminal', error)
        }
        projectStore.updateMember(member.id, { status: 'offline' })
        return null
      } finally {
        startInflight.delete(key)
        startMerge.delete(key)
      }
    })()

    startInflight.set(key, promise)
    return promise
  }

  /**
   * Ensure member session exists, create if not
   */
  const ensureMemberSession = async (
    member: Member,
    options?: { openTab?: boolean }
  ): Promise<MemberSession | null> => {
    const workspaceId = workspaceStore.currentWorkspace?.id
    const existing = getSession(member.id, workspaceId)

    if (existing && existing.status !== 'disconnected') {
      const wantOpenTab = options?.openTab ?? true
      if (wantOpenTab) {
        terminalStore.openTab(existing.terminalId, {
          title: member.name,
          memberId: member.id,
          terminalType: member.terminalType as 'native' | 'web' | 'claude' | 'custom',
          keepAlive: true,
          activate: true
        })
        // autoStart 用 openTab:false 只建了后端 PTY，未拉 WS；用户点开终端时必须补连
        if (!terminalStore.getConnection(existing.terminalId)) {
          try {
            await terminalStore.createConnection(existing.terminalId)
          } catch {
            /* createConnection / axios 已提示错误 */
          }
        }
      }
      return existing
    }

    return startMemberSession(member, options)
  }

  /**
   * Open member terminal (ensure session exists and open tab)
   */
  const openMemberTerminal = async (member: Member): Promise<MemberSession | null> => {
    if (!hasTerminalConfig(member.terminalType, member.terminalCommand)) {
      return null
    }

    const session = await ensureMemberSession(member, { openTab: true })

    // Mark autoStartTerminal as true if not already
    if (session && member.autoStartTerminal === false) {
      projectStore.updateMember(member.id, { autoStartTerminal: true })
    }

    return session
  }

  /**
   * Stop member terminal session
   */
  const stopMemberSession = async (memberId: string) => {
    const workspaceId = workspaceStore.currentWorkspace?.id
    const session = getSession(memberId, workspaceId)

    if (!session) return

    terminalStore.closeConnection(session.terminalId)
    delete memberSessions.value[buildMemberKey(memberId, workspaceId)]
    projectStore.updateMember(memberId, { status: 'offline' })
  }

  /**
   * Auto-start terminals for members with autoStartTerminal flag
   */
  const autoStartMemberTerminals = async () => {
    const members = projectStore.members
    for (const member of members) {
      if (
        (member.roleType === 'assistant' ||
          member.roleType === 'secretary' ||
          member.roleType === 'member') &&
        member.autoStartTerminal &&
        hasTerminalConfig(member.terminalType, member.terminalCommand)
      ) {
        const existing = getSession(member.id, workspaceStore.currentWorkspace?.id)
        if (!existing || existing.status === 'disconnected') {
          await startMemberSession(member, { openTab: false, quietAutostart: true })
        }
      }
    }
  }

  /**
   * Reset all sessions
   */
  const reset = () => {
    stopServerPoll()
    memberSessions.value = {}
    serverPTYByMember.value = {}
  }

  return {
    memberSessions,
    serverPTYByMember,
    getSession,
    getServerTerminalForMember,
    refreshServerTerminalStatus,
    startMemberSession,
    ensureMemberSession,
    openMemberTerminal,
    stopMemberSession,
    autoStartMemberTerminals,
    reset
  }
})

if (import.meta.hot) {
  import.meta.hot.accept(acceptHMRUpdate(useTerminalMemberStore, import.meta.hot))
}