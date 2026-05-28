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
    font-size: 0.92rem;
  }

  .todo-title {
    font-weight: 600;
    color: var(--text);
  }

  .todo-desc {
    color: var(--text-muted);
    font-weight: 400;
  }

  .todo-chips {
    display: flex;
    align-items: center;
    gap: 5px;
    flex-shrink: 0;
  }

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

  /* Expanded detail */
  .todo-detail {
    display: grid;
    gap: 10px;
    padding: 14px;
    border-radius: 14px;
    background: rgba(0,0,0,0.25);
    border: 1px solid var(--border);
    animation: expand-in 0.18s cubic-bezier(0.16,1,0.3,1);
    margin-top: 2px;
  }

  @keyframes expand-in {
    from { opacity: 0; transform: translateY(-4px); max-height: 0; }
    to   { opacity: 1; transform: translateY(0); max-height: 300px; }
  }

  .todo-detail label {
    display: flex;
    flex-direction: column;
    gap: 5px;
    font-size: 0.74rem;
    font-weight: 700;
    color: var(--text-muted);
    text-transform: uppercase;
    letter-spacing: 0.06em;
  }

  .todo-detail textarea,
  .todo-detail input[type='date'],
  .todo-detail select {
    border-radius: 10px;
    padding: 8px 12px;
    font-size: 0.86rem;
    background: var(--glass);
    color: var(--text);
    border-color: var(--border);
    resize: none;
    width: 100%;
    box-sizing: border-box;
  }

  .todo-detail textarea:focus,
  .todo-detail input[type='date']:focus,
  .todo-detail select:focus {
    border-color: color-mix(in srgb, var(--accent-color), transparent 50%);
    box-shadow: 0 0 0 3px color-mix(in srgb, var(--accent-color), transparent 80%);
  }

  .todo-detail select option { background: var(--bg-2); }

  .date-picker-wrap {
    position: relative;
    display: flex;
    align-items: center;
  }

  .date-picker-wrap input[type='date'] {
    padding-right: 34px;
    cursor: pointer;
    color-scheme: dark;
  }

  .date-picker-wrap input[type='date']::-webkit-calendar-picker-indicator {
    opacity: 0.5;
    cursor: pointer;
    filter: invert(1);
  }

  .date-clear {
    position: absolute;
    right: 8px;
    width: 20px;
    height: 20px;
    border-radius: 999px;
    border: none;
    background: var(--glass-md);
    color: var(--text-muted);
    font-size: 0.85rem;
    font-weight: 900;
    line-height: 1;
    cursor: pointer;
    display: grid;
    place-items: center;
    padding: 0;
    transition: background 0.15s, color 0.15s;
  }
  .date-clear:hover { background: rgba(248,113,113,0.2); color: #fca5a5; }

  .detail-row {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 10px;
  }

  /* Checkbox */
  .check {
    width: 26px;
    height: 26px;
    border: 1.5px solid color-mix(in srgb, var(--accent-color), transparent 45%);
    border-radius: 8px;
    background: transparent;
    color: var(--accent-color);
    font-weight: 900;
    flex-shrink: 0;
    transition: background 0.15s, border-color 0.15s, box-shadow 0.15s, transform 0.12s;
    display: grid;
    place-items: center;
  }
  .check:hover {
    background: color-mix(in srgb, var(--accent-color), transparent 82%);
    border-color: var(--accent-color);
    box-shadow: 0 0 12px color-mix(in srgb, var(--accent-color), transparent 55%);
    transform: scale(1.08);
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

  @media (max-width: 900px) {
    .detail-row { grid-template-columns: 1fr; }
  }
</style>
