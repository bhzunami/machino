export const API = {
  register: '/api/auth/register',
  login: '/api/auth/login',
  logout: '/api/auth/logout',
  passwordResetRequest: '/api/auth/password-reset/request',
  passwordResetConfirm: '/api/auth/password-reset/confirm',
  me: '/api/me',
  profile: '/api/profile',
  profilePassword: '/api/profile/password',
  projects: '/api/projects',
  project: (id) => `/api/projects/${id}`,
  projectFavorite: (id) => `/api/projects/${id}/favorite`,
  projectTodos: (id) => `/api/projects/${id}/todos`,
  projectTodosCompleted: (id) => `/api/projects/${id}/todos/completed`,
  projectTodosReorder: (id) => `/api/projects/${id}/todos/reorder`,
  projectWS: (id) => `/api/projects/${id}/ws`,
  todo: (id) => `/api/todos/${id}`,
}

export const PRIORITY = { HIGH: 'high', NORMAL: 'normal', LOW: 'low' }
export const DEFAULT_PROJECT_COLOR = '#4f46e5'
export const TODO_CACHE_KEY = (id) => `todos:${id}`
