import { useKeyboard, formatShortcutForDisplay } from './useKeyboard'
import type { Ref } from 'vue'

export interface AppShortcut {
  id: string
  key: string
  displayKey: string
  description: string
  scope: 'global' | 'input' | 'terminal'
  callback?: () => void
}

export interface AppShortcutsOptions {
  onFocusSearch?: () => void
  onToggleHelp?: () => void
  onCloseModal?: () => void
  onNavigateChat?: () => void
  onNavigateTerminal?: () => void
  onNavigateMembers?: () => void
  onNavigateSettings?: () => void
  onNavigateSkills?: () => void
  helpVisible?: Ref<boolean>
}

const APP_SHORTCUTS: Omit<AppShortcut, 'displayKey'>[] = [
  {
    id: 'focus-search',
    key: 'Ctrl+K',
    description: 'Focus search/command palette',
    scope: 'global'
  },
  {
    id: 'toggle-help',
    key: 'Ctrl+/',
    description: 'Toggle keyboard shortcuts help',
    scope: 'global'
  },
  {
    id: 'close-modal',
    key: 'Escape',
    description: 'Close modals/dropdowns',
    scope: 'global'
  },
  {
    id: 'navigate-chat',
    key: 'Ctrl+1',
    description: 'Navigate to Chat',
    scope: 'global'
  },
  {
    id: 'navigate-terminal',
    key: 'Ctrl+2',
    description: 'Navigate to Terminal',
    scope: 'global'
  },
  {
    id: 'navigate-members',
    key: 'Ctrl+3',
    description: 'Navigate to Members',
    scope: 'global'
  },
  {
    id: 'navigate-settings',
    key: 'Ctrl+4',
    description: 'Navigate to Settings',
    scope: 'global'
  },
  {
    id: 'navigate-skills',
    key: 'Ctrl+5',
    description: 'Navigate to Skills',
    scope: 'global'
  }
]

/**
 * Composable for registering common application shortcuts
 *
 * @example
 * ```ts
 * const { shortcuts, cleanup } = useAppShortcuts({
 *   onFocusSearch: () => searchInputRef.value?.focus(),
 *   onCloseModal: () => closeModal(),
 *   helpVisible: ref(false)
 * })
 * ```
 */
export function useAppShortcuts(options: AppShortcutsOptions = {}) {
  const { registerShortcut, unregisterAll, setScope } = useKeyboard()

  const shortcuts: AppShortcut[] = APP_SHORTCUTS.map((s) => ({
    ...s,
    displayKey: formatShortcutForDisplay(s.key)
  }))

  const registeredIds: string[] = []

  // Register shortcuts with provided callbacks
  if (options.onFocusSearch) {
    const id = registerShortcut('Ctrl+K', () => options.onFocusSearch?.(), {
      scope: 'global',
      description: 'Focus search/command palette'
    })
    registeredIds.push(id)
  }

  if (options.onToggleHelp && options.helpVisible) {
    const id = registerShortcut('Ctrl+/', () => {
      options.helpVisible!.value = !options.helpVisible!.value
    }, {
      scope: 'global',
      description: 'Toggle keyboard shortcuts help'
    })
    registeredIds.push(id)
  }

  if (options.onCloseModal) {
    const id = registerShortcut('Escape', () => {
      // Don't close if we're in terminal mode (Escape might be used by terminal)
      options.onCloseModal?.()
    }, {
      scope: 'global',
      preventDefault: false,
      description: 'Close modals/dropdowns'
    })
    registeredIds.push(id)
  }

  if (options.onNavigateChat) {
    const id = registerShortcut('Ctrl+1', () => options.onNavigateChat?.(), {
      scope: 'global',
      description: 'Navigate to Chat'
    })
    registeredIds.push(id)
  }

  if (options.onNavigateTerminal) {
    const id = registerShortcut('Ctrl+2', () => options.onNavigateTerminal?.(), {
      scope: 'global',
      description: 'Navigate to Terminal'
    })
    registeredIds.push(id)
  }

  if (options.onNavigateMembers) {
    const id = registerShortcut('Ctrl+3', () => options.onNavigateMembers?.(), {
      scope: 'global',
      description: 'Navigate to Members'
    })
    registeredIds.push(id)
  }

  if (options.onNavigateSettings) {
    const id = registerShortcut('Ctrl+4', () => options.onNavigateSettings?.(), {
      scope: 'global',
      description: 'Navigate to Settings'
    })
    registeredIds.push(id)
  }

  if (options.onNavigateSkills) {
    const id = registerShortcut('Ctrl+5', () => options.onNavigateSkills?.(), {
      scope: 'global',
      description: 'Navigate to Skills'
    })
    registeredIds.push(id)
  }

  const cleanup = () => {
    unregisterAll()
  }

  const enterTerminalScope = () => setScope('terminal')
  const enterInputScope = () => setScope('input')
  const enterGlobalScope = () => setScope('global')

  return {
    shortcuts,
    registeredIds,
    cleanup,
    enterTerminalScope,
    enterInputScope,
    enterGlobalScope,
    formatShortcutForDisplay
  }
}

/**
 * Get all available app shortcuts for display (without callbacks)
 */
export function getAppShortcuts(): AppShortcut[] {
  return APP_SHORTCUTS.map((s) => ({
    ...s,
    displayKey: formatShortcutForDisplay(s.key)
  }))
}