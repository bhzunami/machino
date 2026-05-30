<script>
  import { createEventDispatcher } from 'svelte'
  import { get } from 'svelte/store'
  import CreateProjectModal from './CreateProjectModal.svelte'
  import { api } from '../lib/api.js'
  import { enqueue, setCache } from '../lib/db.js'
  import { API, DEFAULT_PROJECT_COLOR } from '../lib/constants.js'
  import {
    projects,
    selectedProjectId,
    online,
    favoriteProjects,
    otherProjects,
    theme,
  } from '../lib/stores.js'

  const dispatch = createEventDispatcher()

  export let open = false

  let showProjectForm = false
  let projectFormSaving = false
  let projectFormError = ''

  function openProjectForm() {
    projectFormError = ''
    showProjectForm = true
  }

  function closeProjectForm() {
    if (projectFormSaving) return
    projectFormError = ''
    showProjectForm = false
  }

  async function createProject(e) {
    projectFormError = ''
    const form = {
      ...e.detail.form,
      title: e.detail.form.title.trim(),
      description: e.detail.form.description.trim(),
    }
    const columnTitles = e.detail.columns || []
    if (!form.title) {
      projectFormError = 'Projekt-Titel ist Pflicht.'
      return
    }
    projectFormSaving = true
    const optimistic = {
      id: `local-${Date.now()}`,
      ...form,
      favorite: false,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    }
    projects.update(($p) => [optimistic, ...$p])
    selectedProjectId.set(optimistic.id)
    await setCache('projects', get(projects))
    try {
      if (!get(online)) {
        await enqueue({ method: 'POST', path: API.projects, body: form })
      } else {
        const created = await api(API.projects, { method: 'POST', body: form })
        const createdProject = created?.project
        if (createdProject?.id && columnTitles.length > 0) {
          for (const title of columnTitles) {
            await api(API.projectColumns(createdProject.id), { method: 'POST', body: { title } })
          }
        }
        if (createdProject?.id) {
          selectedProjectId.set(createdProject.id)
          await setCache('selectedProjectId', createdProject.id)
        }
        dispatch('reload-projects')
      }
      showProjectForm = false
    } catch (err) {
      projectFormError = err.message
      dispatch('error', err.message)
    } finally {
      projectFormSaving = false
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
  }  async function deleteProject(e, project) {
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
    <div class="brand-logo-wrap">
      <img class="brand-logo" src={$theme === 'light' ? '/logo-white.png' : '/logo-dark.png'} alt="Machino" />
    </div>
    <span class:offline={!$online} class="status-badge">{$online ? 'Online' : 'Offline'}</span>
  </div>

  <div class="sidebar-title">
    <h2>Projekte</h2>
    <button class="icon-btn" aria-label="Projekt erstellen" on:click={showProjectForm ? closeProjectForm : openProjectForm}>
      {showProjectForm ? '×' : '+'}
    </button>
  </div>

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
            <button class="project-select" on:click={() => selectProject(project.id)}>
              <span class="dot"></span>
              <span class="proj-title-wrap">
                <span>{project.title}</span>
                {#if project.memberCount > 1}
                  <span class="proj-shared-dot" title="{project.isOwner ? 'Geteilt' : 'Geteilt mit dir'}">
                    <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M23 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/></svg>
                  </span>
                {/if}
                {#if project.moveDone === false}
                  <span class="proj-packlist-dot" title="Packlisten-Modus">
                    <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M16 4h2a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2h2"/><rect x="8" y="2" width="8" height="4" rx="1"/><path d="m9 14 2 2 4-4"/></svg>
                  </span>
                {/if}
              </span>
            </button>
            <div class="project-actions">
              <button class="star active" aria-label="Favorit entfernen" on:click={() => setFavorite(project)}>★</button>
              <div class="proj-menu-wrap">
                <button class="proj-dots" aria-label="Optionen" on:click={(e) => openProjectMenu(e, project.id)}>⋯</button>
              </div>
            </div>
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
          <button class="project-select" on:click={() => selectProject(project.id)}>
            <span class="dot"></span>
            <span class="proj-title-wrap">
              <span>{project.title}</span>
              {#if project.memberCount > 1}
                <span class="proj-shared-dot" title="{project.isOwner ? 'Geteilt' : 'Geteilt mit dir'}">
                  <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M23 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/></svg>
                </span>
              {/if}
              {#if project.moveDone === false}
                <span class="proj-packlist-dot" title="Packlisten-Modus">
                  <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M16 4h2a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2h2"/><rect x="8" y="2" width="8" height="4" rx="1"/><path d="m9 14 2 2 4-4"/></svg>
                </span>
              {/if}
            </span>
          </button>
          <div class="project-actions">
            <button class="star" aria-label="Als Favorit markieren" on:click={() => setFavorite(project)}>☆</button>
            <div class="proj-menu-wrap">
              <button class="proj-dots" aria-label="Optionen" on:click={(e) => openProjectMenu(e, project.id)}>⋯</button>
            </div>
          </div>
        </div>
      {/each}
    </section>
  </nav>
</aside>

{#if showProjectForm}
  <CreateProjectModal
    defaultColor={DEFAULT_PROJECT_COLOR}
    saving={projectFormSaving}
    error={projectFormError}
    on:save={createProject}
    on:close={closeProjectForm}
  />
{/if}

<!-- Portal dropdown is rendered in App.svelte -->

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
    background: var(--sidebar-bg);
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

  .brand { padding: 0 4px 2px; }
  .sidebar-title { padding: 2px 10px 2px 4px; }

  .brand-logo-wrap {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    min-width: 0;
  }

  .brand-logo {
    width: 128px;
    max-width: 100%;
    height: auto;
    display: block;
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

  .project-select {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 9px 6px 9px 10px;
    background: transparent;
    color: var(--text);
    text-align: left;
    font-weight: 600;
    font-size: 0.88rem;
    transition: color 0.15s;
    min-width: 0;
    flex: 1;
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

  .proj-title-wrap {
    display: flex;
    align-items: center;
    gap: 5px;
    min-width: 0;
    overflow: hidden;
  }

  .proj-title-wrap > span:first-child {
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .proj-shared-dot {
    flex-shrink: 0;
    display: inline-flex;
    align-items: center;
    color: var(--text-faint);
    opacity: 0.7;
  }

  .proj-packlist-dot {
    flex-shrink: 0;
    display: inline-flex;
    align-items: center;
    color: color-mix(in srgb, #f59e0b, transparent 10%);
    opacity: 0.72;
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
