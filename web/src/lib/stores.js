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

// Theme store — persisted in localStorage, applied to <html data-theme>
function createThemeStore() {
  const stored = typeof localStorage !== 'undefined' ? localStorage.getItem('machino-theme') : null
  const initial = stored || 'dark'
  if (typeof document !== 'undefined') document.documentElement.dataset.theme = initial
  const { subscribe, update } = writable(initial)
  return {
    subscribe,
    toggle() {
      update(current => {
        const next = current === 'dark' ? 'light' : 'dark'
        if (typeof localStorage !== 'undefined') localStorage.setItem('machino-theme', next)
        if (typeof document !== 'undefined') document.documentElement.dataset.theme = next
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
