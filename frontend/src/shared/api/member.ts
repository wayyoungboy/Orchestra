import client from './client'
import type { Member, MemberCreate } from '@/shared/types/member'

export const memberApi = {
  list: (workspaceId: string) =>
    client.get<Member[]>(`/workspaces/${workspaceId}/members`),

  create: (workspaceId: string, data: MemberCreate) =>
    client.post<Member>(`/workspaces/${workspaceId}/members`, data),

  update: (workspaceId: string, memberId: string, data: Partial<Member>) =>
    client.put<Member>(`/workspaces/${workspaceId}/members/${memberId}`, data),

  delete: (workspaceId: string, memberId: string) =>
    client.delete(`/workspaces/${workspaceId}/members/${memberId}`),
}