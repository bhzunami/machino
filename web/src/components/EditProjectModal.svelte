<script>
  import { createEventDispatcher } from 'svelte';

  export let project = null;

  const dispatch = createEventDispatcher();

  let form = { title: '', description: '', color: '#4f46e5' };
  let saving = false;
  let error = '';

  $: if (project) {
    form = {
      title: project.title || '',
      description: project.description || '',
      color: project.color || '#4f46e5',
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
</script>

<!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
<div class="modal-overlay" on:click|self={close}>
  <div class="modal" role="dialog" aria-modal="true">
    <div class="modal-header">
      <h2>Projekt bearbeiten</h2>
      <button class="close-btn" on:click={close} aria-label="Schließen">✕</button>
    </div>

    {#if error}
      <p class="msg error">{error}</p>
    {/if}

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

      <div class="actions">
        <button type="button" class="btn secondary" on:click={close}>Abbrechen</button>
        <button type="submit" class="btn" disabled={saving || !form.title.trim()}>
          {saving ? 'Speichern…' : 'Speichern'}
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
    padding: 1.5rem;
    width: 100%;
    max-width: 440px;
    box-shadow: 0 24px 64px rgba(0, 0, 0, 0.5);
    animation: pop-in 0.18s cubic-bezier(0.16,1,0.3,1);
  }

  @keyframes pop-in {
    from { opacity: 0; transform: scale(0.95) translateY(-8px); }
    to   { opacity: 1; transform: scale(1) translateY(0); }
  }

  .modal-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 1.25rem;
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
    margin-top: 0.25rem;
  }
</style>
