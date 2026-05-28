<script>
  import { createEventDispatcher } from 'svelte'
  import { get } from 'svelte/store'
  import { api } from '../lib/api.js'
  import { enqueue, setCache } from '../lib/db.js'
  import { API, DEFAULT_PROJECT_COLOR } from '../lib/constants.js'
  import {
    projects,
    selectedProjectId,
    online,
    favoriteProjects,
    otherProjects,
  } from '../lib/stores.js'

  const dispatch = createEventDispatcher()

  export let open = false

  let showProjectForm = false
  let projectForm = { title: '', description: '', color: DEFAULT_PROJECT_COLOR }
  let editingProjectId = ''
  let editProjectForm = { title: '', description: '', color: DEFAULT_PROJECT_COLOR }

  async function createProject() {
    if (!projectForm.title.trim()) {
      dispatch('error', 'Projekt-Titel ist Pflicht.')
      return
    }
    const optimistic = {
      id: `local-${Date.now()}`,
      ...projectForm,
      favorite: false,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    }
    projects.update(($p) => [optimistic, ...$p])
    selectedProjectId.set(optimistic.id)
    await setCache('projects', get(projects))
    try {
      if (!get(online)) {
        await enqueue({ method: 'POST', path: API.projects, body: projectForm })
      } else {
        await api(API.projects, { method: 'POST', body: projectForm })
        dispatch('reload-projects')
      }
      projectForm = { title: '', description: '', color: DEFAULT_PROJECT_COLOR }
      showProjectForm = false
    } catch (err) {
      dispatch('error', err.message)
    }
  }

  async function setFavorite(project) {
    const favorite = !project.favorite
    projects.update(($p) => $p.map((item) => (item.id === project.id ? { ...item, favorite } : item)))
    await setCache('projects', get(projects))
    dispatch('run-or-queue', {
      method: 'PUT',
      path: API.projectFavorite(project.id),
      body: { favorite },
    })
  }

  function openProjectMenu(e, projectId) {
    e.stopPropagation()
    if (get(projects).find((p) => p.id === projectId) === undefined) return
    const btn = e.currentTarget
    const rect = btn.getBoundingClientRect()
    const pos = { top: rect.bottom + 6, right: window.innerWidth - rect.right }
    dispatch('open-project-menu', { projectId, pos })
    editingProjectId = ''
  }

  function startEditProject(project) {
    editingProjectId = project.id
    editProjectForm = {
      title: project.title,
      description: project.description || '',
      color: project.color || DEFAULT_PROJECT_COLOR,
    }
  }

  async function saveEditProject(project) {
    if (!editProjectForm.title.trim()) return
    projects.update(($p) =>
      $p.map((p) => (p.id === project.id ? { ...p, ...editProjectForm } : p)),
    )
    await setCache('projects', get(projects))
    editingProjectId = ''
    dispatch('run-or-queue', {
      method: 'PUT',
      path: API.project(project.id),
      body: editProjectForm,
    })
  }

  async function deleteProject(e, project) {
    e.stopPropagation()
    dispatch('close-project-menu')
    if (!confirm(`Projekt "${project.title}" und alle Todos löschen?`)) return
    const $selectedProjectId = get(selectedProjectId)
    if ($selectedProjectId === project.id) {
      const next = get(projects).find((p) => p.id !== project.id)
      selectedProjectId.set(next?.id || '')
      dispatch('clear-todos')
    }
    projects.update(($p) => $p.filter((p) => p.id !== project.id))
    await setCache('projects', get(projects))
    dispatch('run-or-queue', {
      method: 'DELETE',
      path: API.project(project.id),
      body: null,
    })
  }

  function selectProject(projectId) {
    dispatch('select-project', projectId)
  }
</script>

<aside class="sidebar" class:open>
  <div class="brand">
    <div class="brand-mark">
      <span class="brand-dot"></span>
      <span class="brand-dot"></span>
      <span class="brand-dot"></span>
    </div>
    <span class="brand-name">Mach I No</span>
    <span class:offline={!$online} class="status-badge">{$online ? 'Online' : 'Offline'}</span>
  </div>

  <div class="sidebar-title">
    <h2>Projekte</h2>
    <button class="icon-btn" aria-label="Projekt erstellen" on:click={() => (showProjectForm = !showProjectForm)}>
      {showProjectForm ? '×' : '+'}
    </button>
  </div>

  {#if showProjectForm}
    <form class="project-form stack" on:submit|preventDefault={createProject}>
      <label>Titel <input bind:value={projectForm.title} required /></label>
      <label>Beschreibung <textarea bind:value={projectForm.description}></textarea></label>
      <label>Farbe <input bind:value={projectForm.color} type="color" /></label>
      <button class="btn" type="submit">Projekt hinzufügen</button>
    </form>
  {/if}

  <nav class="projects" aria-label="Projekte">
    {#if $favoriteProjects.length}
      <section class="project-section" aria-label="Favoriten">
        <div class="section-label">
          <span>★</span>
          <strong>Favoriten</strong>
        </div>
        {#each $favoriteProjects as project}
          <div
            class="project-item"
            class:active={project.id === $selectedProjectId}
            style={`--project-color:${project.color}`}
          >
            {#if editingProjectId === project.id}
              <form class="project-edit-form" on:submit|preventDefault={() => saveEditProject(project)}>
                <input bind:value={editProjectForm.title} placeholder="Titel" required />
                <input bind:value={editProjectForm.description} placeholder="Beschreibung" />
                <div class="edit-form-row">
                  <input bind:value={editProjectForm.color} type="color" />
                  <button class="btn" type="submit">Speichern</button>
                  <button class="btn secondary" type="button" on:click={() => (editingProjectId = '')}>Abbrechen</button>
                </div>
              </form>
            {:else}
              <button class="project-select" on:click={() => selectProject(project.id)}>
                <span class="dot"></span>
                <span>{project.title}</span>
              </button>
              <div class="project-actions">
                <button class="star active" aria-label="Favorit entfernen" on:click={() => setFavorite(project)}>★</button>
                <div class="proj-menu-wrap">
                  <button class="proj-dots" aria-label="Optionen" on:click={(e) => openProjectMenu(e, project.id)}>⋯</button>
                </div>
              </div>
            {/if}
          </div>
        {/each}
      </section>
    {/if}

    <section class="project-section" aria-label="Weitere Projekte">
      <div class="section-label">
        <span>▦</span>
        <strong>{$favoriteProjects.length ? 'Weitere Projekte' : 'Alle Projekte'}</strong>
      </div>
      {#each $otherProjects as project}
        <div
          class="project-item"
          class:active={project.id === $selectedProjectId}
          style={`--project-color:${project.color}`}
        >
          {#if editingProjectId === project.id}
            <form class="project-edit-form" on:submit|preventDefault={() => saveEditProject(project)}>
              <input bind:value={editProjectForm.title} placeholder="Titel" required />
              <input bind:value={editProjectForm.description} placeholder="Beschreibung" />
              <div class="edit-form-row">
                <input bind:value={editProjectForm.color} type="color" />
                <button class="btn" type="submit">Speichern</button>
                <button class="btn secondary" type="button" on:click={() => (editingProjectId = '')}>Abbrechen</button>
              </div>
            </form>
          {:else}
            <button class="project-select" on:click={() => selectProject(project.id)}>
              <span class="dot"></span>
              <span>{project.title}</span>
            </button>
            <div class="project-actions">
              <button class="star" aria-label="Als Favorit markieren" on:click={() => setFavorite(project)}>☆</button>
              <div class="proj-menu-wrap">
                <button class="proj-dots" aria-label="Optionen" on:click={(e) => openProjectMenu(e, project.id)}>⋯</button>
              </div>
            </div>
          {/if}
        </div>
      {/each}
    </section>
  </nav>
</aside>

<!-- Portal dropdown is rendered in App.svelte; expose startEditProject for it -->
<svelte:window on:start-edit-project={(e) => startEditProject(e.detail)} />

<style>
  .sidebar {
    position: sticky;
    top: 0;
    height: 100vh;
    display: flex;
    flex-direction: column;
    gap: 0;
    overflow: auto;
    border-right: 1px solid var(--border);
    background: rgba(8,8,15,0.7);
    backdrop-filter: blur(24px);
    -webkit-backdrop-filter: blur(24px);
    z-index: 10;
  }


  .brand, .sidebar-title {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
  }

  .brand { padding: 4px 4px 2px; }

  .brand-mark { display: flex; gap: 7px; align-items: center; }

  .brand-dot {
    width: 9px;
    height: 9px;
    border-radius: 999px;
    animation: dot-pulse 2.4s ease-in-out infinite;
  }
  .brand-dot:nth-child(1) { background: #818cf8; animation-delay: 0s; }
  .brand-dot:nth-child(2) { background: #a78bfa; animation-delay: 0.35s; }
  .brand-dot:nth-child(3) { background: #38bdf8; animation-delay: 0.7s; }

  @keyframes dot-pulse {
    0%, 100% { opacity: 1; transform: scale(1); }
    50%       { opacity: 0.45; transform: scale(0.78); }
  }

  .brand-name {
    flex: 1;
    font-weight: 800;
    font-size: 0.92rem;
    color: var(--text);
    letter-spacing: -0.01em;
  }

  .status-badge {
    border-radius: 999px;
    padding: 3px 9px;
    background: rgba(74,222,128,0.12);
    border: 1px solid rgba(74,222,128,0.22);
    color: #86efac;
    font-weight: 700;
    font-size: 0.68rem;
    letter-spacing: 0.04em;
  }
  .status-badge.offline {
    background: rgba(251,191,36,0.1);
    border-color: rgba(251,191,36,0.2);
    color: #fde68a;
  }

  .sidebar-title h2 { margin: 0; }

  .icon-btn {
    width: 32px;
    height: 32px;
    border-radius: 10px;
    background: var(--glass-md);
    border: 1px solid var(--border);
    color: var(--text-muted);
    font-size: 1.3rem;
    font-weight: 400;
    line-height: 1;
    display: grid;
    place-items: center;
    transition: background 0.15s, border-color 0.15s, color 0.15s;
  }
  .icon-btn:hover {
    background: color-mix(in srgb, var(--accent-color), transparent 78%);
    border-color: color-mix(in srgb, var(--accent-color), transparent 55%);
    color: color-mix(in srgb, var(--accent-color), white 30%);
  }

  .project-form {
    border: 1px solid var(--border);
    border-radius: 16px;
    padding: 14px;
    background: var(--glass);
    animation: slide-down 0.18s cubic-bezier(0.16,1,0.3,1);
  }

  @keyframes slide-down {
    from { opacity: 0; transform: translateY(-8px); }
    to   { opacity: 1; transform: translateY(0); }
  }

  .stack { display: grid; gap: 12px; }

  .projects { display: grid; gap: 18px; }

  .project-section { display: grid; gap: 4px; }

  .section-label {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 0 8px 6px;
    color: var(--text-faint);
    font-size: 0.68rem;
    font-weight: 800;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    border-bottom: 1px solid var(--border);
    margin-bottom: 4px;
  }

  .section-label span {
    display: grid;
    place-items: center;
    width: 18px;
    height: 18px;
    border-radius: 6px;
    background: var(--glass-md);
    color: var(--text-faint);
    font-size: 0.7rem;
    font-weight: 700;
  }

  .project-item {
    position: relative;
    display: grid;
    grid-template-columns: minmax(0, 1fr) auto;
    align-items: center;
    border-radius: 12px;
    background: transparent;
    overflow: visible;
    transition: background 0.15s;
  }

  .project-item::before {
    content: '';
    position: absolute;
    inset: 10px auto 10px 0;
    width: 3px;
    border-radius: 999px;
    background: transparent;
    transition: background 0.2s, box-shadow 0.2s;
  }

  .project-item:hover {
    background: var(--glass);
  }

  .project-item.active {
    background: color-mix(in srgb, var(--project-color, #6366f1), transparent 88%);
  }

  .project-item.active::before {
    background: var(--project-color, #6366f1);
    box-shadow: 0 0 8px var(--project-color, #6366f1);
  }

  .project-actions {
    display: flex;
    align-items: center;
  }

  .proj-menu-wrap { position: relative; }

  .proj-dots {
    padding: 8px 10px;
    background: transparent;
    color: var(--text-faint);
    font-size: 1rem;
    letter-spacing: 0.05em;
    border-radius: 8px;
    opacity: 0;
    transition: opacity 0.15s, color 0.12s, background 0.12s;
  }
  .project-item:hover .proj-dots,
  .project-item.active .proj-dots { opacity: 1; }
  .proj-dots:hover { color: var(--text); background: var(--glass-md); opacity: 1; }

  .project-edit-form {
    grid-column: 1 / -1;
    display: grid;
    gap: 8px;
    padding: 10px 12px;
    background: color-mix(in srgb, var(--project-color, #6366f1), transparent 88%);
    border: 1px solid color-mix(in srgb, var(--project-color, #6366f1), transparent 70%);
    border-radius: 12px;
    margin: 4px 0;
    animation: slide-down 0.15s ease;
  }

  .project-edit-form input {
    border-radius: 8px;
    padding: 7px 10px;
    font-size: 0.84rem;
    background: var(--glass-md);
    color: var(--text);
    border-color: var(--border);
  }

  .edit-form-row {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .edit-form-row input[type='color'] {
    width: 32px;
    height: 32px;
    padding: 2px;
    border-radius: 8px;
    border: 1px solid var(--border);
    cursor: pointer;
    background: transparent;
  }

  .project-select {
    display: grid;
    grid-template-columns: auto minmax(0, 1fr);
    align-items: center;
    gap: 10px;
    padding: 9px 6px 9px 10px;
    background: transparent;
    color: var(--text);
    text-align: left;
    font-weight: 600;
    font-size: 0.88rem;
    transition: color 0.15s;
  }
  .project-item.active .project-select { color: var(--text); }

  .star {
    padding: 8px 10px;
    background: transparent;
    color: var(--text-faint);
    font-size: 1rem;
    transition: color 0.15s, transform 0.15s;
  }
  .star.active { color: #fbbf24; filter: drop-shadow(0 0 6px rgba(251,191,36,0.6)); }
  .star:hover  { color: #fbbf24; transform: scale(1.2); }

  .dot {
    width: 10px;
    height: 10px;
    border-radius: 999px;
    background: var(--project-color, #6366f1);
    box-shadow: 0 0 0 3px color-mix(in srgb, var(--project-color, #6366f1), transparent 75%),
                0 0 10px color-mix(in srgb, var(--project-color, #6366f1), transparent 60%);
    flex-shrink: 0;
    transition: box-shadow 0.2s;
  }

  @media (max-width: 900px) {
    .sidebar {
      position: fixed;
      top: 0;
      left: 0;
      height: 100dvh;
      width: 280px;
      z-index: 200;
      transform: translateX(-100%);
      transition: transform 0.28s cubic-bezier(0.16, 1, 0.3, 1);
      border-right: 1px solid var(--border);
      box-shadow: 8px 0 40px rgba(0,0,0,0.5);
    }
    .sidebar.open {
      transform: translateX(0);
    }
  }
</style>
