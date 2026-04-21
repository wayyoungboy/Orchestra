import client from '@/shared/api/client'
import type { Workspace, WorkspaceCreate, WorkspaceUpdate } from '@/shared/types/workspace'

export async function listWorkspaces(): Promise<Workspace[]> {
  const { data } = await client.get('/workspaces')
  return data
}

export async function getWorkspace(id: string): Promise<Workspace> {
  const { data } = await client.get(`/workspaces/${id}`)
  return data
}

export async function createWorkspace(ws: WorkspaceCreate): Promise<Workspace> {
  const { data } = await client.post('/workspaces', ws)
  return data
}

export async function updateWorkspace(id: string, ws: WorkspaceUpdate): Promise<Workspace> {
  const { data } = await client.put(`/workspaces/${id}`, ws)
  return data
}

export async function deleteWorkspace(id: string): Promise<void> {
  await client.delete(`/workspaces/${id}`)
}

export async function validatePath(path: string): Promise<{ valid: boolean; error?: string }> {
  const { data } = await client.post('/workspaces/validate-path', { path })
  return data
}

export async function browseWorkspace(id: string, path?: string): Promise<{ path: string; entries: { name: string; isDir: boolean; size: number }[] }> {
  const params = path ? { path } : {}
  const { data } = await client.get(`/workspaces/${id}/browse`, { params })
  return data
}
