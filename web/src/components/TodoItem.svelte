<script>
  import { createEventDispatcher } from 'svelte'

  export let todo
  export let dragOverId = ''
  export let dropAfter = false

  const dispatch = createEventDispatcher()

  function handleEdit() {
    dispatch('edit', todo)
  }

  function handleDragStart(e) {
    if (e.dataTransfer) {
      e.dataTransfer.effectAllowed = 'move'
      e.dataTransfer.setData('text/plain', todo.id)
    }
    dispatch('dragstart', todo.id)
  }
</script>

<article
  class="todo card"
  class:completed={todo.completed}
  class:drop-before={dragOverId === todo.id && !dropAfter}
  class:drop-after={dragOverId === todo.id && dropAfter}
  draggable="true"
  on:dragstart={handleDragStart}
  on:dragover|preventDefault={(e) => dispatch('dragover', { e, todoId: todo.id })}
  on:dragleave={() => dispatch('dragleave')}
  on:dragend={() => dispatch('dragend')}
  on:drop|preventDefault|stopPropagation={() => dispatch('drop', todo.id)}
>
  <button class="check" aria-label={todo.completed ? 'Todo wieder öffnen' : 'Todo erledigen'} on:click|stopPropagation={() => dispatch('toggle', todo)}><span class="check-inner">✓</span></button>
  <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
  <div class="todo-content" on:click={handleEdit}>
    <div class="todo-row">
      <div class="todo-text">
        <span class="todo-title">{todo.title}</span>
        {#if todo.description}
          <span class="todo-desc"> · {todo.description}</span>
        {/if}
      </div>
      <div class="todo-meta">
        {#if todo.dueDate}
          <span class="meta-chip date-chip">📅 {String(todo.dueDate).slice(0, 10)}</span>
        {/if}
        {#if todo.priority && todo.priority !== 'normal'}
          <span class="prio-dot prio-{todo.priority}" title={todo.priority === 'high' ? 'Hoch' : 'Niedrig'}></span>
        {/if}
        <span class="edit-icon" aria-hidden="true">Bearbeiten</span>
      </div>
    </div>
  </div>
  <span class="drag-handle" aria-hidden="true">⋮⋮</span>
</article>

<style>
  .todo {
    position: relative;
    display: grid;
    grid-template-columns: auto minmax(0, 1fr) auto;
    align-items: center;
    gap: 12px;
    padding: 11px 14px;
    background: var(--glass);
    border-color: var(--border);
    transition:
      transform 0.18s cubic-bezier(0.16,1,0.3,1),
      border-color 0.18s,
      box-shadow 0.18s,
      background 0.18s;
    animation: todo-in 0.22s cubic-bezier(0.16,1,0.3,1) both;
  }

  @keyframes todo-in {
    from { opacity: 0; transform: translateY(-6px) scale(0.98); }
    to   { opacity: 1; transform: translateY(0) scale(1); }
  }

  .todo[draggable='true'] { cursor: grab; }

  .todo[draggable='true']:hover {
    transform: translateY(-2px);
    border-color: color-mix(in srgb, var(--accent-color), transparent 60%);
    box-shadow: 0 12px 40px rgba(0,0,0,0.4), 0 0 0 1px color-mix(in srgb, var(--accent-color), transparent 72%);
    background: var(--glass-md);
  }

  .todo.drop-before,
  .todo.drop-after {
    border-color: color-mix(in srgb, var(--accent-color), transparent 35%);
    background: color-mix(in srgb, var(--accent-color), transparent 90%);
  }

  .todo.drop-before::before,
  .todo.drop-after::after {
    content: '↓ Hier einfügen';
    position: absolute;
    left: 14px;
    right: 14px;
    display: grid;
    place-items: center;
    height: 22px;
    border-radius: 999px;
    background: linear-gradient(135deg, var(--accent-color), color-mix(in srgb, var(--accent-color), #7c3aed 50%));
    color: #fff;
    font-size: 0.68rem;
    font-weight: 800;
    letter-spacing: 0.09em;
    text-transform: uppercase;
    box-shadow: 0 0 24px color-mix(in srgb, var(--accent-color), transparent 45%);
    pointer-events: none;
    z-index: 2;
  }

  .todo.drop-before::before { top: -17px; }
  .todo.drop-after::after   { bottom: -17px; }

  .todo.completed {
    opacity: 0.72;
  }

  .todo.completed .check {
    background: color-mix(in srgb, var(--accent-color), transparent 45%);
    border-color: var(--accent-color);
    color: #fff;
  }

  .todo.completed .todo-title {
    text-decoration: line-through;
    color: var(--text-faint);
  }

  .todo.completed .todo-desc {
    color: var(--text-faint);
  }

  .todo-content {
    cursor: pointer;
    display: flex;
    flex-direction: column;
    gap: 10px;
    min-width: 0;
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
    display: flex;
    align-items: baseline;
    gap: 0;
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis;
    font-size: 0.92rem;
  }

  .todo-title {
    font-weight: 600;
    color: var(--text);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    flex-shrink: 1;
    min-width: 0;
  }

  .todo-desc {
    color: var(--text-muted);
    font-weight: 400;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    flex-shrink: 2;
  }

  .todo-meta {
    display: flex;
    align-items: center;
    gap: 5px;
    flex-shrink: 0;
  }

  .edit-icon {
    font-size: 0.68rem;
    font-weight: 800;
    letter-spacing: 0.06em;
    text-transform: uppercase;
    color: var(--text-faint);
    opacity: 0;
    transition: opacity 0.15s, color 0.15s;
  }
  .todo:hover .edit-icon { opacity: 1; color: var(--text-muted); }

  .meta-chip {
    border-radius: 999px;
    padding: 2px 9px;
    font-size: 0.72rem;
    font-weight: 700;
    white-space: nowrap;
  }

  .meta-chip.date-chip {
    background: rgba(56,189,248,0.1);
    border: 1px solid rgba(56,189,248,0.2);
    color: #7dd3fc;
  }

  .meta-chip.prio-high {
    background: rgba(248,113,113,0.1);
    border: 1px solid rgba(248,113,113,0.2);
    color: #fca5a5;
  }

  .meta-chip.prio-low {
    background: rgba(99,102,241,0.1);
    border: 1px solid rgba(99,102,241,0.2);
    color: #a5b4fc;
  }

  /* Priority dot — compact colored indicator */
  .prio-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    flex-shrink: 0;
    display: inline-block;
  }
  .prio-dot.prio-high { background: #f87171; box-shadow: 0 0 4px rgba(248,113,113,0.5); }
  .prio-dot.prio-low  { background: #818cf8; box-shadow: 0 0 4px rgba(129,140,248,0.5); }

  /* Light mode chip overrides */
  :global([data-theme="light"]) .meta-chip.date-chip {
    background: rgba(14,165,233,0.1);
    border-color: rgba(14,165,233,0.3);
    color: #0369a1;
  }
  :global([data-theme="light"]) .meta-chip.prio-high {
    background: rgba(220,38,38,0.08);
    border-color: rgba(220,38,38,0.25);
    color: #b91c1c;
  }
  :global([data-theme="light"]) .meta-chip.prio-low {
    background: rgba(99,102,241,0.1);
    border-color: rgba(99,102,241,0.3);
    color: #4338ca;
  }

  /* Checkbox */
  .check {
    width: 22px;
    height: 22px;
    border: 1.5px solid var(--border-hover);
    border-radius: 6px;
    background: transparent;
    color: transparent;
    font-weight: 900;
    flex-shrink: 0;
    transition: background 0.15s, border-color 0.15s, box-shadow 0.15s, transform 0.12s, color 0.15s;
    display: grid;
    place-items: center;
  }
  .check:hover {
    background: color-mix(in srgb, var(--accent-color), transparent 82%);
    border-color: var(--accent-color);
    box-shadow: 0 0 10px color-mix(in srgb, var(--accent-color), transparent 60%);
    transform: scale(1.08);
    color: var(--accent-color);
  }
  .check-inner {
    font-size: 0.7rem;
    line-height: 1;
    pointer-events: none;
  }

  .drag-handle {
    color: var(--text-faint);
    font-size: 1rem;
    font-weight: 900;
    letter-spacing: -0.15em;
    flex-shrink: 0;
    opacity: 0;
    transition: opacity 0.15s;
  }
  .todo:hover .drag-handle { opacity: 1; }

</style>
