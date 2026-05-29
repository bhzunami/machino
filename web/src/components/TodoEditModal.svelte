<script>
  import { createEventDispatcher } from 'svelte'

  export let todo

  const dispatch = createEventDispatcher()

  let form = { title: '', description: '', dueDate: '', priority: 'normal' }
  let error = ''

  $: if (todo) {
    form = {
      title: todo.title || '',
      description: todo.description || '',
      dueDate: todo.dueDate ? String(todo.dueDate).slice(0, 10) : '',
      priority: todo.priority || 'normal',
    }
    error = ''
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
            <input id="todo-due-date" type="date" bind:value={form.dueDate} />
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
        <button type="button" class="btn secondary" on:click={close}>Abbrechen</button>
        <button type="submit" class="btn" disabled={!form.title.trim()}>Speichern</button>
      </div>
    </form>
  </div>
</div>

<style>
  .modal-overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.6);
    backdrop-filter: blur(4px);
    -webkit-backdrop-filter: blur(4px);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
    padding: 1rem;
  }

  .modal {
    background: var(--bg-2);
    border: 1px solid var(--border);
    border-radius: 14px;
    width: 100%;
    max-width: 440px;
    box-shadow: 0 24px 64px rgba(0, 0, 0, 0.5);
    animation: pop-in 0.18s cubic-bezier(0.16,1,0.3,1);
    display: flex;
    flex-direction: column;
    max-height: 90dvh;
    overflow: hidden;
  }

  @keyframes pop-in {
    from { opacity: 0; transform: scale(0.95) translateY(-8px); }
    to   { opacity: 1; transform: scale(1) translateY(0); }
  }

  .modal-handle {
    display: none;
    width: 36px;
    height: 4px;
    border-radius: 2px;
    background: var(--border-hover);
    margin: 10px auto 0;
    flex-shrink: 0;
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

  .close-btn {
    background: var(--glass);
    border: 1px solid var(--border);
    color: var(--text-muted);
    cursor: pointer;
    font-size: 0.9rem;
    padding: 0.3rem 0.5rem;
    border-radius: 6px;
    line-height: 1;
    transition: color 0.15s, background 0.15s;
  }
  .close-btn:hover { color: var(--text); background: var(--glass-hover); }

  .msg.error {
    padding: 0.6rem 0.8rem;
    border-radius: 8px;
    font-size: 0.85rem;
    margin: 0.75rem 1.5rem 0;
    background: rgba(239, 68, 68, 0.12);
    color: #f87171;
    border: 1px solid rgba(239, 68, 68, 0.25);
    flex-shrink: 0;
  }
  :global([data-theme="light"]) .msg.error { color: #dc2626; }

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
    justify-content: flex-end;
    gap: 8px;
    padding: 1rem 1.5rem;
    border-top: 1px solid var(--border);
    flex-shrink: 0;
  }

  @media (max-width: 560px) {
    .field-grid { grid-template-columns: 1fr; }
    .modal-overlay { padding: 0; align-items: flex-end; }
    .modal {
      border-bottom-left-radius: 0;
      border-bottom-right-radius: 0;
      max-width: 100%;
      max-height: 92dvh;
      animation: slide-up 0.22s cubic-bezier(0.16,1,0.3,1);
    }
    .modal-handle { display: block; }
  }

  @keyframes slide-up {
    from { opacity: 0; transform: translateY(24px); }
    to   { opacity: 1; transform: translateY(0); }
  }
</style>
