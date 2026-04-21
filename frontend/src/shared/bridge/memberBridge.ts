import client from '@/shared/api/client'
import type { Member, MemberCreate, MemberStatus } from '@/shared/types/member'

export async function listMembers(workspaceId: string): Promise<Member[]> {
  const { data } = await client.get(`/workspaces/${workspaceId}/members`)
  return data
}

export async function getMember(workspaceId: string, memberId: string): Promise<Member> {
  const { data } = await client.get(`/workspaces/${workspaceId}/members/${memberId}`)
  return data
}

export async function createMember(workspaceId: string, input: MemberCreate): Promise<Member> {
  const { data } = await client.post(`/workspaces/${workspaceId}/members`, input)
  return data
}

export async function updateMember(workspaceId: string, memberId: string, updates: Partial<Member>): Promise<Member> {
  const { data } = await client.put(`/workspaces/${workspaceId}/members/${memberId}`, updates)
  return data
}

export async function deleteMember(workspaceId: string, memberId: string): Promise<void> {
  await client.delete(`/workspaces/${workspaceId}/members/${memberId}`)
}

export async function updatePresence(workspaceId: string, memberId: string, status: MemberStatus): Promise<void> {
  await client.post(`/workspaces/${workspaceId}/members/${memberId}/presence`, { status })
}
