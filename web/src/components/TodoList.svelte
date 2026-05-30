<script>
  import { createEventDispatcher } from 'svelte'
  import { get } from 'svelte/store'
  import { setCache } from '../lib/db.js'
  import { API, TODO_CACHE_KEY } from '../lib/constants.js'
  import { todos, columns, selectedProjectId, selectedProject, activeTodos, completedTodos, projects } from '../lib/stores.js'
  import TodoItem from './TodoItem.svelte'
  import TodoEditModal from './TodoEditModal.svelte'

  const dispatch = createEventDispatcher()

  let draggedTodoId = ''
  let dragOverTodoId = ''
  let dropAfter = false
  let editingTodo = null
  let completedCollapsed = true

  // Board: new todo input per column
  let newTodoByColumn = {}

  // Board: column being dragged over
  let dragOverColumnId = ''
  let moveDone = true
  let visibleTodos = []

  $: moveDone = $selectedProject?.moveDone !== false
  $: visibleTodos = moveDone ? $activeTodos : orderForDisplay($todos)

  function orderForDisplay(list) {
    return [...list].sort((a, b) => {
      if (a.completed !== b.completed) return a.completed ? 1 : -1
      return (a.position || 0) - (b.position || 0)
    })
  }

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

  function openEditTodo(todo) {
    editingTodo = todo
  }

  async function saveTodoDetail({ todoId, form }) {
    const $selectedProjectId = get(selectedProjectId)
    const title = form.title?.trim()
    if (!title) return
    const body = {
      title,
      description: form.description || null,
      dueDate: form.dueDate || null,
      priority: form.priority,
    }
    todos.update(($todos) =>
      $todos.map((t) =>
        t.id === todoId
          ? { ...t, title: body.title, description: body.description, dueDate: body.dueDate, priority: body.priority }
          : t,
      ),
    )
    await setCache(TODO_CACHE_KEY($selectedProjectId), get(todos))
    editingTodo = null
    dispatch('run-or-queue', { method: 'PATCH', path: API.todo(todoId), body })
  }

  async function reorderActiveTodo(targetTodoId, after) {
    if (!draggedTodoId || draggedTodoId === targetTodoId) {
      clearDragState()
      return
    }
    const nextActive = moveDone ? [...get(activeTodos)] : orderForDisplay(get(todos))
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
    const tailTodos = moveDone ? get(completedTodos) : []
    const reordered = [...nextActive, ...tailTodos].map((todo, position) => ({
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
    dragOverColumnId = ''
  }

  // Board: move todo to a different column via drag & drop
  async function dropTodoOnColumn(colId) {
    if (!draggedTodoId) { clearDragState(); return }
    const todoId = draggedTodoId
    const todo = get(todos).find(t => t.id === todoId)
    const targetColumnId = colId === '__none__' ? null : colId
    if (!todo || (todo.columnId ?? null) === targetColumnId) { clearDragState(); return }
    todos.update($todos => $todos.map(t => t.id === todoId ? { ...t, columnId: targetColumnId } : t))
    const $selectedProjectId = get(selectedProjectId)
    await setCache(TODO_CACHE_KEY($selectedProjectId), get(todos))
    const body = targetColumnId ? { columnId: targetColumnId } : { clearColumn: true }
    dispatch('run-or-queue', { method: 'PATCH', path: API.todo(todoId), body })
    clearDragState()
  }

  // Board: add new todo to a specific column
  async function addTodoToColumn(colId) {
    const title = (newTodoByColumn[colId] || '').trim()
    if (!title) return
    const $selectedProjectId = get(selectedProjectId)
    const columnId = colId === '__none__' ? undefined : colId
    newTodoByColumn[colId] = ''
    dispatch('run-or-queue', {
      method: 'POST',
      path: API.projectTodos($selectedProjectId),
      body: { title, columnId: columnId || null },
    })
    // Optimistic: add to local todos immediately
    const tempId = 'temp-' + Date.now()
    todos.update($todos => [
      { id: tempId, projectId: $selectedProjectId, columnId: columnId || null, title, description: '', priority: 'normal', completed: false, position: 0, createdAt: new Date().toISOString(), updatedAt: new Date().toISOString() },
      ...$todos,
    ])
  }

</script>

{#if $columns.length > 0}
  <!-- ── Board view ───────────────────────────────────────────── -->
  <div class="board">
    {#each [...$columns, { id: '__none__', title: 'Nicht zugewiesen' }] as col (col.id)}
      {@const colTodos = visibleTodos.filter(t => (t.columnId ?? null) === (col.id === '__none__' ? null : col.id))}
      <div
        class="board-col"
        class:drag-over={dragOverColumnId === col.id}
        role="list"
        on:dragover|preventDefault={() => (dragOverColumnId = col.id)}
        on:dragleave={() => (dragOverColumnId = '')}
        on:drop|preventDefault={() => dropTodoOnColumn(col.id)}
      >
        <div class="board-col-header">
          <span class="board-col-title">{col.title}</span>
          <span class="board-col-count">{colTodos.length}</span>
        </div>
        <div class="board-col-body">
          {#each colTodos as todo (todo.id)}
            <TodoItem
              {todo}
              dragOverId={dragOverTodoId}
              {dropAfter}
              on:toggle={(e) => toggleTodo(e.detail)}
              on:edit={(e) => openEditTodo(e.detail)}
              on:dragstart={(e) => (draggedTodoId = e.detail)}
              on:dragover={(e) => { dragOverColumnId = col.id; markDropTarget(e.detail.e, e.detail.todoId) }}
              on:dragleave={() => (dragOverTodoId = '')}
              on:dragend={clearDragState}
              on:drop={() => dropTodoOnColumn(col.id)}
            />
          {/each}
          {#if colTodos.length === 0}
            <div class="board-col-empty">Kein Todo</div>
          {/if}
        </div>
        <form class="board-add-form" on:submit|preventDefault={() => addTodoToColumn(col.id)}>
          <input
            class="board-add-input"
            type="text"
            placeholder="+ Todo hinzufügen"
            bind:value={newTodoByColumn[col.id]}
          />
        </form>
      </div>
    {/each}
  </div>

{:else}
  <!-- ── List view (default) ───────────────────────────────────── -->
  <section class="todo-list" aria-label="Offene Todos">
    {#each visibleTodos as todo (todo.id)}
      <TodoItem
        {todo}
        dragOverId={dragOverTodoId}
        {dropAfter}
        on:toggle={(e) => toggleTodo(e.detail)}
        on:edit={(e) => openEditTodo(e.detail)}
        on:dragstart={(e) => (draggedTodoId = e.detail)}
        on:dragover={(e) => markDropTarget(e.detail.e, e.detail.todoId)}
        on:dragleave={() => (dragOverTodoId = '')}
        on:dragend={clearDragState}
        on:drop={(e) => reorderActiveTodo(e.detail, dropAfter)}
      />
    {:else}
      {#if $projects.length === 0}
        <div class="empty card">
          <span class="empty-icon">
            <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" style="opacity:0.35"><rect x="3" y="5" width="18" height="14" rx="3"/><path d="M3 10h18"/><path d="M8 3v4"/><path d="M16 3v4"/></svg>
          </span>
          <strong>Noch kein Projekt</strong>
          <button class="btn" on:click={() => dispatch('new-project')}>＋ Neues Projekt</button>
        </div>
      {:else}
        <div class="empty card">
          <span class="empty-icon">
            <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" style="opacity:0.35"><rect x="3" y="5" width="18" height="14" rx="3"/><path d="M3 10h18"/><path d="M8 3v4"/><path d="M16 3v4"/></svg>
          </span>
          <strong>Keine offenen Todos</strong>
          <span>Oben eingeben und Enter drücken, um eine Aufgabe zu erstellen.</span>
        </div>
      {/if}
    {/each}
  </section>
{/if}

{#if moveDone && $completedTodos.length}
  <section class="completed-section">
    <div class="completed-header">
      <button class="collapse-btn" on:click={() => (completedCollapsed = !completedCollapsed)}>
        <span class="chevron" class:rotated={!completedCollapsed}>›</span>
        <h2>Erledigt <span class="count">({$completedTodos.length})</span></h2>
      </button>
      {#if !completedCollapsed}
        <button class="btn-clear-done" on:click={deleteCompletedTodos} title="Alle erledigten Todos löschen">
          🗑 Alle löschen
        </button>
      {/if}
    </div>
    {#if !completedCollapsed}
      <div class="todo-list">
        {#each $completedTodos as todo (todo.id)}
          <article class="todo done card">
            <button class="check" on:click|stopPropagation={() => toggleTodo(todo)}>✓</button>
            <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
            <div on:click={() => openEditTodo(todo)}>
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
    {/if}
  </section>
{/if}

{#if editingTodo}
  <TodoEditModal
    todo={editingTodo}
    on:save={(e) => saveTodoDetail(e.detail)}
    on:close={() => (editingTodo = null)}
  />
{/if}

<style>
  /* ── Board layout ──────────────────────────────────────────── */
  .board {
    display: flex;
    gap: 14px;
    overflow-x: auto;
    align-items: flex-start;
    padding-bottom: 12px;
    min-height: 0;
  }

  .board-col {
    flex: 0 0 280px;
    min-width: 280px;
    background: var(--card-bg);
    border: 1px solid var(--border);
    border-radius: 16px;
    display: flex;
    flex-direction: column;
    gap: 0;
    transition: box-shadow 0.15s, border-color 0.15s;
  }

  .board-col.drag-over {
    border-color: var(--accent-color);
    box-shadow: 0 0 0 2px color-mix(in srgb, var(--accent-color), transparent 70%);
  }

  .board-col-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 12px 14px 8px;
    border-bottom: 1px solid var(--border);
  }

  .board-col-title {
    font-size: 0.78rem;
    font-weight: 800;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    color: var(--text-muted);
  }

  .board-col-count {
    font-size: 0.72rem;
    font-weight: 700;
    color: var(--text-faint);
    background: var(--surface-2, rgba(128,128,128,0.1));
    border-radius: 10px;
    padding: 1px 7px;
  }

  .board-col-body {
    display: flex;
    flex-direction: column;
    gap: 6px;
    padding: 10px 10px 4px;
    min-height: 60px;
  }

  .board-col-empty {
    color: var(--text-faint);
    font-size: 0.8rem;
    text-align: center;
    padding: 16px 8px;
    border: 1px dashed var(--border);
    border-radius: 10px;
  }

  .board-add-form {
    padding: 6px 10px 10px;
  }

  .board-add-input {
    width: 100%;
    box-sizing: border-box;
    background: transparent;
    border: 1px solid var(--border);
    border-radius: 8px;
    padding: 7px 10px;
    font-size: 0.85rem;
    color: var(--text);
    outline: none;
    transition: border-color 0.12s;
  }
  .board-add-input:focus {
    border-color: var(--accent-color);
  }
  .board-add-input::placeholder {
    color: var(--text-faint);
  }

  /* ── List layout ───────────────────────────────────────────── */
  .todo-list {
    display: grid;
    gap: 6px;
  }

  .completed-section {
    display: grid;
    gap: 6px;
    margin-top: 16px;
  }

  .completed-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0 4px 6px;
    border-bottom: 1px solid var(--border);
    margin-bottom: 4px;
  }

  .collapse-btn {
    display: flex;
    align-items: center;
    gap: 8px;
    background: transparent;
    border: none;
    padding: 0;
    cursor: pointer;
    color: inherit;
  }

  .chevron {
    font-size: 1.1rem;
    color: var(--text-faint);
    transition: transform 0.2s cubic-bezier(0.16,1,0.3,1);
    display: inline-block;
    transform: rotate(0deg);
    line-height: 1;
  }
  .chevron.rotated { transform: rotate(90deg); }

  .completed-header h2 {
    margin: 0;
    color: var(--text-faint);
    font-size: 0.72rem;
    font-weight: 800;
    letter-spacing: 0.1em;
    text-transform: uppercase;
  }

  .count {
    color: var(--text-faint);
    font-weight: 600;
  }

  .btn-clear-done {
    padding: 4px 12px;
    border-radius: 8px;
    background: transparent;
    border: 1px solid rgba(248,113,113,0.25);
    color: #f87171;
    font-size: 0.76rem;
    font-weight: 700;
    cursor: pointer;
    transition: background 0.12s, border-color 0.12s;
  }
  .btn-clear-done:hover {
    background: rgba(248,113,113,0.1);
    border-color: rgba(248,113,113,0.45);
  }

  .empty {
    padding: 52px 24px;
    text-align: center;
    color: var(--text-faint);
    font-size: 0.88rem;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 10px;
    border: 1px dashed var(--border);
    background: transparent;
    box-shadow: none;
  }

  .empty strong {
    color: var(--text-muted);
    font-size: 0.95rem;
  }

  .empty-icon {
    color: var(--text-faint);
    margin-bottom: 4px;
  }

  /* Done-state overrides for inline completed todos */
  .todo.done {
    opacity: 0.62;
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 10px 14px;
    border-radius: 14px;
    cursor: pointer;
  }
  .todo.done .check {
    flex-shrink: 0;
    width: 22px;
    height: 22px;
    border-radius: 6px;
    border: 2px solid var(--accent-color);
    background: color-mix(in srgb, var(--accent-color), transparent 55%);
    color: #fff;
    font-size: 0.7rem;
    display: grid;
    place-items: center;
    cursor: pointer;
    transition: transform 0.15s;
  }
  .todo.done .check:hover { transform: scale(1.1); }
  .todo.done > div { flex: 1; min-width: 0; }
  .todo.done .todo-row {
    display: flex;
    align-items: center;
    gap: 8px;
    flex-wrap: wrap;
  }
  .todo.done .todo-text {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-size: 0.92rem;
  }
  .todo.done .todo-title {
    text-decoration: line-through;
    color: var(--text-faint);
    font-size: 0.92rem;
  }
  .todo.done .todo-chips {
    display: flex;
    gap: 6px;
    flex-shrink: 0;
  }
</style>
