import { ref, readonly } from 'vue'

export interface ContextMenuItem {
  id: string
  label: string
  icon?: string
  danger?: boolean
  action?: () => void
}

interface ContextMenuState {
  open: boolean
  x: number
  y: number
  entries: ContextMenuItem[]
}

const state = ref<ContextMenuState>({
  open: false,
  x: 0,
  y: 0,
  entries: []
})

export function useContextMenu() {
  const openMenu = (x: number, y: number, entries: ContextMenuItem[]) => {
    state.value = { open: true, x, y, entries }
  }

  const closeMenu = () => {
    state.value.open = false
  }

  const runAction = async (entry: ContextMenuItem) => {
    closeMenu()
    entry.action?.()
  }

  return {
    state: readonly(state),
    openMenu,
    closeMenu,
    runAction
  }
}