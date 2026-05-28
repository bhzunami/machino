<script>
  import { createEventDispatcher } from 'svelte'

  export let todo
  export let expanded = false
  export let detailForm = { description: '', dueDate: '', priority: 'normal' }
  export let dragOverId = ''
  export let dropAfter = false

  const dispatch = createEventDispatcher()

  function handleToggleExpand() {
    dispatch('expand', todo)
  }

  function handleSaveDetail() {
    dispatch('save-detail', { todoId: todo.id, form: { ...detailForm } })
  }
</script>

<article
  class="todo card"
  class:drop-before={dragOverId === todo.id && !dropAfter}
  class:drop-after={dragOverId === todo.id && dropAfter}
  draggable="true"
  on:dragstart={() => dispatch('dragstart', todo.id)}
  on:dragover|preventDefault={(e) => dispatch('dragover', { e, todoId: todo.id })}
  on:dragleave={() => dispatch('dragleave')}
  on:dragend={() => dispatch('dragend')}
  on:drop={() => dispatch('drop', todo.id)}
>
  <button class="check" aria-label="Todo erledigen" on:click|stopPropagation={() => dispatch('toggle', todo)}></button>
  <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
  <div class="todo-content" on:click={handleToggleExpand}>
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
    {#if expanded}
      <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
      <div class="todo-detail" on:click|stopPropagation>
        <label>
          Beschreibung
          <textarea
            rows="2"
            placeholder="Beschreibung hinzufügen..."
            bind:value={detailForm.description}
            on:blur={handleSaveDetail}
          ></textarea>
        </label>
        <div class="detail-row">
          <label>
            Fällig bis
            <div class="date-picker-wrap">
              <input
                type="date"
                bind:value={detailForm.dueDate}
                on:change={handleSaveDetail}
              />
              {#if detailForm.dueDate}
                <button class="date-clear" type="button" aria-label="Datum löschen"
                  on:click={() => { detailForm.dueDate = ''; handleSaveDetail(); }}>×</button>
              {/if}
            </div>
          </label>
          <label>
            Priorität
            <select bind:value={detailForm.priority} on:change={handleSaveDetail}>
              <option value="low">Niedrig</option>
              <option value="normal">Normal</option>
              <option value="high">Hoch</option>
            </select>
          </label>
        </div>
      </div>
    {/if}
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
    transition:
      transform 0.16s ease,
      border-color 0.16s ease,
      box-shadow 0.16s ease;
  }

  .todo[draggable='true'] {
    cursor: grab;
  }

  .todo[draggable='true']:hover {
    transform: translateY(-1px);
    border-color: color-mix(in srgb, var(--accent-color), white 58%);
    box-shadow: 0 18px 40px color-mix(in srgb, var(--accent-color), transparent 86%);
  }

  .todo.drop-before,
  .todo.drop-after {
    border-color: var(--accent-color);
    background: color-mix(in srgb, var(--accent-color), white 92%);
  }

  .todo.drop-before::before,
  .todo.drop-after::after {
    content: 'Hier einfügen';
    position: absolute;
    left: 18px;
    right: 18px;
    display: grid;
    place-items: center;
    height: 24px;
    border-radius: 999px;
    background: var(--accent-color);
    color: #fff;
    font-size: 0.72rem;
    font-weight: 900;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    box-shadow: 0 10px 24px color-mix(in srgb, var(--accent-color), transparent 62%);
    pointer-events: none;
  }

  .todo.drop-before::before {
    top: -18px;
  }

  .todo.drop-after::after {
    bottom: -18px;
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

  .todo-detail {
    display: grid;
    gap: 10px;
    padding: 12px 14px;
    border-radius: 12px;
    background: color-mix(in srgb, var(--accent-color), white 94%);
    border: 1.5px solid color-mix(in srgb, var(--accent-color), white 78%);
    animation: expand-in 0.16s ease;
  }

  @keyframes expand-in {
    from { opacity: 0; transform: translateY(-4px); }
    to   { opacity: 1; transform: translateY(0); }
  }

  .todo-detail label {
    display: flex;
    flex-direction: column;
    gap: 4px;
    font-size: 0.78rem;
    font-weight: 700;
    color: #4b5563;
  }

  .todo-detail textarea,
  .todo-detail input[type='date'],
  .todo-detail select {
    border-radius: 8px;
    border: 1.5px solid #e2e8f0;
    padding: 7px 10px;
    font-size: 0.85rem;
    background: #fff;
    color: #111827;
    resize: none;
    transition: border-color 0.15s;
    width: 100%;
    box-sizing: border-box;
  }

  .todo-detail textarea:focus,
  .todo-detail input[type='date']:focus,
  .todo-detail select:focus {
    outline: none;
    border-color: var(--accent-color);
    box-shadow: 0 0 0 3px color-mix(in srgb, var(--accent-color), transparent 82%);
  }

  .date-picker-wrap {
    position: relative;
    display: flex;
    align-items: center;
  }

  .date-picker-wrap input[type='date'] {
    padding-right: 34px;
    cursor: pointer;
  }

  .date-picker-wrap input[type='date']::-webkit-calendar-picker-indicator {
    opacity: 0.6;
    cursor: pointer;
    filter: invert(30%) sepia(80%) saturate(600%) hue-rotate(220deg);
  }

  .date-clear {
    position: absolute;
    right: 8px;
    width: 20px;
    height: 20px;
    border-radius: 999px;
    border: none;
    background: #e5e7eb;
    color: #374151;
    font-size: 0.85rem;
    font-weight: 900;
    line-height: 1;
    cursor: pointer;
    display: grid;
    place-items: center;
    padding: 0;
    transition: background 0.15s;
  }

  .date-clear:hover {
    background: #d1d5db;
  }

  .detail-row {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 12px;
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

  .drag-handle {
    color: #9ca3af;
    font-size: 1.2rem;
    font-weight: 900;
    letter-spacing: -0.2em;
    flex-shrink: 0;
  }

  @media (max-width: 900px) {
    .detail-row {
      grid-template-columns: 1fr;
    }
  }
</style>
