<script>
  import { createEventDispatcher, onMount } from 'svelte';
  import { API } from '../lib/constants.js';

  export let project = null;

  const dispatch = createEventDispatcher();

  let members = [];
  let email = '';
  let loading = false;
  let error = '';
  let successMsg = '';

  // Typeahead state
  let searchQuery = '';
  let searchResults = [];
  let searchLoading = false;
  let selectedUser = null; // { id, name } chosen from dropdown
  let searchDebounce = null;
  let showDropdown = false;

  $: isOwner = project?.isOwner;

  onMount(() => {
    if (project) loadMembers();
  });

  async function loadMembers() {
    try {
      const res = await fetch(API.projectMembers(project.id), { credentials: 'include' });
      if (res.ok) members = (await res.json()) ?? [];
    } catch (e) {
      console.error('load members', e);
    }
  }

  function onSearchInput() {
    selectedUser = null;
    clearTimeout(searchDebounce);
    if (searchQuery.length < 3) {
      searchResults = [];
      showDropdown = false;
      return;
    }
    searchDebounce = setTimeout(async () => {
      searchLoading = true;
      try {
        const res = await fetch(`${API.usersSearch}?q=${encodeURIComponent(searchQuery)}`, {
          credentials: 'include',
        });
        if (res.ok) {
          const data = await res.json();
          searchResults = data.users ?? [];
          showDropdown = searchResults.length > 0;
        }
      } catch (e) {
        console.error('user search', e);
      } finally {
        searchLoading = false;
      }
    }, 250);
  }

  function pickUser(u) {
    selectedUser = u;
    searchQuery = u.name;
    showDropdown = false;
    searchResults = [];
  }

  async function invite() {
    error = '';
    successMsg = '';
    // If user was picked from typeahead, use their name as label; always send email field
    // (server accepts email OR we can extend to accept userId — we send email for backward compat)
    // If user typed a raw email without picking, send that directly.
    const payload = selectedUser
      ? { userId: selectedUser.id }
      : { email: email.trim().toLowerCase() };

    const hasValue = selectedUser || email.trim();
    if (!hasValue) return;
    loading = true;
    try {
      const res = await fetch(API.projectMembers(project.id), {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify(payload),
      });
      if (res.ok) {
        const label = selectedUser ? selectedUser.name : email.trim();
        successMsg = `${label} wurde eingeladen.`;
        email = '';
        searchQuery = '';
        selectedUser = null;
        await loadMembers();
      } else {
        const body = await res.json().catch(() => ({}));
        if (res.status === 404) error = 'Kein Benutzer mit dieser E-Mail gefunden.';
        else if (res.status === 401) error = 'Nur der Besitzer darf Mitglieder einladen.';
        else error = body.error || 'Einladung fehlgeschlagen.';
      }
    } finally {
      loading = false;
    }
  }

  async function removeMember(uid) {
    error = '';
    try {
      const res = await fetch(API.projectMember(project.id, uid), {
        method: 'DELETE',
        credentials: 'include',
      });
      if (res.ok) {
        members = members.filter((m) => m.userId !== uid);
      } else {
        error = 'Entfernen fehlgeschlagen.';
      }
    } catch (e) {
      error = 'Entfernen fehlgeschlagen.';
    }
  }

  function close() {
    dispatch('close');
  }
</script>

<!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
<div class="modal-overlay" on:click|self={close}>
  <div class="modal" role="dialog" aria-modal="true">
    <div class="modal-header">
      <h2>Teilen — {project.title}</h2>
      <button class="close-btn" on:click={close} aria-label="Schließen">✕</button>
    </div>

    {#if error}
      <p class="msg error">{error}</p>
    {/if}
    {#if successMsg}
      <p class="msg success">{successMsg}</p>
    {/if}

    {#if isOwner}
      <form class="invite-form" on:submit|preventDefault={invite}>
        <div class="search-wrap">
          <input
            type="text"
            bind:value={searchQuery}
            on:input={onSearchInput}
            on:blur={() => setTimeout(() => (showDropdown = false), 150)}
            on:focus={() => searchResults.length > 0 && (showDropdown = true)}
            placeholder="Name oder E-Mail suchen…"
            autocomplete="off"
          />
          {#if searchLoading}
            <span class="search-spinner">⟳</span>
          {/if}
          {#if showDropdown}
            <ul class="search-dropdown">
              {#each searchResults as u (u.id)}
                <!-- svelte-ignore a11y-click-events-have-key-events -->
                <li on:mousedown={() => pickUser(u)}>
                  <span class="sd-avatar">{u.name[0].toUpperCase()}</span>
                  <span class="sd-name">{u.name}</span>
                </li>
              {/each}
            </ul>
          {/if}
        </div>
        {#if !selectedUser}
          <input
            class="email-fallback"
            type="email"
            bind:value={email}
            placeholder="…oder direkt E-Mail eingeben"
            autocomplete="off"
          />
        {:else}
          <div class="selected-chip">
            <span class="sd-avatar small">{selectedUser.name[0].toUpperCase()}</span>
            {selectedUser.name}
            <button type="button" class="chip-remove" on:click={() => { selectedUser = null; searchQuery = ''; }}>✕</button>
          </div>
        {/if}
        <button type="submit" disabled={loading || (!selectedUser && !email.trim())}>
          {loading ? '…' : 'Einladen'}
        </button>
      </form>
      <p class="search-hint">Nutzer erscheinen in der Suche nur wenn sie dies in ihrem Profil aktiviert haben. Du kannst immer direkt per E-Mail einladen.</p>
    {/if}

    <ul class="member-list">
      {#each members as m (m.userId)}
        <li class="member-row">
          <div class="member-avatar">{(m.name || m.email || '?')[0].toUpperCase()}</div>
          <div class="member-info">
            <span class="member-name">{m.name || '—'}</span>
            <span class="member-email">{m.email}</span>
          </div>
          <span class="member-role {m.role}">{m.role === 'owner' ? 'Besitzer' : 'Mitglied'}</span>
          {#if isOwner && m.role !== 'owner'}
            <button class="remove-btn" on:click={() => removeMember(m.userId)} title="Entfernen">✕</button>
          {/if}
        </li>
      {/each}
    </ul>
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
    max-width: 480px;
    max-height: 90vh;
    overflow-y: auto;
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
    font-weight: 600;
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

  .msg {
    padding: 0.6rem 0.8rem;
    border-radius: 8px;
    font-size: 0.85rem;
    margin-bottom: 0.75rem;
  }
  .msg.error {
    background: rgba(239, 68, 68, 0.12);
    color: #f87171;
    border: 1px solid rgba(239, 68, 68, 0.25);
  }
  .msg.success {
    background: rgba(34, 197, 94, 0.12);
    color: #4ade80;
    border: 1px solid rgba(34, 197, 94, 0.25);
  }
  :global([data-theme="light"]) .msg.error { color: #dc2626; }
  :global([data-theme="light"]) .msg.success { color: #16a34a; }

  /* Invite form */
  .invite-form {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    margin-bottom: 0.5rem;
  }

  .search-wrap {
    position: relative;
  }

  .search-wrap input {
    width: 100%;
    box-sizing: border-box;
    background: var(--bg-3);
    border: 1px solid var(--border);
    border-radius: 8px;
    padding: 0.5rem 2rem 0.5rem 0.75rem;
    color: var(--text);
    font-size: 0.9rem;
    outline: none;
    transition: border-color 0.15s;
  }
  .search-wrap input:focus { border-color: rgba(99, 102, 241, 0.7); }

  .search-spinner {
    position: absolute;
    right: 10px;
    top: 50%;
    transform: translateY(-50%);
    color: var(--text-muted);
    font-size: 1rem;
    animation: spin 0.7s linear infinite;
    pointer-events: none;
  }
  @keyframes spin { to { transform: translateY(-50%) rotate(360deg); } }

  .search-dropdown {
    position: absolute;
    top: calc(100% + 4px);
    left: 0;
    right: 0;
    z-index: 100;
    background: var(--bg-2);
    border: 1px solid var(--border);
    border-radius: 10px;
    box-shadow: 0 8px 24px rgba(0,0,0,0.3);
    list-style: none;
    margin: 0;
    padding: 4px;
    max-height: 180px;
    overflow-y: auto;
  }

  .search-dropdown li {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 8px 10px;
    border-radius: 7px;
    cursor: pointer;
    transition: background 0.1s;
    color: var(--text);
    font-size: 0.88rem;
  }
  .search-dropdown li:hover { background: var(--glass-hover); }

  .sd-avatar {
    width: 26px;
    height: 26px;
    border-radius: 50%;
    background: #6366f1;
    color: #fff;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 0.75rem;
    font-weight: 700;
    flex-shrink: 0;
  }
  .sd-avatar.small { width: 20px; height: 20px; font-size: 0.65rem; }

  .email-fallback {
    background: var(--bg-3);
    border: 1px solid var(--border);
    border-radius: 8px;
    padding: 0.45rem 0.75rem;
    color: var(--text);
    font-size: 0.85rem;
    outline: none;
    transition: border-color 0.15s;
    width: 100%;
    box-sizing: border-box;
  }
  .email-fallback:focus { border-color: rgba(99, 102, 241, 0.5); }

  .selected-chip {
    display: flex;
    align-items: center;
    gap: 8px;
    background: rgba(99, 102, 241, 0.12);
    border: 1px solid rgba(99, 102, 241, 0.3);
    border-radius: 8px;
    padding: 0.4rem 0.75rem;
    font-size: 0.88rem;
    color: var(--text);
  }

  .chip-remove {
    margin-left: auto;
    background: none;
    border: none;
    color: var(--text-muted);
    cursor: pointer;
    font-size: 0.8rem;
    padding: 0 2px;
    line-height: 1;
  }
  .chip-remove:hover { color: #f87171; }

  .invite-form button[type="submit"] {
    background: #6366f1;
    color: #fff;
    border: none;
    border-radius: 8px;
    padding: 0.5rem 1rem;
    font-size: 0.9rem;
    font-weight: 500;
    cursor: pointer;
    transition: opacity 0.15s, background 0.15s;
    align-self: flex-end;
  }
  .invite-form button[type="submit"]:hover:not(:disabled) { background: #4f46e5; }
  .invite-form button[type="submit"]:disabled { opacity: 0.45; cursor: not-allowed; }

  .search-hint {
    font-size: 0.75rem;
    color: var(--text-faint);
    margin: 0 0 1rem;
    line-height: 1.4;
  }

  /* Member list */
  .member-list {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .member-row {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 0.6rem 0.75rem;
    border-radius: 8px;
    background: var(--bg-3);
    border: 1px solid var(--border);
  }

  .member-avatar {
    width: 32px;
    height: 32px;
    border-radius: 50%;
    background: #6366f1;
    color: #fff;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 0.85rem;
    font-weight: 600;
    flex-shrink: 0;
  }

  .member-info {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
  }

  .member-name {
    font-size: 0.9rem;
    font-weight: 500;
    color: var(--text);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .member-email {
    font-size: 0.78rem;
    color: var(--text-muted);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .member-role {
    font-size: 0.75rem;
    padding: 0.2rem 0.5rem;
    border-radius: 12px;
    white-space: nowrap;
    background: var(--glass);
    color: var(--text-muted);
    border: 1px solid var(--border);
  }
  .member-role.owner {
    background: rgba(99, 102, 241, 0.15);
    color: #818cf8;
    border-color: rgba(99, 102, 241, 0.3);
  }
  :global([data-theme="light"]) .member-role.owner { color: #4f46e5; }

  .remove-btn {
    background: none;
    border: none;
    color: var(--text-muted);
    cursor: pointer;
    font-size: 0.85rem;
    padding: 0.25rem 0.4rem;
    border-radius: 4px;
    transition: color 0.15s, background 0.15s;
    flex-shrink: 0;
  }
  .remove-btn:hover { color: #f87171; background: rgba(239, 68, 68, 0.1); }
  :global([data-theme="light"]) .remove-btn:hover { color: #dc2626; }
</style>
