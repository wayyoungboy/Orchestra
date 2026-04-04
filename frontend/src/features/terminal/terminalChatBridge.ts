// Subscribers receive reference-desktop-shaped terminal chat stream payloads from any assistant terminal WS.
import type { TerminalChatStreamPayload } from '@/shared/types/terminal'

type Listener = (sessionId: string, payload: TerminalChatStreamPayload) => void

const listeners = new Set<Listener>()

export function onTerminalChatStream(listener: Listener): () => void {
  listeners.add(listener)
  return () => listeners.delete(listener)
}

export function notifyTerminalChatStream(sessionId: string, payload: TerminalChatStreamPayload): void {
  listeners.forEach((fn) => {
    fn(sessionId, payload)
  })
}
