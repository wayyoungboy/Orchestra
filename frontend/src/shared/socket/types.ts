export interface TerminalInput {
  type: 'input'
  data: string
}

export interface TerminalResize {
  type: 'resize'
  cols: number
  rows: number
}

export interface TerminalClose {
  type: 'close'
}

export type TerminalClientMessage = TerminalInput | TerminalResize | TerminalClose

export interface TerminalOutput {
  type: 'output'
  data: string
}

export interface TerminalError {
  type: 'error'
  message: string
}

export interface TerminalExit {
  type: 'exit'
  code: number
}

export interface TerminalConnected {
  type: 'connected'
  sessionId: string
}