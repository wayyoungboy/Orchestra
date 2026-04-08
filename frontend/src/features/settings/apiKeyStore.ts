import { defineStore } from 'pinia'
import { ref } from 'vue'
import client from '@/shared/api/client'
import { notifyUserError } from '@/shared/notifyError'

export type APIKeyProvider = 'anthropic' | 'openai' | 'google' | 'custom'

export interface APIKeyResponse {
  id: string
  provider: APIKeyProvider
  keyPreview: string
  isValid: boolean
  createdAt: string
  updatedAt: string
}

export interface APIKeyTestResult {
  valid: boolean
  message: string
  provider: string
}

export const useAPIKeyStore = defineStore('apiKeys', () => {
  const keys = ref<APIKeyResponse[]>([])
  const loading = ref(false)
  const testing = ref(false)

  /**
   * List all stored API keys
   */
  async function listKeys(): Promise<void> {
    loading.value = true
    try {
      const response = await client.get<APIKeyResponse[]>('/api-keys')
      keys.value = response.data
    } catch (e) {
      notifyUserError('Failed to list API keys', e)
    } finally {
      loading.value = false
    }
  }

  /**
   * Save an API key for a provider
   */
  async function saveKey(provider: APIKeyProvider, key: string): Promise<boolean> {
    try {
      await client.post('/api-keys', { provider, key })
      await listKeys()
      return true
    } catch (e) {
      notifyUserError('Failed to save API key', e)
      return false
    }
  }

  /**
   * Delete an API key
   */
  async function deleteKey(id: string): Promise<boolean> {
    try {
      await client.delete(`/api-keys/${id}`)
      keys.value = keys.value.filter(k => k.id !== id)
      return true
    } catch (e) {
      notifyUserError('Failed to delete API key', e)
      return false
    }
  }

  /**
   * Test an API key (either a new one or stored one)
   */
  async function testKey(provider: APIKeyProvider, key?: string): Promise<APIKeyTestResult | null> {
    testing.value = true
    try {
      const response = await client.post<APIKeyTestResult>('/api-keys/test', {
        provider,
        key: key || undefined
      })
      return response.data
    } catch (e) {
      notifyUserError('Failed to test API key', e)
      return null
    } finally {
      testing.value = false
    }
  }

  return {
    keys,
    loading,
    testing,
    listKeys,
    saveKey,
    deleteKey,
    testKey
  }
})