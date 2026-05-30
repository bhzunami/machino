export const API = {
  register: '/api/auth/register',
  login: '/api/auth/login',
  logout: '/api/auth/logout',
  passwordResetRequest: '/api/auth/password-reset/request',
  passwordResetConfirm: '/api/auth/password-reset/confirm',
  me: '/api/me',
  profile: '/api/profile',
  profilePassword: '/api/profile/password',
  adminUsers: '/api/admin/users',
  adminUser: (id) => `/api/admin/users/${id}`,
  adminUserPassword: (id) => `/api/admin/users/${id}/password`,
  projects: '/api/projects',
  project: (id) => `/api/projects/${id}`,
  projectFavorite: (id) => `/api/projects/${id}/favorite`,
  projectTodos: (id) => `/api/projects/${id}/todos`,
  projectTodosCompleted: (id) => `/api/projects/${id}/todos/completed`,
  projectTodosReorder: (id) => `/api/projects/${id}/todos/reorder`,
  projectColumns: (id) => `/api/projects/${id}/columns`,
  projectColumnsReorder: (id) => `/api/projects/${id}/columns/reorder`,
  projectWS: (id) => `/api/projects/${id}/ws`,
  projectMembers: (id) => `/api/projects/${id}/members`,
  projectMember: (pid, uid) => `/api/projects/${pid}/members/${uid}`,
  usersSearch: '/api/users/search',
  todo: (id) => `/api/todos/${id}`,
  column: (id) => `/api/columns/${id}`,
}

export const PRIORITY = { HIGH: 'high', NORMAL: 'normal', LOW: 'low' }
export const DEFAULT_PROJECT_COLOR = '#4f46e5'
export const TODO_CACHE_KEY = (id) => `todos:${id}`
