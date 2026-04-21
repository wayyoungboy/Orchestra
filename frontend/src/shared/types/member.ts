export type MemberRole = 'owner' | 'secretary' | 'assistant'

// Manual presence (user-set, like Discord status)
export type MemberStatus = 'online' | 'working' | 'dnd' | 'offline'

// Connection state (system-detected)
export type MemberConnection = 'connecting' | 'connected' | 'working' | 'disconnected'

export interface Member {
  id: string
  workspaceId: string
  name: string
  roleType: MemberRole
  roleKey?: string
  role?: string
  avatar?: string
  status: MemberStatus
  manualStatus?: MemberStatus
  connection?: MemberConnection
  createdAt: string
  acpEnabled?: boolean
  acpCommand?: string
  acpArgs?: string[]
}

// Compute display status from manual + connection states.
// Manual "offline" overrides connection.
// Connection "disconnected" shows as "offline" when manual is not explicitly "offline".
export function displayMemberStatus(m: Member): MemberStatus {
  if (m.manualStatus && m.manualStatus !== 'online') {
    return m.manualStatus
  }
  if (m.connection === 'disconnected') {
    return 'offline'
  }
  if (m.connection === 'working') {
    return 'working'
  }
  if (m.connection === 'connecting') {
    return 'online'
  }
  return m.status || 'online'
}

export interface MemberCreate {
  name: string
  roleType: MemberRole
}
