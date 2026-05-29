<script>
  import { createEventDispatcher } from 'svelte';
  import { columns, selectedProjectId } from '../lib/stores.js';
  import { API } from '../lib/constants.js';

  export let project = null;

  const dispatch = createEventDispatcher();

  let form = { title: '', description: '', color: '#4f46e5', moveDone: true };
  let saving = false;
  let error = '';
  let newColumnTitle = '';
  let editingColumnId = '';
  let editingColumnTitle = '';

  $: if (project) {
    form = {
      title: project.title || '',
      description: project.description || '',
      color: project.color || '#4f46e5',
      moveDone: project.moveDone !== false,
    };
  }

  async function save() {
    error = '';
    if (!form.title.trim()) { error = 'Titel ist Pflicht.'; return; }
    saving = true;
    dispatch('save', { project, form: { ...form, title: form.title.trim(), description: form.description.trim() } });
  }

  function close() {
    dispatch('close');
  }

  function addColumn() {
    const title = newColumnTitle.trim();
    if (!title) return;
    newColumnTitle = '';
    dispatch('run-or-queue', {
      method: 'POST',
      path: API.projectColumns($selectedProjectId),
      body: { title },
    });
  }

  function startEditColumn(col) {
    editingColumnId = col.id;
    editingColumnTitle = col.title;
  }

  function saveEditColumn(col) {
    const title = editingColumnTitle.trim();
    editingColumnId = '';
    if (!title || title === col.title) return;
    columns.update($cols => $cols.map(c => c.id === col.id ? { ...c, title } : c));
    dispatch('run-or-queue', {
      method: 'PATCH',
      path: API.column(col.id),
      body: { title },
    });
  }

  function deleteColumn(col) {
    columns.update($cols => $cols.filter(c => c.id !== col.id));
    dispatch('run-or-queue', {
      method: 'DELETE',
      path: API.column(col.id),
      body: null,
    });
  }
</script>

<!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
<div class="modal-overlay" on:click|self={close}>
  <div class="modal" role="dialog" aria-modal="true">
    <div class="modal-handle"></div>
    <div class="modal-header">
      <h2>Projekt bearbeiten</h2>
      <button class="close-btn" on:click={close} aria-label="Schließen">✕</button>
    </div>

    {#if error}
      <p class="msg error">{error}</p>
    {/if}

    <div class="modal-body">
      <form on:submit|preventDefault={save}>
        <div class="field">
          <label for="ep-title">Titel <span class="required">*</span></label>
          <input id="ep-title" type="text" bind:value={form.title} placeholder="Projektname…" required autocomplete="off" />
        </div>

        <div class="field">
          <label for="ep-desc">Beschreibung</label>
          <textarea id="ep-desc" rows="2" bind:value={form.description} placeholder="Kurze Beschreibung (optional)…"></textarea>
        </div>

        <div class="field">
          <label for="ep-color">Projektfarbe</label>
          <div class="color-row">
            <input id="ep-color" type="color" bind:value={form.color} />
            <span class="color-preview" style="background:{form.color}"></span>
            <span class="color-value">{form.color}</span>
          </div>
        </div>

        <label class="toggle-row" for="ep-move-done">
          <input id="ep-move-done" type="checkbox" bind:checked={form.moveDone} />
          <span class="toggle-text">Packlisten-Modus</span>
          <span class="info-icon" title="Abgehakte Todos bleiben in ihrer Kategorie sichtbar, statt nach unten verschoben zu werden.">ℹ</span>
        </label>
      </form>

      <!-- Columns management (outside the save form) -->
      <div class="columns-section">
        <div class="columns-header">
          <span class="columns-label">Spalten</span>
        </div>
        <ul class="columns-list">
          {#each $columns as col (col.id)}
            <li class="column-item">
              {#if editingColumnId === col.id}
                <input
                  class="column-edit-input"
                  type="text"
                  bind:value={editingColumnTitle}
                  on:blur={() => saveEditColumn(col)}
                  on:keydown={(e) => { if (e.key === 'Enter') saveEditColumn(col); if (e.key === 'Escape') editingColumnId = ''; }}
                  autofocus
                />
              {:else}
                <span class="column-title" on:click={() => startEditColumn(col)} on:keydown={() => {}} role="button" tabindex="0">{col.title}</span>
                <button class="column-delete-btn" on:click={() => deleteColumn(col)} title="Spalte löschen" aria-label="Spalte löschen">✕</button>
              {/if}
            </li>
          {/each}
        </ul>
        <form class="column-add-form" on:submit|preventDefault={addColumn}>
          <input
            class="column-add-input"
            type="text"
            placeholder="Neue Spalte…"
            bind:value={newColumnTitle}
            autocomplete="off"
          />
          <button type="submit" class="btn-add-col" disabled={!newColumnTitle.trim()}>+</button>
        </form>
      </div>
    </div>

    <div class="actions">
      <button type="button" class="btn secondary" on:click={close}>Abbrechen</button>
      <button type="button" class="btn" disabled={saving || !form.title.trim()} on:click={save}>
        {saving ? 'Speichern…' : 'Speichern'}
      </button>
    </div>
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

  .msg.error {
    margin: 0.75rem 1.5rem 0;
    padding: 0.6rem 0.8rem;
    border-radius: 8px;
    font-size: 0.85rem;
    background: rgba(239, 68, 68, 0.12);
    color: #f87171;
    border: 1px solid rgba(239, 68, 68, 0.25);
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
    margin-bottom: 0.75rem;
    background: rgba(239, 68, 68, 0.12);
    color: #f87171;
    border: 1px solid rgba(239, 68, 68, 0.25);
  }
  :global([data-theme="light"]) .msg.error { color: #dc2626; }

  .modal-body {
    overflow-y: auto;
    padding: 1.25rem 1.5rem;
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  form {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .field {
    display: flex;
    flex-direction: column;
    gap: 6px;
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

  .field input[type='text'],
  .field textarea {
    background: var(--bg-3);
    border: 1px solid var(--border);
    border-radius: 10px;
    padding: 0.55rem 0.85rem;
    color: var(--text);
    font-size: 0.92rem;
    outline: none;
    transition: border-color 0.15s;
    resize: none;
    width: 100%;
    box-sizing: border-box;
  }
  .field input[type='text']:focus,
  .field textarea:focus {
    border-color: rgba(99, 102, 241, 0.7);
    box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.12);
  }

  .color-row {
    display: flex;
    align-items: center;
    gap: 10px;
  }

  .field input[type='color'] {
    width: 42px;
    height: 42px;
    border-radius: 10px;
    border: 1px solid var(--border);
    padding: 3px;
    cursor: pointer;
    background: var(--bg-3);
    flex-shrink: 0;
  }

  .color-preview {
    width: 24px;
    height: 24px;
    border-radius: 50%;
    border: 2px solid var(--border);
    flex-shrink: 0;
  }

  .color-value {
    font-size: 0.82rem;
    color: var(--text-muted);
    font-family: monospace;
  }

  .actions {
    display: flex;
    justify-content: flex-end;
    gap: 8px;
    padding: 1rem 1.5rem;
    border-top: 1px solid var(--border);
    flex-shrink: 0;
  }

  .toggle-row {
    display: flex;
    align-items: center;
    gap: 8px;
    cursor: pointer;
    margin: -0.25rem 0;
  }

  .toggle-row input[type='checkbox'] {
    width: 15px;
    height: 15px;
    margin: 0;
    flex-shrink: 0;
    accent-color: var(--accent-color);
    cursor: pointer;
  }

  .toggle-text {
    font-size: 0.88rem;
    color: var(--text);
  }

  .info-icon {
    font-size: 0.7rem;
    color: var(--text-muted);
    cursor: help;
    border: 1px solid var(--border);
    border-radius: 50%;
    width: 15px;
    height: 15px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
    user-select: none;
    line-height: 1;
  }

  .columns-section {
    padding-top: 1rem;
    border-top: 1px solid var(--border);
  }

  .columns-header {
    margin-bottom: 0.65rem;
  }

  .columns-label {
    font-size: 0.78rem;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.06em;
    color: var(--text-muted);
  }

  .columns-list {
    list-style: none;
    margin: 0 0 0.6rem;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 5px;
  }

  .column-item {
    display: flex;
    align-items: center;
    gap: 8px;
    background: var(--bg-3);
    border: 1px solid var(--border);
    border-radius: 8px;
    padding: 6px 10px;
  }

  .column-title {
    flex: 1;
    font-size: 0.9rem;
    cursor: pointer;
    color: var(--text);
  }
  .column-title:hover { color: var(--accent-color); }

  .column-edit-input {
    flex: 1;
    background: transparent;
    border: none;
    outline: none;
    color: var(--text);
    font-size: 0.9rem;
    padding: 0;
  }

  .column-delete-btn {
    background: transparent;
    border: none;
    color: var(--text-faint);
    cursor: pointer;
    font-size: 0.75rem;
    padding: 2px 4px;
    border-radius: 4px;
    line-height: 1;
    transition: color 0.12s, background 0.12s;
  }
  .column-delete-btn:hover { color: #f87171; background: rgba(248,113,113,0.1); }

  .column-add-form {
    display: flex;
    gap: 6px;
  }

  .column-add-input {
    flex: 1;
    background: var(--bg-3);
    border: 1px solid var(--border);
    border-radius: 8px;
    padding: 6px 10px;
    color: var(--text);
    font-size: 0.88rem;
    outline: none;
    transition: border-color 0.12s;
  }
  .column-add-input:focus { border-color: var(--accent-color); }
  .column-add-input::placeholder { color: var(--text-faint); }

  .btn-add-col {
    background: var(--accent-color);
    border: none;
    color: #fff;
    font-size: 1.1rem;
    font-weight: 700;
    border-radius: 8px;
    width: 32px;
    cursor: pointer;
    transition: opacity 0.12s;
    flex-shrink: 0;
  }
  .btn-add-col:disabled { opacity: 0.4; cursor: not-allowed; }
  .btn-add-col:not(:disabled):hover { opacity: 0.85; }

  @media (max-width: 560px) {
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
