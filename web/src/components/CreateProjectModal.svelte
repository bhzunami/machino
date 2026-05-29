<script>
  import { createEventDispatcher } from 'svelte'

  export let defaultColor = '#4f46e5'
  export let saving = false
  export let error = ''

  const dispatch = createEventDispatcher()

  let form = { title: '', description: '', color: defaultColor, moveDone: true }
  let newColumnTitle = ''
  let localColumns = []

  function addColumn() {
    const title = newColumnTitle.trim()
    if (!title) return
    localColumns = [...localColumns, { id: `local-${Date.now()}-${Math.random()}`, title }]
    newColumnTitle = ''
  }

  function removeColumn(id) {
    localColumns = localColumns.filter(c => c.id !== id)
  }

  function save() {
    if (!form.title.trim()) return
    dispatch('save', {
      form: {
        ...form,
        title: form.title.trim(),
        description: form.description.trim(),
      },
      columns: localColumns.map(c => c.title),
    })
  }

  function close() {
    dispatch('close')
  }
</script>

<!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
<div class="modal-overlay" on:click|self={close}>
  <div class="modal" role="dialog" aria-modal="true">
    <div class="modal-handle"></div>
    <div class="modal-header">
      <h2>Projekt erstellen</h2>
      <button class="close-btn" on:click={close} aria-label="Schließen">✕</button>
    </div>

    {#if error}
      <p class="msg error">{error}</p>
    {/if}

    <form on:submit|preventDefault={save}>
      <div class="modal-body">
        <div class="field">
          <label for="cp-title">Titel <span class="required">*</span></label>
          <input id="cp-title" type="text" bind:value={form.title} placeholder="Projektname…" required autocomplete="off" />
        </div>

        <div class="field">
          <label for="cp-desc">Beschreibung</label>
          <textarea id="cp-desc" rows="2" bind:value={form.description} placeholder="Kurze Beschreibung (optional)…"></textarea>
        </div>

        <div class="field">
          <label for="cp-color">Projektfarbe</label>
          <div class="color-row">
            <input id="cp-color" type="color" bind:value={form.color} />
            <span class="color-preview" style="background:{form.color}"></span>
            <span class="color-value">{form.color}</span>
          </div>
        </div>

        <label class="toggle-row" for="cp-move-done">
          <input id="cp-move-done" type="checkbox" checked={!form.moveDone} on:change={(e) => form.moveDone = !e.target.checked} />
          <span class="toggle-text">Packlisten-Modus</span>
          <span class="info-icon" title="Abgehakte Todos bleiben in ihrer Kategorie sichtbar, statt nach unten verschoben zu werden.">ℹ</span>
        </label>

        <div class="columns-section">
          <div class="columns-header">
            <span class="columns-label">Spalten</span>
          </div>
          {#if localColumns.length > 0}
            <ul class="columns-list">
              {#each localColumns as col (col.id)}
                <li class="column-item">
                  <span class="column-title">{col.title}</span>
                  <button class="column-delete-btn" type="button" on:click={() => removeColumn(col.id)} title="Spalte entfernen" aria-label="Spalte entfernen">✕</button>
                </li>
              {/each}
            </ul>
          {/if}
          <div class="column-add-form">
            <input
              class="column-add-input"
              type="text"
              placeholder="Neue Spalte…"
              bind:value={newColumnTitle}
              autocomplete="off"
              on:keydown={(e) => { if (e.key === 'Enter') { e.preventDefault(); addColumn(); } }}
            />
            <button type="button" class="btn-add-col" on:click={addColumn} disabled={!newColumnTitle.trim()}>+</button>
          </div>
        </div>
      </div>

      <div class="actions">
        <button type="button" class="btn secondary" on:click={close} disabled={saving}>Abbrechen</button>
        <button type="submit" class="btn" disabled={saving || !form.title.trim()}>
          {saving ? 'Erstellen…' : 'Erstellen'}
        </button>
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

  form {
    display: flex;
    flex-direction: column;
    gap: 1rem;
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

  .columns-header { margin-bottom: 0.65rem; }

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
    color: var(--text);
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
