// Bridge layer: domain-specific API abstractions.
// Each bridge wraps the HTTP client with typed, purpose-built methods.
// Stores should import bridges, not the raw API client.

export type { Workspace, WorkspaceCreate, WorkspaceUpdate } from '@/shared/types/workspace'
export type { Member, MemberCreate, MemberStatus, MemberRole, MemberConnection } from '@/shared/types/member'
export * from './workspaceBridge'
export * from './memberBridge'
