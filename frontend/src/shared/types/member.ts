export type MemberRole = 'owner' | 'admin' | 'secretary' | 'assistant' | 'member'
export type MemberStatus = 'online' | 'working' | 'dnd' | 'offline'

export interface Member {
  id: string
  workspaceId: string
  name: string
  roleType: MemberRole
  roleKey?: string
  role?: string
  avatar?: string
  terminalType?: string
  terminalCommand?: string
  terminalPath?: string
  autoStartTerminal: boolean
  status: MemberStatus
  manualStatus?: MemberStatus
  createdAt: string
}

export interface MemberCreate {
  name: string
  roleType: MemberRole
  terminalType?: string
  terminalCommand?: string
}