import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/features/auth/authStore'

const routes = [
  {
    path: '/login',
    name: 'login',
    component: () => import('@/features/auth/LoginPage.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/',
    name: 'home',
    redirect: '/workspaces',
    meta: { requiresAuth: true }
  },
  {
    path: '/workspaces',
    name: 'workspaces',
    component: () => import('@/features/workspace/WorkspaceSelection.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/workspace/:id',
    name: 'workspace',
    component: () => import('@/features/workspace/WorkspaceMain.vue'),
    meta: { requiresAuth: true },
    children: [
      {
        path: '',
        redirect: { name: 'chat' },
      },
      {
        path: 'chat',
        name: 'chat',
        component: () => import('@/features/chat/ChatInterface.vue'),
      },
      {
        path: 'terminal',
        name: 'terminal',
        component: () => import('@/features/terminal/TerminalWorkspace.vue'),
      },
      {
        path: 'members',
        name: 'members',
        component: () => import('@/features/members/MembersPage.vue'),
      },
      {
        path: 'settings',
        name: 'settings',
        component: () => import('@/features/settings/Settings.vue'),
      },
      {
        path: 'skills',
        name: 'skills',
        component: () => import('@/features/skills/SkillsPlaceholder.vue'),
      },
    ],
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

// Auth guard
router.beforeEach(async (to, _from, next) => {
  const authStore = useAuthStore()

  // Ensure config is fetched before the first routing decision
  if (!authStore.initialized) {
    await authStore.fetchConfig()
  }

  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    next({ name: 'login' })
  } else if (to.name === 'login' && authStore.isAuthenticated) {
    next({ name: 'workspaces' })
  } else {
    next()
  }
})

export default router