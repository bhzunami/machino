<script>
  import { createEventDispatcher } from 'svelte'

  export let todo

  const dispatch = createEventDispatcher()

  let form = { title: '', description: '', dueDate: '', priority: 'normal' }
  let error = ''
  let confirmDelete = false

  $: if (todo) {
    form = {
      title: todo.title || '',
      description: todo.description || '',
      dueDate: todo.dueDate ? String(todo.dueDate).slice(0, 10) : '',
      priority: todo.priority || 'normal',
    }
    error = ''
    confirmDelete = false
  }

  function save() {
    error = ''
    if (!form.title.trim()) {
      error = 'Titel ist Pflicht.'
      return
    }
    dispatch('save', {
      todoId: todo.id,
      form: {
        ...form,
        title: form.title.trim(),
        description: form.description.trim(),
      },
    })
  }

  function close() {
    dispatch('close')
  }

  function deleteTodo() {
    if (!confirmDelete) {
      confirmDelete = true
      return
    }
    dispatch('delete', todo)
  }
</script>

<!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
<div class="modal-overlay" on:click|self={close}>
  <div class="modal" role="dialog" aria-modal="true" aria-labelledby="todo-edit-title">
    <div class="modal-handle"></div>
    <div class="modal-header">
      <h2 id="todo-edit-title">Todo bearbeiten</h2>
      <button class="close-btn" type="button" on:click={close} aria-label="Schließen">✕</button>
    </div>

    {#if error}
      <p class="msg error">{error}</p>
    {/if}

    <form on:submit|preventDefault={save}>
      <div class="modal-body">
        <div class="field">
          <label for="todo-title">Titel <span class="required">*</span></label>
          <input id="todo-title" type="text" bind:value={form.title} placeholder="Todo…" required autocomplete="off" />
        </div>

        <div class="field">
          <label for="todo-desc">Beschreibung</label>
          <textarea id="todo-desc" rows="3" bind:value={form.description} placeholder="Beschreibung hinzufügen…"></textarea>
        </div>

        <div class="field-grid">
          <div class="field">
            <label for="todo-due-date">Fällig bis</label>
            <div class="date-row">
              <input id="todo-due-date" type="date" bind:value={form.dueDate} />
              {#if form.dueDate}
                <button type="button" class="btn-clear-date" on:click={() => (form.dueDate = '')} aria-label="Datum löschen" title="Datum löschen">✕</button>
              {/if}
            </div>
          </div>

          <div class="field">
            <label for="todo-priority">Priorität</label>
            <select id="todo-priority" bind:value={form.priority}>
              <option value="low">Niedrig</option>
              <option value="normal">Normal</option>
              <option value="high">Hoch</option>
            </select>
          </div>
        </div>
      </div>

      <div class="actions">
        <button type="button" class="btn danger" class:confirm={confirmDelete} on:click={deleteTodo}>
          {confirmDelete ? 'Wirklich löschen?' : 'Löschen'}
        </button>
        <div class="actions-right">
          <button type="button" class="btn secondary" on:click={close}>Abbrechen</button>
          <button type="submit" class="btn" disabled={!form.title.trim()}>Speichern</button>
        </div>
      </div>
    </form>
  </div>
</div>

<style>
  .modal {
    background: var(--bg-2);
    border: 1px solid var(--border);
    border-radius: 14px;
    width: 100%;
    max-width: 440px;
    box-shadow: 0 24px 64px rgba(0, 0, 0, 0.5);
    animation: pop-in 0.18s cubic-bezier(0.16, 1, 0.3, 1);
    display: flex;
    flex-direction: column;
    max-height: 90dvh;
    overflow: hidden;
  }

  .modal-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 1.25rem 1.5rem 0;
    flex-shrink: 0;
  }

  h2 {
    font-size: 1.1rem;
    font-weight: 700;
    color: var(--text);
    margin: 0;
  }

  .msg.error {
    margin: 0.75rem 1.5rem 0;
    flex-shrink: 0;
  }

  form {
    display: flex;
    flex-direction: column;
    flex: 1;
    min-height: 0;
    overflow: hidden;
  }

  .modal-body {
    overflow-y: auto;
    padding: 1.25rem 1.5rem;
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .field,
  .field-grid {
    min-width: 0;
  }

  .field {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .field-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 10px;
  }

  .field label {
    display: flex;
    align-items: center;
    gap: 3px;
    font-size: 0.78rem;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.06em;
    color: var(--text-muted);
  }

  .required { color: #f87171; }

  .field input,
  .field textarea,
  .field select {
    background: var(--bg-3);
    border: 1px solid var(--border);
    border-radius: 10px;
    padding: 0.55rem 0.85rem;
    color: var(--text);
    font-size: 0.92rem;
    outline: none;
    transition: border-color 0.15s, box-shadow 0.15s;
    resize: none;
    width: 100%;
    box-sizing: border-box;
  }

  .field input:focus,
  .field textarea:focus,
  .field select:focus {
    border-color: rgba(99, 102, 241, 0.7);
    box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.12);
  }

  .field select option { background: var(--bg-2); }

  .actions {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 8px;
    padding: 1rem 1.5rem;
    border-top: 1px solid var(--border);
    flex-shrink: 0;
  }

  .actions-right {
    display: flex;
    gap: 8px;
  }

  .date-row {
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .date-row input {
    flex: 1;
  }

  .btn-clear-date {
    background: transparent;
    border: 1px solid var(--border);
    border-radius: 6px;
    color: var(--text-faint);
    font-size: 0.72rem;
    padding: 4px 7px;
    cursor: pointer;
    line-height: 1;
    flex-shrink: 0;
    transition: color 0.12s, border-color 0.12s;
  }
  .btn-clear-date:hover { color: #f87171; border-color: rgba(248,113,113,0.4); }

  .btn.danger {
    background: transparent;
    border: 1px solid rgba(248,113,113,0.45);
    color: #f87171;
    box-shadow: none;
    transition: opacity 0.2s, transform 0.15s, box-shadow 0.2s, background 0.12s, border-color 0.12s;
  }
  .btn.danger:hover { background: rgba(248,113,113,0.1); border-color: rgba(248,113,113,0.7); box-shadow: none; transform: translateY(-1px); }
  .btn.danger.confirm { background: rgba(248,113,113,0.15); border-color: #f87171; }

  @media (max-width: 560px) {
    .field-grid { grid-template-columns: 1fr; }
    .modal {
      border-bottom-left-radius: 0;
      border-bottom-right-radius: 0;
      max-width: 100%;
      max-height: 92dvh;
      animation: slide-up 0.22s cubic-bezier(0.16, 1, 0.3, 1);
    }
  }
</style>
