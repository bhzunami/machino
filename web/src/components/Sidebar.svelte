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

<aside class="sidebar card">
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
    top: 16px;
    height: calc(100vh - 32px);
    display: flex;
    flex-direction: column;
    gap: 16px;
    padding: 16px;
    overflow: auto;
    border-color: rgba(226, 232, 240, 0.9);
    background:
      linear-gradient(180deg, color-mix(in srgb, var(--accent-color), white 94%), rgba(255, 255, 255, 0.96) 120px),
      #fff;
  }

  .brand,
  .sidebar-title {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
  }

  .brand {
    padding: 2px 0;
  }

  .brand-mark {
    display: flex;
    gap: 8px;
  }

  .brand-dot {
    width: 10px;
    height: 10px;
    border-radius: 999px;
  }
  .brand-dot:nth-child(1) { background: #a5b4fc; }
  .brand-dot:nth-child(2) { background: #7c3aed; }
  .brand-dot:nth-child(3) { background: #38bdf8; }

  .brand-name {
    flex: 1;
    font-weight: 800;
    font-size: 0.95rem;
    color: #111827;
    letter-spacing: -0.01em;
  }

  .status-badge {
    border-radius: 999px;
    padding: 3px 9px;
    background: #dcfce7;
    color: #166534;
    font-weight: 700;
    font-size: 0.72rem;
    letter-spacing: 0.02em;
  }

  .status-badge.offline {
    background: #fef3c7;
    color: #92400e;
  }

  .sidebar-title h2 {
    margin: 0;
  }

  .icon-btn {
    width: 36px;
    height: 36px;
    border-radius: 999px;
    background: linear-gradient(135deg, var(--accent-color), #111827);
    color: #fff;
    font-size: 1.35rem;
    font-weight: 850;
    line-height: 1;
  }

  .project-form {
    border: 1px solid #e5e7eb;
    border-radius: 20px;
    padding: 16px;
    background: rgba(248, 250, 252, 0.92);
  }

  .stack {
    display: grid;
    gap: 14px;
  }

  .projects {
    display: grid;
    gap: 20px;
  }

  .project-section {
    display: grid;
    gap: 6px;
  }

  .section-label {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 0 10px 4px;
    color: #6b7280;
    font-size: 0.72rem;
    font-weight: 850;
    letter-spacing: 0.08em;
    text-transform: uppercase;
  }

  .section-label span {
    display: grid;
    place-items: center;
    width: 18px;
    height: 18px;
    border-radius: 6px;
    background: #f3f4f6;
    color: #6b7280;
    font-size: 0.72rem;
  }

  .project-item {
    position: relative;
    display: grid;
    grid-template-columns: minmax(0, 1fr) auto;
    align-items: center;
    border-radius: 12px;
    background: transparent;
    overflow: visible;
  }

  .project-item::before {
    content: '';
    position: absolute;
    inset: 8px auto 8px 0;
    width: 3px;
    border-radius: 999px;
    background: transparent;
  }

  .project-item.active,
  .project-item:hover {
    background: color-mix(in srgb, var(--project-color), white 91%);
  }

  .project-item.active::before {
    background: var(--project-color);
  }

  .project-actions {
    display: flex;
    align-items: center;
    gap: 0;
  }

  .proj-menu-wrap {
    position: relative;
  }

  .proj-dots {
    padding: 8px 10px;
    background: transparent;
    color: #6b7280;
    font-size: 1.1rem;
    letter-spacing: 0.05em;
    border-radius: 8px;
    opacity: 0;
    transition: opacity 0.15s, color 0.12s, background 0.12s;
  }

  .project-item:hover .proj-dots,
  .project-item.active .proj-dots {
    opacity: 1;
  }

  .proj-dots:hover {
    color: #111827;
    background: rgba(0,0,0,0.06);
    opacity: 1;
  }

  .project-edit-form {
    grid-column: 1 / -1;
    display: grid;
    gap: 7px;
    padding: 10px 12px;
    background: color-mix(in srgb, var(--project-color, #4f46e5), white 93%);
    border-radius: 10px;
    margin: 2px 0;
  }

  .project-edit-form input {
    border-radius: 7px;
    border: 1.5px solid #e2e8f0;
    padding: 6px 9px;
    font-size: 0.83rem;
    background: #fff;
    color: #111827;
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
    border-radius: 7px;
    border: 1.5px solid #e2e8f0;
    cursor: pointer;
  }

  .project-select {
    display: grid;
    grid-template-columns: auto minmax(0, 1fr);
    align-items: center;
    gap: 10px;
    padding: 10px 8px 10px 12px;
    background: transparent;
    color: #111827;
    text-align: left;
    font-weight: 700;
  }

  .star {
    padding: 10px 12px;
    background: transparent;
    color: #9ca3af;
    font-size: 1.2rem;
  }

  .star.active,
  .star:hover {
    color: #f59e0b;
  }

  .dot {
    width: 11px;
    height: 11px;
    border-radius: 999px;
    background: var(--project-color);
    box-shadow: 0 0 0 3px color-mix(in srgb, var(--project-color), transparent 82%);
  }

  @media (max-width: 900px) {
    .sidebar {
      position: static;
      height: auto;
    }
  }
</style>
