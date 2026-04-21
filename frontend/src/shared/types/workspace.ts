export interface Workspace {
  id: string
  name: string
  path: string
  lastOpenedAt: string
  createdAt: string
}

export interface WorkspaceCreate {
  name: string
  path: string
  ownerDisplayName?: string
}

export interface WorkspaceUpdate {
  name?: string
  path?: string
}

export interface FileInfo {
  name: string
  path: string
  isDir: boolean
  size: number
  modTime: string
  mode: string
}

export interface BrowseResult {
  basePath: string
  home?: string
  files: FileInfo[]
}