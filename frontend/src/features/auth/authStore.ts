import { defineStore } from 'pinia'
import { ref } from 'vue'
import { notifyUserError } from '@/shared/notifyError'
import client from '@/shared/api/client'
import axios from 'axios'

export interface UserInfo {
  id: string
  username: string
}

export const useAuthStore = defineStore('auth', () => {
  const isAuthEnabled = ref(true)
  const initialized = ref(false)
  const isAuthenticated = ref(!!localStorage.getItem('orchestra.auth.token'))
  const currentUser = ref<string | null>(localStorage.getItem('orchestra.user'))
  const currentUserId = ref<string | null>(null)

  /**
   * Fetch current user info including user ID
   */
  async function fetchCurrentUser(): Promise<UserInfo | null> {
    if (!isAuthenticated.value) return null
    try {
      const response = await client.get('/auth/me')
      const user = response.data?.user
      if (user) {
        currentUserId.value = user.id || user.ID || 'default'
        currentUser.value = user.username || user.Username || 'anonymous'
        return {
          id: currentUserId.value as string,
          username: currentUser.value as string
        }
      }
    } catch (e) {
      console.warn('Failed to fetch current user', e)
    }
    return null
  }

  /**
   * Fetch authentication configuration from backend
   */
  async function fetchConfig() {
    if (initialized.value) return

    try {
      const response = await client.get('/auth/config')
      const { enabled } = response.data
      isAuthEnabled.value = enabled

      // If auth is disabled globally, treat as authenticated
      if (!enabled) {
        isAuthenticated.value = true
        if (!currentUser.value) currentUser.value = 'guest'
        currentUserId.value = 'default'
      } else if (isAuthenticated.value) {
        // Fetch actual user info when auth is enabled
        await fetchCurrentUser()
      }
    } catch (e) {
      console.warn('Failed to fetch auth config, defaulting to enabled', e)
    } finally {
      initialized.value = true
    }
  }

  /**
   * Perform login with username and password
   */
  async function login(username: string, password: string): Promise<boolean> {
    try {
      const response = await client.post('/auth/login', { username, password })
      const { token, user } = response.data

      // Handle special "disabled-auth-mode" token
      if (token === 'disabled-auth-mode') {
        isAuthenticated.value = true
        currentUser.value = username || 'guest'
        localStorage.setItem('orchestra.auth.token', 'disabled-auth-mode')
        localStorage.setItem('orchestra.user', currentUser.value)
        return true
      }

      isAuthenticated.value = true
      currentUser.value = user.username
      
      localStorage.setItem('orchestra.auth.token', token)
      localStorage.setItem('orchestra.user', user.username)
      return true
    } catch (error) {
      if (axios.isAxiosError(error) && error.response?.status === 401) {
        // Return false to let component show error message
      } else {
        notifyUserError('Login failed', error)
      }
      return false
    }
  }

  function logout() {
    isAuthenticated.value = false
    currentUser.value = null
    localStorage.removeItem('orchestra.auth.token')
    localStorage.removeItem('orchestra.user')
  }

  return {
    isAuthEnabled,
    initialized,
    isAuthenticated,
    currentUser,
    currentUserId,
    fetchConfig,
    fetchCurrentUser,
    login,
    logout
  }
})
