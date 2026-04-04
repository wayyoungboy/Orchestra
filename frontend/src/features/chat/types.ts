// Re-export types from shared types for backward compatibility
export * from '@/shared/types/chat'

// Additional local types if needed
export interface LocalMessage {
  id: string
  conversationId: string
  senderId: string
  senderName: string
  senderAvatar?: string
  content: string
  createdAt: string
}