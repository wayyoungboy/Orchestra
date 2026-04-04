import { onMounted, onUnmounted, ref, type Ref } from 'vue'

export interface ShortcutOptions {
  /** Scope where the shortcut is active (default: 'global') */
  scope?: 'global' | 'input' | 'terminal'
  /** Prevent default browser behavior (default: true) */
  preventDefault?: boolean
  /** Stop event propagation (default: false) */
  stopPropagation?: boolean
  /** Allow shortcut to trigger repeatedly when key is held (default: false) */
  allowRepeat?: boolean
  /** Description for help/UI display */
  description?: string
}

export interface ShortcutDefinition {
  key: string
  modifiers: ModifierState
  callback: (event: KeyboardEvent) => void
  options: Required<ShortcutOptions>
}

interface ModifierState {
  ctrl: boolean
  shift: boolean
  alt: boolean
  meta: boolean
}

interface RegisteredShortcuts {
  shortcuts: Map<string, ShortcutDefinition[]>
  activeScope: Ref<string>
}

const globalShortcuts: RegisteredShortcuts = {
  shortcuts: new Map(),
  activeScope: ref('global')
}

/**
 * Parse a keyboard shortcut string like "Ctrl+K" or "Cmd+Shift+/"
 * Returns the key and modifier state
 */
function parseShortcutString(shortcut: string): { key: string; modifiers: ModifierState } {
  const parts = shortcut.split('+').map((p) => p.trim().toLowerCase())
  const modifiers: ModifierState = {
    ctrl: false,
    shift: false,
    alt: false,
    meta: false
  }

  let key = ''

  for (const part of parts) {
    if (part === 'ctrl' || part === 'control') {
      modifiers.ctrl = true
    } else if (part === 'cmd' || part === 'meta' || part === 'command') {
      modifiers.meta = true
    } else if (part === 'shift') {
      modifiers.shift = true
    } else if (part === 'alt' || part === 'option') {
      modifiers.alt = true
    } else {
      key = part
    }
  }

  // Normalize key names
  if (key === 'escape' || key === 'esc') {
    key = 'Escape'
  } else if (key === '/') {
    key = '/'
  } else if (key.length === 1) {
    key = key.toUpperCase()
  }

  return { key, modifiers }
}

/**
 * Generate a unique identifier for a shortcut
 */
function getShortcutId(key: string, modifiers: ModifierState, scope: string): string {
  const mods = [
    modifiers.ctrl && 'ctrl',
    modifiers.meta && 'meta',
    modifiers.shift && 'shift',
    modifiers.alt && 'alt'
  ]
    .filter(Boolean)
    .join('-')
  return `${scope}:${mods}:${key}`
}

/**
 * Check if event matches shortcut modifiers
 */
function matchesModifiers(event: KeyboardEvent, modifiers: ModifierState): boolean {
  // On Mac, Ctrl key is usually represented as Meta for Cmd key shortcuts
  const isMac = navigator.platform.toUpperCase().indexOf('MAC') >= 0

  if (modifiers.ctrl) {
    // On Mac, "Ctrl" in shortcut definition means Cmd key
    if (isMac) {
      if (!event.metaKey) return false
    } else {
      if (!event.ctrlKey) return false
    }
  } else {
    // If Ctrl is not required, ensure neither Ctrl nor Meta is pressed
    if (isMac) {
      if (event.metaKey) return false
    } else {
      if (event.ctrlKey) return false
    }
  }

  if (modifiers.meta && !event.metaKey) return false
  if (modifiers.shift !== event.shiftKey) return false
  if (modifiers.alt !== event.altKey) return false

  return true
}

/**
 * Handle keyboard events and trigger matching shortcuts
 */
function handleKeyEvent(event: KeyboardEvent, type: 'down' | 'up') {
  if (type !== 'down') return

  // Ignore repeated events unless allowed
  if (event.repeat) {
    const matchingShortcuts = findMatchingShortcuts(event)
    const hasAllowRepeat = matchingShortcuts.some((s) => s.options.allowRepeat)
    if (!hasAllowRepeat) return
  }

  const matchingShortcuts = findMatchingShortcuts(event)

  for (const shortcut of matchingShortcuts) {
    // Check scope
    if (shortcut.options.scope !== 'global' && shortcut.options.scope !== globalShortcuts.activeScope.value) {
      continue
    }

    // Execute callback
    if (shortcut.options.preventDefault) {
      event.preventDefault()
    }
    if (shortcut.options.stopPropagation) {
      event.stopPropagation()
    }

    shortcut.callback(event)
    break // Only trigger first matching shortcut
  }
}

/**
 * Find shortcuts matching the current keyboard event
 */
function findMatchingShortcuts(event: KeyboardEvent): ShortcutDefinition[] {
  const key = event.key
  const allMatching: ShortcutDefinition[] = []

  for (const [, shortcuts] of globalShortcuts.shortcuts) {
    for (const shortcut of shortcuts) {
      if (shortcut.key === key && matchesModifiers(event, shortcut.modifiers)) {
        allMatching.push(shortcut)
      }
    }
  }

  return allMatching
}

/** 供设置页等展示当前已注册的快捷键（REQ-402）。 */
export function getRegisteredShortcutsSnapshot(): ShortcutDefinition[] {
  const all: ShortcutDefinition[] = []
  for (const [, shortcuts] of globalShortcuts.shortcuts) {
    all.push(...shortcuts)
  }
  return all
}

export function formatShortcutDefinitionForDisplay(d: ShortcutDefinition): string {
  const isMac = navigator.platform.toUpperCase().indexOf('MAC') >= 0
  const parts: string[] = []
  if (d.modifiers.ctrl) {
    parts.push(isMac ? 'Cmd' : 'Ctrl')
  }
  if (d.modifiers.meta && !d.modifiers.ctrl) {
    parts.push('Cmd')
  }
  if (d.modifiers.alt) {
    parts.push(isMac ? 'Opt' : 'Alt')
  }
  if (d.modifiers.shift) {
    parts.push('Shift')
  }
  const k =
    d.key === 'Escape'
      ? 'Esc'
      : d.key.length === 1
        ? d.key.toUpperCase()
        : d.key
  parts.push(k)
  return parts.join('+')
}

/**
 * Composable for registering keyboard shortcuts
 *
 * @example
 * ```ts
 * const { registerShortcut, unregisterShortcut, setScope } = useKeyboard()
 *
 * // Register a global shortcut
 * registerShortcut('Ctrl+K', () => console.log('Search opened'), { description: 'Open search' })
 *
 * // Register an input-scoped shortcut
 * registerShortcut('Escape', () => closeModal(), { scope: 'input' })
 *
 * // Change active scope
 * setScope('terminal')
 * ```
 */
export function useKeyboard() {
  let registeredIds: string[] = []

  const registerShortcut = (
    shortcutKey: string,
    callback: (event: KeyboardEvent) => void,
    options?: ShortcutOptions
  ) => {
    const { key, modifiers } = parseShortcutString(shortcutKey)
    const resolvedOptions: Required<ShortcutOptions> = {
      scope: options?.scope ?? 'global',
      preventDefault: options?.preventDefault ?? true,
      stopPropagation: options?.stopPropagation ?? false,
      allowRepeat: options?.allowRepeat ?? false,
      description: options?.description ?? ''
    }

    const id = getShortcutId(key, modifiers, resolvedOptions.scope)
    const definition: ShortcutDefinition = {
      key,
      modifiers,
      callback,
      options: resolvedOptions
    }

    // Add to shortcuts map
    const existing = globalShortcuts.shortcuts.get(id) ?? []
    existing.push(definition)
    globalShortcuts.shortcuts.set(id, existing)

    registeredIds.push(id)

    return id
  }

  const unregisterShortcut = (id: string) => {
    const shortcuts = globalShortcuts.shortcuts.get(id)
    if (shortcuts) {
      // Remove the last registered shortcut with this id
      if (shortcuts.length <= 1) {
        globalShortcuts.shortcuts.delete(id)
      } else {
        shortcuts.pop()
      }
    }
    registeredIds = registeredIds.filter((r) => r !== id)
  }

  const unregisterAll = () => {
    for (const id of registeredIds) {
      unregisterShortcut(id)
    }
    registeredIds = []
  }

  const setScope = (scope: string) => {
    globalShortcuts.activeScope.value = scope
  }

  const getScope = () => globalShortcuts.activeScope.value

  const getAllShortcuts = (): ShortcutDefinition[] => {
    const all: ShortcutDefinition[] = []
    for (const [, shortcuts] of globalShortcuts.shortcuts) {
      all.push(...shortcuts)
    }
    return all
  }

  // Setup global event listener on mount
  onMounted(() => {
    window.addEventListener('keydown', (e) => handleKeyEvent(e, 'down'))
    window.addEventListener('keyup', (e) => handleKeyEvent(e, 'up'))
  })

  // Cleanup on unmount
  onUnmounted(() => {
    unregisterAll()
  })

  return {
    registerShortcut,
    unregisterShortcut,
    unregisterAll,
    setScope,
    getScope,
    getAllShortcuts,
    activeScope: globalShortcuts.activeScope
  }
}

/**
 * Format a shortcut for display (e.g., "Ctrl+K" -> "Cmd+K" on Mac)
 */
export function formatShortcutForDisplay(shortcut: string): string {
  const isMac = navigator.platform.toUpperCase().indexOf('MAC') >= 0
  return shortcut
    .replace(/Ctrl/gi, isMac ? 'Cmd' : 'Ctrl')
    .replace(/Cmd/gi, isMac ? 'Cmd' : 'Ctrl')
    .replace(/Meta/gi, isMac ? 'Cmd' : 'Ctrl')
}