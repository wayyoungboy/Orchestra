import { defineStore } from 'pinia'
import { ref } from 'vue'
import { notifyUserError } from '@/shared/notifyError'
import client from '@/shared/api/client'
import axios from 'axios'

export const useAuthStore = defineStore('auth', () => {
  const isAuthEnabled = ref(true)
  const initialized = ref(false)
  const isAuthenticated = ref(!!localStorage.getItem('orchestra.auth.token'))
  const currentUser = ref<string | null>(localStorage.getItem('orchestra.user'))

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
    isAuthenticated,
    currentUser,
    fetchConfig,
    login,
    logout
  }
})
