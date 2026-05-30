import { writable, derived } from 'svelte/store'

export const user = writable(null)
export const projects = writable([])
export const todos = writable([])
export const columns = writable([])
export const selectedProjectId = writable('')
export const online = writable(typeof navigator !== 'undefined' ? navigator.onLine : true)
export const currentView = writable('todos')
export const error = writable('')
export const success = writable('')

// Theme store — defaults to the visitor's OS/browser preference and allows manual override.
function createThemeStore() {
  const stored = typeof localStorage !== 'undefined' ? localStorage.getItem('machino-theme') : null
  const storedTheme = stored === 'dark' || stored === 'light' ? stored : null
  const colorScheme = typeof window !== 'undefined'
    ? window.matchMedia?.('(prefers-color-scheme: dark)')
    : null
  const preferredTheme = colorScheme?.matches ? 'dark' : 'light'
  const initial = storedTheme || preferredTheme
  let followsSystemTheme = !storedTheme

  function applyTheme(next) {
    if (typeof document !== 'undefined') document.documentElement.dataset.theme = next
  }

  applyTheme(initial)

  const { subscribe, set, update } = writable(initial)

  colorScheme?.addEventListener?.('change', event => {
    if (!followsSystemTheme) return
    const next = event.matches ? 'dark' : 'light'
    applyTheme(next)
    set(next)
  })

  return {
    subscribe,
    toggle() {
      followsSystemTheme = false
      update(current => {
        const next = current === 'dark' ? 'light' : 'dark'
        if (typeof localStorage !== 'undefined') localStorage.setItem('machino-theme', next)
        applyTheme(next)
        return next
      })
    },
  }
}
export const theme = createThemeStore()

export const selectedProject = derived(
  [projects, selectedProjectId],
  ([$p, $id]) => $p.find((p) => p.id === $id) ?? null,
)
export const activeTodos = derived(todos, ($t) => $t.filter((t) => !t.completed))
export const completedTodos = derived(todos, ($t) => $t.filter((t) => t.completed))
export const openTodoCount = derived(activeTodos, ($t) => $t.length)
export const favoriteProjects = derived(projects, ($p) => $p.filter((p) => p.favorite))
export const otherProjects = derived(projects, ($p) => $p.filter((p) => !p.favorite))
export const avatarInitial = derived(user, ($u) =>
  ($u?.name || $u?.email || 'U').trim().slice(0, 1).toUpperCase(),
)
