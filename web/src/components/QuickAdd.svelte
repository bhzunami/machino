<script>
  import { createEventDispatcher } from 'svelte'
  import { get } from 'svelte/store'
  import { api } from '../lib/api.js'
  import { enqueue, setCache } from '../lib/db.js'
  import { API, TODO_CACHE_KEY, PRIORITY } from '../lib/constants.js'
  import { todos, selectedProjectId, online, activeTodos, completedTodos } from '../lib/stores.js'

  const dispatch = createEventDispatcher()

  let todoForm = { title: '', description: '', dueDate: '', priority: PRIORITY.NORMAL }

  async function createTodo() {
    const $selectedProjectId = get(selectedProjectId)
    if (!$selectedProjectId) {
      dispatch('error', 'Bitte zuerst ein Projekt auswählen.')
      return
    }
    if (!todoForm.title.trim()) {
      dispatch('error', 'Todo-Titel ist Pflicht.')
      return
    }
    const $activeTodos = get(activeTodos)
    const $completedTodos = get(completedTodos)
    const $todos = get(todos)
    const optimistic = {
      id: `local-${Date.now()}`,
      projectId: $selectedProjectId,
      ...todoForm,
      completed: false,
      position: $todos.length + 1,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    }
    todos.set(
      [optimistic, ...$activeTodos, ...$completedTodos].map((todo, position) => ({
        ...todo,
        position: position + 1,
      })),
    )
    await setCache(TODO_CACHE_KEY($selectedProjectId), get(todos))
    const body = { ...todoForm, dueDate: todoForm.dueDate || null }
    dispatch('run-or-queue', {
      method: 'POST',
      path: API.projectTodos($selectedProjectId),
      body,
    })
    todoForm = { title: '', description: '', dueDate: '', priority: PRIORITY.NORMAL }
  }
</script>

<form class="quick-add card" on:submit|preventDefault={createTodo}>
  <input
    bind:value={todoForm.title}
    placeholder="Neue Aufgabe…"
    aria-label="Neues Todo"
    disabled={!$selectedProjectId}
  />
  <button class="btn enter-btn" type="submit" disabled={!$selectedProjectId} title="Hinzufügen (Enter)">
    <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="9 10 4 15 9 20"/><path d="M20 4v7a4 4 0 0 1-4 4H4"/></svg>
  </button>
</form>

<style>
  .quick-add {
    display: grid;
    grid-template-columns: minmax(0, 1fr) auto;
    gap: 10px;
    padding: 10px 10px 10px 16px;
    background: var(--glass);
    border-color: var(--border);
    backdrop-filter: blur(24px);
    transition: border-color 0.2s, box-shadow 0.2s;
  }

  .quick-add:focus-within {
    border-color: color-mix(in srgb, var(--accent-color), transparent 52%);
    box-shadow: 0 0 0 1px color-mix(in srgb, var(--accent-color), transparent 72%),
                0 8px 32px rgba(0,0,0,0.3);
  }

  .quick-add input {
    border: none;
    background: transparent;
    padding: 6px 4px;
    font-size: 0.95rem;
    color: var(--text);
    border-radius: 0;
    box-shadow: none !important;
  }

  .quick-add input::placeholder { color: var(--text-faint); }

  .quick-add .enter-btn {
    padding: 8px 14px;
    display: grid;
    place-items: center;
    min-width: 42px;
  }

  @media (max-width: 900px) {
    .quick-add { grid-template-columns: 1fr; }
  }
</style>
