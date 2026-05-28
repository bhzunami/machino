<script>
  import { createEventDispatcher } from 'svelte'
  import { get } from 'svelte/store'
  import { setCache } from '../lib/db.js'
  import { API, TODO_CACHE_KEY } from '../lib/constants.js'
  import { todos, selectedProjectId, activeTodos, completedTodos } from '../lib/stores.js'
  import TodoItem from './TodoItem.svelte'

  const dispatch = createEventDispatcher()

  let draggedTodoId = ''
  let dragOverTodoId = ''
  let dropAfter = false
  let expandedTodoId = ''
  let detailForm = { description: '', dueDate: '', priority: 'normal' }

  async function toggleTodo(todo) {
    const completed = !todo.completed
    const $selectedProjectId = get(selectedProjectId)
    todos.update(($todos) => $todos.map((item) => (item.id === todo.id ? { ...item, completed } : item)))
    await setCache(TODO_CACHE_KEY($selectedProjectId), get(todos))
    dispatch('run-or-queue', { method: 'PATCH', path: API.todo(todo.id), body: { completed } })
  }

  async function deleteCompletedTodos() {
    const $selectedProjectId = get(selectedProjectId)
    todos.update(($todos) => $todos.filter((t) => !t.completed))
    await setCache(TODO_CACHE_KEY($selectedProjectId), get(todos))
    dispatch('run-or-queue', {
      method: 'DELETE',
      path: API.projectTodosCompleted($selectedProjectId),
      body: null,
    })
  }

  function toggleExpand(todo) {
    if (expandedTodoId === todo.id) {
      expandedTodoId = ''
      return
    }
    expandedTodoId = todo.id
    detailForm = {
      description: todo.description || '',
      dueDate: todo.dueDate ? String(todo.dueDate).slice(0, 10) : '',
      priority: todo.priority || 'normal',
    }
  }

  async function saveTodoDetail({ todoId, form }) {
    const $selectedProjectId = get(selectedProjectId)
    const body = {
      description: form.description || null,
      dueDate: form.dueDate || null,
      priority: form.priority,
    }
    todos.update(($todos) =>
      $todos.map((t) =>
        t.id === todoId
          ? { ...t, description: body.description, dueDate: body.dueDate, priority: body.priority }
          : t,
      ),
    )
    await setCache(TODO_CACHE_KEY($selectedProjectId), get(todos))
    dispatch('run-or-queue', { method: 'PATCH', path: API.todo(todoId), body })
  }

  async function reorderActiveTodo(targetTodoId, after) {
    if (!draggedTodoId || draggedTodoId === targetTodoId) {
      clearDragState()
      return
    }
    const nextActive = [...get(activeTodos)]
    const source = nextActive.findIndex((todo) => todo.id === draggedTodoId)
    const target = nextActive.findIndex((todo) => todo.id === targetTodoId)
    if (source < 0 || target < 0) {
      clearDragState()
      return
    }
    const [item] = nextActive.splice(source, 1)
    let insertAt = target
    if (!after && source < target) insertAt = target - 1
    if (after && source > target) insertAt = target + 1
    if (after && source < target) insertAt = target
    nextActive.splice(insertAt, 0, item)
    const $completedTodos = get(completedTodos)
    const reordered = [...nextActive, ...$completedTodos].map((todo, position) => ({
      ...todo,
      position: position + 1,
    }))
    todos.set(reordered)
    clearDragState()
    const $selectedProjectId = get(selectedProjectId)
    await setCache(TODO_CACHE_KEY($selectedProjectId), get(todos))
    dispatch('run-or-queue', {
      method: 'PUT',
      path: API.projectTodosReorder($selectedProjectId),
      body: { ids: get(todos).map((todo) => todo.id) },
    })
  }

  function markDropTarget(e, todoId) {
    const rect = e.currentTarget.getBoundingClientRect()
    dragOverTodoId = todoId
    dropAfter = e.clientY > rect.top + rect.height / 2
  }

  function clearDragState() {
    draggedTodoId = ''
    dragOverTodoId = ''
    dropAfter = false
  }
</script>

<section class="todo-list" aria-label="Offene Todos">
  {#each $activeTodos as todo (todo.id)}
    <TodoItem
      {todo}
      expanded={expandedTodoId === todo.id}
      detailForm={expandedTodoId === todo.id ? detailForm : { description: todo.description || '', dueDate: todo.dueDate ? String(todo.dueDate).slice(0, 10) : '', priority: todo.priority || 'normal' }}
      dragOverId={dragOverTodoId}
      {dropAfter}
      on:toggle={(e) => toggleTodo(e.detail)}
      on:expand={(e) => toggleExpand(e.detail)}
      on:save-detail={(e) => saveTodoDetail(e.detail)}
      on:dragstart={(e) => (draggedTodoId = e.detail)}
      on:dragover={(e) => markDropTarget(e.detail.e, e.detail.todoId)}
      on:dragleave={() => (dragOverTodoId = '')}
      on:dragend={clearDragState}
      on:drop={(e) => reorderActiveTodo(e.detail, dropAfter)}
    />
  {:else}
    <div class="empty card">Noch keine offenen Todos. Neue Aufgaben erscheinen hier untereinander.</div>
  {/each}
</section>

{#if $completedTodos.length}
  <section class="completed-section">
    <div class="completed-header">
      <h2>Erledigte Todos</h2>
      <button class="btn-clear-done" on:click={deleteCompletedTodos} title="Alle erledigten Todos löschen">
        🗑 Alle löschen
      </button>
    </div>
    <div class="todo-list">
      {#each $completedTodos as todo (todo.id)}
        <article class="todo done card">
          <button class="check" on:click={() => toggleTodo(todo)}>✓</button>
          <div>
            <div class="todo-row">
              <div class="todo-text">
                <span class="todo-title">{todo.title}</span>
                {#if todo.description}
                  <span class="todo-desc"> · {todo.description}</span>
                {/if}
              </div>
              <div class="todo-chips">
                {#if todo.dueDate}
                  <span class="meta-chip date-chip">📅 {String(todo.dueDate).slice(0, 10)}</span>
                {/if}
                {#if todo.priority && todo.priority !== 'normal'}
                  <span class="meta-chip prio-{todo.priority}">{todo.priority === 'high' ? '🔴 Hoch' : '🔵 Niedrig'}</span>
                {/if}
              </div>
            </div>
          </div>
        </article>
      {/each}
    </div>
  </section>
{/if}

<style>
  .todo-list {
    display: grid;
    gap: 8px;
  }

  .todo {
    position: relative;
    display: grid;
    grid-template-columns: auto minmax(0, 1fr) auto;
    align-items: center;
    gap: 12px;
    padding: 11px 14px;
  }

  .todo-row {
    display: flex;
    align-items: center;
    gap: 10px;
    min-width: 0;
  }

  .todo-text {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-size: 0.95rem;
  }

  .todo-chips {
    display: flex;
    align-items: center;
    gap: 6px;
    flex-shrink: 0;
  }

  .todo-title {
    font-weight: 700;
    color: #111827;
  }

  .todo-desc {
    color: #6b7280;
    font-weight: 400;
  }

  .meta-chip {
    border-radius: 999px;
    padding: 2px 9px;
    font-size: 0.74rem;
    font-weight: 700;
    white-space: nowrap;
  }

  .meta-chip.date-chip {
    background: color-mix(in srgb, #06b6d4, white 88%);
    color: #0e7490;
  }

  .meta-chip.prio-high {
    background: color-mix(in srgb, #ef4444, white 86%);
    color: #b91c1c;
  }

  .meta-chip.prio-low {
    background: color-mix(in srgb, #3b82f6, white 86%);
    color: #1d4ed8;
  }

  .todo.done {
    opacity: 0.65;
  }

  .todo.done .todo-title {
    text-decoration: line-through;
    color: #6b7280;
  }

  .check {
    width: 28px;
    height: 28px;
    border: 2px solid var(--accent-color);
    border-radius: 999px;
    background: color-mix(in srgb, var(--accent-color), white 92%);
    color: var(--accent-color);
    font-weight: 900;
    flex-shrink: 0;
  }

  .completed-section {
    display: grid;
    gap: 12px;
    margin-top: 10px;
  }

  .completed-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .completed-section h2 {
    margin: 0;
    color: #4b5563;
    font-size: 1rem;
  }

  .btn-clear-done {
    padding: 5px 12px;
    border-radius: 8px;
    background: transparent;
    border: 1px solid #fca5a5;
    color: #dc2626;
    font-size: 0.8rem;
    font-weight: 600;
    cursor: pointer;
    transition: background 0.12s, color 0.12s;
  }

  .btn-clear-done:hover {
    background: #fef2f2;
    border-color: #dc2626;
  }

  .empty {
    padding: 30px;
    text-align: center;
    color: #6b7280;
  }
</style>
