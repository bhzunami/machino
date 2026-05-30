<script>
  import { onMount } from 'svelte'
  import { api } from '../lib/api.js'
  import { API } from '../lib/constants.js'
  import { user, error, success } from '../lib/stores.js'

  let users = []
  let loading = true
  let saving = false
  let selectedUser = null
  let form = emptyForm()

  onMount(loadUsers)

  function emptyForm() {
    return {
      email: '',
      name: '',
      role: 'user',
      searchable: true,
      password: '',
    }
  }

  function formatDate(value) {
    if (!value) return ''
    return new Date(value).toLocaleDateString('de-DE')
  }

  function selectUser(next) {
    selectedUser = next
    form = {
      email: next.email,
      name: next.name || '',
      role: next.role || 'user',
      searchable: next.searchable ?? true,
      password: '',
    }
    error.set('')
    success.set('')
  }

  async function loadUsers() {
    loading = true
    error.set('')
    try {
      const payload = await api(API.adminUsers)
      users = payload.users || []
      if (selectedUser) {
        const refreshed = users.find((u) => u.id === selectedUser.id)
        if (refreshed) {
          selectUser(refreshed)
        } else {
          selectedUser = null
        }
      }
      if (!selectedUser && users.length) {
        selectUser(users[0])
      }
    } catch (err) {
      error.set(err.message)
    } finally {
      loading = false
    }
  }

  async function saveUser() {
    if (!selectedUser) return
    saving = true
    error.set('')
    success.set('')
    try {
      const payload = await api(API.adminUser(selectedUser.id), {
        method: 'PUT',
        body: {
          email: form.email,
          name: form.name,
          role: form.role,
          searchable: form.searchable,
        },
      })
      users = users.map((u) => (u.id === payload.user.id ? payload.user : u))
      if (payload.user.id === $user.id) {
        user.set(payload.user)
      }
      selectUser(payload.user)
      success.set('Benutzer gespeichert.')
    } catch (err) {
      error.set(err.message)
    } finally {
      saving = false
    }
  }

  async function resetPassword() {
    if (!selectedUser || !form.password) return
    saving = true
    error.set('')
    success.set('')
    try {
      await api(API.adminUserPassword(selectedUser.id), {
        method: 'PUT',
        body: { password: form.password },
      })
      form = { ...form, password: '' }
      success.set('Passwort zurückgesetzt.')
    } catch (err) {
      error.set(err.message)
    } finally {
      saving = false
    }
  }

  async function deleteUser() {
    if (!selectedUser || selectedUser.id === $user.id) return
    if (!confirm(`Benutzer "${selectedUser.email}" wirklich löschen? Alle zugehörigen Daten werden entfernt.`)) return
    saving = true
    error.set('')
    success.set('')
    try {
      await api(API.adminUser(selectedUser.id), { method: 'DELETE' })
      users = users.filter((u) => u.id !== selectedUser.id)
      selectedUser = users[0] || null
      if (selectedUser) {
        selectUser(selectedUser)
      } else {
        form = emptyForm()
      }
      success.set('Benutzer gelöscht.')
    } catch (err) {
      error.set(err.message)
    } finally {
      saving = false
    }
  }
</script>

<section class="admin-page">
  <aside class="card user-list">
    <div class="panel-header">
      <div>
        <h2>Benutzer</h2>
        <p>{users.length} Konten</p>
      </div>
      <button class="btn ghost" type="button" on:click={loadUsers} disabled={loading}>Aktualisieren</button>
    </div>

    {#if loading}
      <p class="muted list-state">Benutzer werden geladen...</p>
    {:else if users.length === 0}
      <p class="muted list-state">Keine Benutzer vorhanden.</p>
    {:else}
      <div class="users">
        {#each users as item (item.id)}
          <button
            type="button"
            class="user-row"
            class:active={selectedUser?.id === item.id}
            on:click={() => selectUser(item)}
          >
            <span class="user-avatar">{(item.name || item.email || '?').slice(0, 1).toUpperCase()}</span>
            <span class="user-meta">
              <strong>{item.name || item.email}</strong>
              <small>{item.email}</small>
            </span>
            {#if item.role === 'admin'}
              <span class="role-badge">Admin</span>
            {/if}
          </button>
        {/each}
      </div>
    {/if}
  </aside>

  <div class="card admin-form stack">
    {#if selectedUser}
      <div class="panel-header">
        <div>
          <h2>Benutzer bearbeiten</h2>
          <p>Erstellt am {formatDate(selectedUser.createdAt)}</p>
        </div>
        {#if selectedUser.id === $user.id}
          <span class="self-badge">Du</span>
        {/if}
      </div>

      <form class="stack" on:submit|preventDefault={saveUser}>
        <label>
          Name
          <input bind:value={form.name} autocomplete="name" />
        </label>
        <label>
          E-Mail
          <input bind:value={form.email} type="email" autocomplete="email" />
        </label>
        <label>
          Rolle
          <select bind:value={form.role} disabled={selectedUser.id === $user.id}>
            <option value="user">Benutzer</option>
            <option value="admin">Admin</option>
          </select>
        </label>

        <div class="toggle-row">
          <div class="toggle-text">
            <span class="toggle-label">In Projektsuche auffindbar</span>
            <span class="toggle-desc">Andere Nutzer können dieses Konto beim Teilen von Projekten finden.</span>
          </div>
          <button
            type="button"
            class="toggle-btn"
            class:on={form.searchable}
            aria-label="Auffindbarkeit umschalten"
            aria-pressed={form.searchable}
            on:click={() => (form.searchable = !form.searchable)}
          >
            <span class="toggle-knob"></span>
          </button>
        </div>

        <button class="btn secondary" type="submit" disabled={saving}>Benutzer speichern</button>
      </form>

      <div class="section-divider"></div>

      <div class="stack">
        <label>
          Neues Passwort
          <input bind:value={form.password} type="password" autocomplete="new-password" placeholder="Mindestens 8 Zeichen..." />
        </label>
        <button class="btn secondary" type="button" on:click={resetPassword} disabled={saving || !form.password}>
          Passwort zurücksetzen
        </button>
      </div>

      <div class="section-divider"></div>

      <button class="danger-btn" type="button" on:click={deleteUser} disabled={saving || selectedUser.id === $user.id}>
        Benutzer löschen
      </button>
      {#if selectedUser.id === $user.id}
        <p class="muted hint">Dein eigener Admin-User kann nicht gelöscht werden.</p>
      {/if}
    {:else}
      <p class="muted list-state">Wähle links einen Benutzer aus.</p>
    {/if}
  </div>
</section>

<style>
  .admin-page {
    display: grid;
    grid-template-columns: minmax(260px, 360px) minmax(360px, 620px);
    gap: 20px;
    padding: 28px 28px 40px;
    align-items: start;
  }

  .user-list,
  .admin-form {
    padding: 20px 24px;
    background: var(--glass);
    border-color: var(--border);
    width: 100%;
    box-sizing: border-box;
  }

  .panel-header {
    display: flex;
    justify-content: space-between;
    gap: 16px;
    align-items: flex-start;
  }

  h2 {
    margin: 0 0 4px;
    font-size: 1.15rem;
    font-weight: 800;
    color: var(--text);
    letter-spacing: -0.02em;
  }

  .panel-header p,
  .hint {
    margin: 0;
    font-size: 0.82rem;
  }

  .stack {
    display: grid;
    gap: 14px;
  }

  .users {
    display: grid;
    gap: 8px;
    margin-top: 16px;
    max-height: min(62vh, 680px);
    overflow: auto;
  }

  .user-row {
    display: grid;
    grid-template-columns: 36px minmax(0, 1fr) auto;
    align-items: center;
    gap: 10px;
    width: 100%;
    padding: 10px;
    border-radius: 14px;
    border: 1px solid transparent;
    background: transparent;
    color: var(--text);
    text-align: left;
  }

  .user-row:hover,
  .user-row.active {
    background: var(--glass-hover);
    border-color: var(--border-hover);
  }

  .user-avatar {
    width: 36px;
    height: 36px;
    border-radius: 12px;
    display: grid;
    place-items: center;
    background: linear-gradient(135deg, #6366f1, #4f46e5);
    color: #fff;
    font-weight: 800;
  }

  .user-meta {
    display: grid;
    min-width: 0;
  }

  .user-meta strong,
  .user-meta small {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .user-meta strong {
    font-size: 0.9rem;
  }

  .user-meta small {
    color: var(--text-muted);
    font-size: 0.75rem;
  }

  .role-badge,
  .self-badge {
    border-radius: 999px;
    padding: 4px 9px;
    background: rgba(99,102,241,0.15);
    border: 1px solid rgba(99,102,241,0.28);
    color: #a5b4fc;
    font-size: 0.72rem;
    font-weight: 800;
  }

  .section-divider {
    height: 1px;
    background: var(--border);
    margin: 4px 0;
  }

  .toggle-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
    padding: 12px 14px;
    border-radius: 12px;
    background: var(--glass);
    border: 1px solid var(--border);
  }

  .toggle-text {
    display: flex;
    flex-direction: column;
    gap: 3px;
  }

  .toggle-label {
    font-size: 0.88rem;
    font-weight: 600;
    color: var(--text);
  }

  .toggle-desc {
    font-size: 0.78rem;
    color: var(--text-muted);
    line-height: 1.4;
  }

  .toggle-btn {
    flex-shrink: 0;
    width: 44px;
    height: 24px;
    border-radius: 12px;
    border: none;
    background: var(--border);
    cursor: pointer;
    position: relative;
    transition: background 0.2s;
    padding: 0;
  }

  .toggle-btn.on {
    background: #6366f1;
  }

  .toggle-knob {
    position: absolute;
    top: 3px;
    left: 3px;
    width: 18px;
    height: 18px;
    border-radius: 50%;
    background: #fff;
    box-shadow: 0 1px 4px rgba(0,0,0,0.3);
    transition: transform 0.2s cubic-bezier(0.16,1,0.3,1);
    display: block;
  }

  .toggle-btn.on .toggle-knob {
    transform: translateX(20px);
  }

  .danger-btn {
    border-radius: 999px;
    padding: 10px 20px;
    background: rgba(248,113,113,0.12);
    border: 1px solid rgba(248,113,113,0.25);
    color: #fca5a5;
    font-weight: 800;
  }

  .danger-btn:disabled {
    opacity: 0.45;
    cursor: not-allowed;
  }

  .list-state {
    padding: 16px 0 0;
  }

  @media (max-width: 900px) {
    .admin-page {
      grid-template-columns: 1fr;
      padding: 18px 16px 32px;
    }

    .user-list,
    .admin-form {
      padding: 14px 12px;
      border-radius: 14px;
    }
  }
</style>
