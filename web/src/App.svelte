<script>
  import { onMount } from 'svelte'
  import { get } from 'svelte/store'
  import { api } from './lib/api.js'
  import { enqueue, getCache, getQueue, removeQueued, setCache } from './lib/db.js'
  import { API, TODO_CACHE_KEY } from './lib/constants.js'
  import {
    user,
    projects,
    todos,
    selectedProjectId,
    online,
    currentView,
    error,
    success,
    selectedProject,
  } from './lib/stores.js'
  import AuthPage from './components/AuthPage.svelte'
  import Sidebar from './components/Sidebar.svelte'
  import TopBar from './components/TopBar.svelte'
  import QuickAdd from './components/QuickAdd.svelte'
  import TodoList from './components/TodoList.svelte'
  import ProfilePage from './components/ProfilePage.svelte'

  let loading = true
  let socket = null
  let menuOpen = false
  let projectMenuId = ''
  let dropdownPos = { top: 0, right: 0 }

  onMount(async () => {
    const cachedProjects = await getCache('projects', [])
    projects.set(cachedProjects)
    const cachedProjectId = await getCache('selectedProjectId', '')
    selectedProjectId.set(cachedProjectId)
    const cachedTodos = await getCache(TODO_CACHE_KEY(cachedProjectId), [])
    todos.set(cachedTodos)
    await bootstrap()
    window.addEventListener('online', handleOnline)
    window.addEventListener('offline', handleOffline)
    const closeMenus = () => { projectMenuId = ''; menuOpen = false }
    window.addEventListener('click', closeMenus)
    return () => {
      window.removeEventListener('online', handleOnline)
      window.removeEventListener('offline', handleOffline)
      window.removeEventListener('click', closeMenus)
      closeSocket()
    }
  })

  async function bootstrap() {
    loading = true
    try {
      const payload = await api(API.me)
      user.set(payload.user)
      await syncQueue()
      await loadProjects()
    } catch (err) {
      if (err.status !== 401) {
        error.set(err.message)
      }
    } finally {
      loading = false
    }
  }

  async function handleAuthenticated(e) {
    user.set(e.detail.user)
    await loadProjects()
  }

  async function logout() {
    await api(API.logout, { method: 'POST' }).catch(() => {})
    user.set(null)
    projects.set([])
    todos.set([])
    selectedProjectId.set('')
    currentView.set('todos')
    menuOpen = false
    closeSocket()
  }

  async function loadProjects() {
    const payload = await api(API.projects)
    const list = payload.projects || []
    projects.set(list)
    await setCache('projects', list)
    const $selectedProjectId = get(selectedProjectId)
    if (!$selectedProjectId && list.length) {
      selectedProjectId.set(list[0].id)
    }
    if (get(selectedProjectId)) {
      await selectProject(get(selectedProjectId))
    }
  }

  async function handleReloadProjects() {
    await loadProjects()
  }

  async function selectProject(projectId) {
    selectedProjectId.set(projectId)
    await setCache('selectedProjectId', projectId)
    const cached = await getCache(TODO_CACHE_KEY(projectId), [])
    todos.set(cached)
    closeSocket()
    if (get(online) && !projectId.startsWith('local-')) {
      await refreshTodos(projectId)
      connectSocket(projectId)
    }
  }

  async function refreshTodos(projectId) {
    const $id = projectId || get(selectedProjectId)
    if (!$id || $id.startsWith('local-')) return
    const payload = await api(API.projectTodos($id))
    const list = payload.todos || []
    todos.set(list)
    await setCache(TODO_CACHE_KEY($id), list)
  }

  async function runOrQueue(operation) {
    error.set('')
    if (!get(online) || operation.path.includes('local-')) {
      await enqueue(operation)
      return
    }
    try {
      await api(operation.path, { method: operation.method, body: operation.body })
      // Optimistic updates already applied — no need to refetch on own mutations.
      // WebSocket events from other users trigger refreshTodos separately.
    } catch (err) {
      if (err.status === 401) {
        user.set(null)
        return
      }
      if (!navigator.onLine) {
        await enqueue(operation)
        online.set(false)
        return
      }
      error.set(err.message)
    }
  }

  async function syncQueue() {
    if (!navigator.onLine) return
    for (const operation of await getQueue()) {
      if (operation.path.includes('local-')) {
        await removeQueued(operation.id)
        continue
      }
      try {
        await api(operation.path, { method: operation.method, body: operation.body })
        await removeQueued(operation.id)
      } catch (err) {
        if (err.status === 401) { user.set(null); return }
        return
      }
    }
  }

  async function handleOnline() {
    online.set(true)
    await syncQueue()
    if (get(user)) await loadProjects()
  }

  function handleOffline() {
    online.set(false)
    closeSocket()
  }

  function connectSocket(projectId) {
    const protocol = location.protocol === 'https:' ? 'wss' : 'ws'
    socket = new WebSocket(`${protocol}://${location.host}${API.projectWS(projectId)}`)
    socket.onmessage = async () => {
      await refreshTodos(projectId).catch(() => {})
    }
    socket.onclose = () => {
      if (get(online) && get(selectedProjectId) === projectId) {
        setTimeout(() => connectSocket(projectId), 1200)
      }
    }
  }

  function closeSocket() {
    if (socket) {
      socket.onclose = null
      socket.close()
      socket = null
    }
  }

  function handleOpenProjectMenu(e) {
    const { projectId, pos } = e.detail
    if (projectMenuId === projectId) {
      projectMenuId = ''
      return
    }
    dropdownPos = pos
    projectMenuId = projectId
  }

  function handleStartEditProject(e) {
    window.dispatchEvent(new CustomEvent('start-edit-project', { detail: e.detail }))
    projectMenuId = ''
  }

  async function handleDeleteProject(e) {
    const project = e.detail
    projectMenuId = ''
    if (!confirm(`Projekt "${project.title}" und alle Todos löschen?`)) return
    const $selectedProjectId = get(selectedProjectId)
    if ($selectedProjectId === project.id) {
      const next = get(projects).find((p) => p.id !== project.id)
      selectedProjectId.set(next?.id || '')
      todos.set([])
    }
    projects.update(($p) => $p.filter((p) => p.id !== project.id))
    await setCache('projects', get(projects))
    await runOrQueue({ method: 'DELETE', path: API.project(project.id), body: null })
  }
</script>

{#if loading}
  <main class="shell centered">
    <div class="card loader">Wird geladen...</div>
  </main>
{:else if !$user}
  <AuthPage on:authenticated={handleAuthenticated} />
{:else}
  <main class="app-layout" style={`--accent-color:${$selectedProject?.color || '#4f46e5'}`}>
    <Sidebar
      on:select-project={(e) => selectProject(e.detail)}
      on:open-project-menu={handleOpenProjectMenu}
      on:run-or-queue={(e) => runOrQueue(e.detail)}
      on:reload-projects={handleReloadProjects}
      on:error={(e) => error.set(e.detail)}
      on:clear-todos={() => todos.set([])}
      on:close-project-menu={() => (projectMenuId = '')}
    />

    <section class="workspace">
      <TopBar
        {menuOpen}
        on:toggle-menu={() => (menuOpen = !menuOpen)}
        on:open-profile={() => { currentView.set('profile'); menuOpen = false }}
        on:open-todos={() => { currentView.set('todos'); menuOpen = false }}
        on:logout={logout}
      />

      {#if $error}<p class="error">{$error}</p>{/if}
      {#if $success}<p class="success">{$success}</p>{/if}

      {#if $currentView === 'profile'}
        <ProfilePage />
      {:else}
        <QuickAdd on:run-or-queue={(e) => runOrQueue(e.detail)} on:error={(e) => error.set(e.detail)} />
        <TodoList on:run-or-queue={(e) => runOrQueue(e.detail)} />
      {/if}
    </section>
  </main>

  {#if projectMenuId}
    {@const menuProject = $projects.find((p) => p.id === projectMenuId)}
    {#if menuProject}
      <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
      <div class="proj-dropdown-portal" style={`top:${dropdownPos.top}px;right:${dropdownPos.right}px`} on:click|stopPropagation>
        <button on:click={() => handleStartEditProject({ detail: menuProject })}>✏️ Bearbeiten</button>
        <button class="danger" on:click={() => handleDeleteProject({ detail: menuProject })}>🗑 Löschen</button>
      </div>
    {/if}
  {/if}
{/if}

<style>
  .app-layout {
    --accent-color: #4f46e5;
    display: grid;
    grid-template-columns: 280px minmax(0, 1fr);
    gap: 16px;
    min-height: 100vh;
    padding: 16px 20px;
    background:
      radial-gradient(circle at 22% 0%, color-mix(in srgb, var(--accent-color), transparent 88%), transparent 30rem),
      radial-gradient(circle at 100% 18%, rgba(14, 165, 233, 0.1), transparent 24rem),
      linear-gradient(135deg, #f8fafc 0%, #f9fafb 48%, #eef2ff 100%);
  }

  .workspace {
    display: grid;
    align-content: start;
    gap: 14px;
  }

  .proj-dropdown-portal {
    position: fixed;
    z-index: 9999;
    min-width: 160px;
    background: #fff;
    border: 1px solid #e5e7eb;
    border-radius: 12px;
    box-shadow: 0 10px 40px rgba(0,0,0,0.18);
    padding: 5px;
    display: grid;
    gap: 2px;
  }

  .proj-dropdown-portal button {
    text-align: left;
    padding: 9px 12px;
    border-radius: 8px;
    background: transparent;
    color: #111827;
    font-size: 0.86rem;
    font-weight: 600;
  }

  .proj-dropdown-portal button:hover {
    background: #f3f4f6;
  }

  .proj-dropdown-portal button.danger {
    color: #dc2626;
  }

  .proj-dropdown-portal button.danger:hover {
    background: #fef2f2;
  }

  .centered {
    min-height: 100vh;
    display: grid;
    place-items: center;
    padding: 32px;
  }

  @media (max-width: 900px) {
    .app-layout {
      grid-template-columns: 1fr;
      padding: 12px;
      gap: 12px;
    }
  }
</style>
