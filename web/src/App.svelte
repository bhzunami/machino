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
    columns,
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
  import AdminPage from './components/AdminPage.svelte'
  import SetupWizard from './components/SetupWizard.svelte'
  import ShareModal from './components/ShareModal.svelte'
  import EditProjectModal from './components/EditProjectModal.svelte'

  let loading = true
  let socket = null
  let menuOpen = false
  let sidebarOpen = false
  let sidebarRef
  let projectMenuId = ''
  let dropdownPos = { top: 0, right: 0 }
  let shareProject = null
  let editProject = null
  let setupRequired = false
  let setupDefaults = {}

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
      const setup = await api(API.setupStatus)
      setupRequired = !!setup.setupRequired
      setupDefaults = setup.settings || {}
      if (setupRequired) {
        return
      }
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
    setupRequired = false
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

  async function selectProject(projectId) {
    selectedProjectId.set(projectId)
    await setCache('selectedProjectId', projectId)
    const cached = await getCache(TODO_CACHE_KEY(projectId), [])
    todos.set(cached)
    columns.set([])
    closeSocket()
    sidebarOpen = false
    if (get(online) && !projectId.startsWith('local-')) {
      await refreshTodos(projectId)
      connectSocket(projectId)
    }
  }

  async function refreshTodos(projectId) {
    const $id = projectId || get(selectedProjectId)
    if (!$id || $id.startsWith('local-')) return
    const [todoPayload, colPayload] = await Promise.all([
      api(API.projectTodos($id)),
      api(API.projectColumns($id)),
    ])
    const list = todoPayload.todos || []
    todos.set(list)
    await setCache(TODO_CACHE_KEY($id), list)
    columns.set(colPayload.columns || [])
  }

  async function runOrQueue(operation) {
    error.set('')
    if (!get(online) || operation.path.includes('local-')) {
      await enqueue(operation)
      return
    }
    try {
      await api(operation.path, { method: operation.method, body: operation.body })
      if (shouldRefreshProjectData(operation)) {
        await refreshTodos(get(selectedProjectId))
      }
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

  function shouldRefreshProjectData(operation) {
    return (
      operation.path.includes('/todos') ||
      operation.path.includes('/columns') ||
      Object.prototype.hasOwnProperty.call(operation.body || {}, 'columnId') ||
      Object.prototype.hasOwnProperty.call(operation.body || {}, 'clearColumn')
    )
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
    editProject = e.detail
    projectMenuId = ''
  }

  async function handleSaveEditProject(e) {
    const { project, form } = e.detail
    // Optimistic update immediately
    projects.update(($p) => $p.map((p) => p.id === project.id ? { ...p, ...form } : p))
    await setCache('projects', get(projects))
    editProject = null
    // Persist to server
    await runOrQueue({ method: 'PUT', path: API.project(project.id), body: form })
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
  {#if setupRequired}
    <SetupWizard defaults={setupDefaults} on:authenticated={handleAuthenticated} />
  {:else}
    <AuthPage registrationEnabled={setupDefaults.registrationEnabled !== false} on:authenticated={handleAuthenticated} />
  {/if}
{:else}
  <main class="app-layout" style={`--accent-color:${$selectedProject?.color || '#4f46e5'}`}>
    <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
    {#if sidebarOpen}
      <div class="sidebar-overlay" on:click={() => (sidebarOpen = false)}></div>
    {/if}
    <Sidebar
      bind:this={sidebarRef}
      open={sidebarOpen}
      on:select-project={(e) => selectProject(e.detail)}
      on:open-project-menu={handleOpenProjectMenu}
      on:run-or-queue={(e) => runOrQueue(e.detail)}
      on:reload-projects={loadProjects}
      on:error={(e) => error.set(e.detail)}
      on:clear-todos={() => todos.set([])}
      on:close-project-menu={() => (projectMenuId = '')}
    />

    <section class="workspace">
      <TopBar
        {menuOpen}
        {sidebarOpen}
        on:toggle-menu={() => (menuOpen = !menuOpen)}
        on:toggle-sidebar={() => (sidebarOpen = !sidebarOpen)}
        on:open-profile={() => { currentView.set('profile'); menuOpen = false }}
        on:open-admin={() => { currentView.set('admin'); menuOpen = false }}
        on:open-todos={() => { currentView.set('todos'); menuOpen = false }}
        on:logout={logout}
      />

      {#if $error}<p class="error">{$error}</p>{/if}
      {#if $success}<p class="success">{$success}</p>{/if}

      {#if $currentView === 'admin' && $user.role === 'admin'}
        <AdminPage />
      {:else if $currentView === 'profile'}
        <ProfilePage />
      {:else}
        <div class="content-area">
          <QuickAdd on:run-or-queue={(e) => runOrQueue(e.detail)} on:error={(e) => error.set(e.detail)} />
          <TodoList on:run-or-queue={(e) => runOrQueue(e.detail)} on:new-project={() => { sidebarOpen = true; sidebarRef?.openProjectForm() }} />
        </div>
      {/if}
    </section>
  </main>

  {#if projectMenuId}
    {@const menuProject = $projects.find((p) => p.id === projectMenuId)}
    {#if menuProject}
      <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
      <div class="proj-dropdown-portal" style={`top:${dropdownPos.top}px;right:${dropdownPos.right}px`} on:click|stopPropagation>
        <button on:click={() => handleStartEditProject({ detail: menuProject })}>✏️ Bearbeiten</button>
        {#if menuProject.isOwner}
          <button on:click={() => { shareProject = menuProject; projectMenuId = '' }}>🔗 Teilen</button>
          <button class="danger" on:click={() => handleDeleteProject({ detail: menuProject })}>🗑 Löschen</button>
        {/if}
      </div>
    {/if}
  {/if}
{/if}

{#if shareProject}
  <ShareModal project={shareProject} on:close={() => (shareProject = null)} />
{/if}

{#if editProject}
  <EditProjectModal
    project={editProject}
    on:save={handleSaveEditProject}
    on:close={() => (editProject = null)}
    on:run-or-queue={(e) => runOrQueue(e.detail)}
  />
{/if}

<style>
  .app-layout {
    --accent-color: #6366f1;
    --accent-glow:  rgba(99, 102, 241, 0.3);
    display: grid;
    grid-template-columns: 280px minmax(0, 1fr);
    min-height: 100vh;
    background: var(--bg);
    position: relative;
  }

  .app-layout::before {
    content: '';
    position: fixed;
    inset: 0;
    background:
      radial-gradient(ellipse at 15% 5%,  color-mix(in srgb, var(--accent-color), transparent 70%) 0%, transparent 50%),
      radial-gradient(ellipse at 88% 12%, rgba(56,189,248,0.1)  0%, transparent 40%),
      radial-gradient(ellipse at 55% 95%, rgba(139,92,246,0.08) 0%, transparent 50%);
    pointer-events: none;
    z-index: 0;
    animation: bg-drift 14s ease-in-out infinite alternate;
    transition: background 0.6s ease;
  }

  @keyframes bg-drift {
    from { opacity: 1; transform: scale(1); }
    to   { opacity: 0.7; transform: scale(1.06) translateY(-8px); }
  }

  .workspace {
    display: grid;
    align-content: start;
    gap: 0;
    padding: 0 0 40px;
    position: relative;
    z-index: 1;
    min-width: 0;
  }

  .proj-dropdown-portal {
    position: fixed;
    z-index: 9999;
    min-width: 164px;
    background: var(--bg-2);
    border: 1px solid var(--border-hover);
    border-radius: 14px;
    box-shadow: 0 24px 64px rgba(0,0,0,0.7), 0 0 0 1px rgba(255,255,255,0.04);
    padding: 5px;
    display: grid;
    gap: 2px;
    backdrop-filter: blur(24px);
    animation: drop-in 0.13s cubic-bezier(0.16,1,0.3,1);
  }

  @keyframes drop-in {
    from { opacity: 0; transform: translateY(-8px) scale(0.96); }
    to   { opacity: 1; transform: translateY(0)  scale(1); }
  }

  .proj-dropdown-portal button {
    text-align: left;
    padding: 9px 12px;
    border-radius: 10px;
    background: transparent;
    color: var(--text);
    font-size: 0.86rem;
    font-weight: 600;
    transition: background 0.12s;
  }

  .proj-dropdown-portal button:hover { background: var(--glass-hover); }

  .proj-dropdown-portal button.danger { color: #f87171; }

  .proj-dropdown-portal button.danger:hover { background: rgba(248,113,113,0.12); }

  .centered {
    min-height: 100vh;
    display: grid;
    place-items: center;
    padding: 32px;
  }

  .content-area {
    display: grid;
    align-content: start;
    gap: 20px;
    padding: 28px 28px 40px;
  }

  @media (max-width: 900px) {
    .app-layout { grid-template-columns: 1fr; }
    .workspace { padding: 0 0 32px; }
    .content-area { padding: 18px 16px 32px; }
  }

  .sidebar-overlay {
    display: none;
    position: fixed;
    inset: 0;
    background: rgba(0,0,0,0.55);
    backdrop-filter: blur(2px);
    z-index: 199;
    animation: fade-in 0.2s ease;
  }

  @keyframes fade-in {
    from { opacity: 0; }
    to   { opacity: 1; }
  }

  @media (max-width: 900px) {
    .sidebar-overlay { display: block; }
  }
</style>
