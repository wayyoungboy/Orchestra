import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { notifyUserError } from '@/shared/notifyError'

export const useAuthStore = defineStore('auth', () => {
  const storedCredentials = localStorage.getItem('orchestra.credentials')
  const defaultUsername = 'orchestra'
  const defaultPassword = 'orchestra'

  let savedUsername = defaultUsername
  let savedPassword = defaultPassword

  if (storedCredentials) {
    try {
      const parsed = JSON.parse(storedCredentials)
      savedUsername = parsed.username || defaultUsername
      savedPassword = parsed.password || defaultPassword
    } catch (e) {
      notifyUserError('Saved login credentials (orchestra.credentials)', e)
    }
  }

  const isAuthenticated = ref(localStorage.getItem('orchestra.auth') === 'true')
  const currentUser = ref<string | null>(localStorage.getItem('orchestra.user'))

  const isLoggedIn = computed(() => isAuthenticated.value)

  function login(username: string, password: string): boolean {
    if (username === savedUsername && password === savedPassword) {
      isAuthenticated.value = true
      currentUser.value = username
      localStorage.setItem('orchestra.auth', 'true')
      localStorage.setItem('orchestra.user', username)
      return true
    }
    return false
  }

  function logout() {
    isAuthenticated.value = false
    currentUser.value = null
    localStorage.removeItem('orchestra.auth')
    localStorage.removeItem('orchestra.user')
  }

  function updateCredentials(newUsername: string, newPassword: string) {
    savedUsername = newUsername
    savedPassword = newPassword
    localStorage.setItem('orchestra.credentials', JSON.stringify({
      username: newUsername,
      password: newPassword
    }))
  }

  function getCredentials() {
    return { username: savedUsername, password: savedPassword }
  }

  return {
    isAuthenticated,
    currentUser,
    isLoggedIn,
    login,
    logout,
    updateCredentials,
    getCredentials
  }
})
