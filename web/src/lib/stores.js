import { writable, derived } from 'svelte/store'

export const user = writable(null)
export const projects = writable([])
export const todos = writable([])
export const selectedProjectId = writable('')
export const online = writable(typeof navigator !== 'undefined' ? navigator.onLine : true)
export const currentView = writable('todos')
export const error = writable('')
export const success = writable('')

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
