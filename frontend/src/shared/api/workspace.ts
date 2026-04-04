import client from './client'
import type { Workspace, WorkspaceCreate, BrowseResult } from '@/shared/types/workspace'

export const workspaceApi = {
  list: () => client.get<Workspace[]>('/workspaces'),

  get: (id: string) => client.get<Workspace>(`/workspaces/${id}`),

  create: (data: WorkspaceCreate) => client.post<Workspace>('/workspaces', data),

  delete: (id: string) => client.delete(`/workspaces/${id}`),

  browse: (workspaceId: string, path?: string) =>
    client.get<BrowseResult>(`/workspaces/${workspaceId}/browse`, {
      params: { path },
    }),

  browseRoot: (path?: string) =>
    client.get<BrowseResult>('/browse', { params: { path } }),
}