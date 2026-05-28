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
    placeholder="Neues Todo eingeben und Enter drücken..."
    aria-label="Neues Todo"
    disabled={!$selectedProjectId}
  />
  <button class="btn" type="submit" disabled={!$selectedProjectId}>Hinzufügen</button>
</form>

<style>
  .quick-add {
    display: grid;
    grid-template-columns: minmax(0, 1fr) auto;
    gap: 12px;
    padding: 10px 14px;
    border-color: rgba(226, 232, 240, 0.9);
    background: #fff;
    box-shadow:
      inset 0 0 0 1px color-mix(in srgb, var(--accent-color), transparent 92%),
      0 8px 24px rgba(15, 23, 42, 0.06);
  }

  .quick-add :global(input:focus) {
    border-color: var(--accent-color);
    outline: 3px solid color-mix(in srgb, var(--accent-color), transparent 82%);
  }

  @media (max-width: 900px) {
    .quick-add {
      grid-template-columns: 1fr;
    }
  }
</style>
