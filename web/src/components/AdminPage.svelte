<script>
  import { onMount } from 'svelte'
  import { api } from '../lib/api.js'
  import { API } from '../lib/constants.js'
  import { user, error, success } from '../lib/stores.js'

  let users = []
  let loading = true
  let saving = false
  let settingsLoading = false
  let activeTab = 'users'
  let selectedUser = null
  let form = emptyForm()
  let settings = null
  let settingsForm = emptySettingsForm()
  let showCreateForm = false
  let createForm = emptyCreateForm()

  onMount(async () => {
    await loadUsers()
    await loadSettings()
  })

  function emptyForm() {
    return {
      email: '',
      name: '',
      role: 'user',
      searchable: true,
      password: '',
    }
  }

  function emptyCreateForm() {
    return { email: '', name: '', password: '', role: 'user' }
  }

  function emptySettingsForm() {
    return {
      appDomain: '',
      registrationEnabled: true,
      smtpHost: '',
      smtpPort: '587',
      smtpUsername: '',
      smtpPassword: '',
      smtpFrom: '',
    }
  }

  function formatDate(value) {
    if (!value) return ''
    return new Date(value).toLocaleDateString('de-DE')
  }

  function selectUser(next) {
    selectedUser = next
    showCreateForm = false
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

  async function createUser() {
    saving = true
    error.set('')
    success.set('')
    try {
      const payload = await api(API.adminUsers, {
        method: 'POST',
        body: createForm,
      })
      showCreateForm = false
      createForm = emptyCreateForm()
      await loadUsers()
      const created = users.find((u) => u.id === payload.user.id)
      if (created) selectUser(created)
      success.set('Benutzer erstellt.')
    } catch (err) {
      error.set(err.message)
    } finally {
      saving = false
    }
  }

  function fillSettingsForm(next) {
    settings = next
    settingsForm = {
      appDomain: next.appDomain || '',
      registrationEnabled: next.registrationEnabled ?? true,
      smtpHost: next.smtpHost || '',
      smtpPort: next.smtpPort || '587',
      smtpUsername: next.smtpUsername || '',
      smtpPassword: '',
      smtpFrom: next.smtpFrom || '',
    }
  }

  async function loadSettings() {
    settingsLoading = true
    error.set('')
    try {
      const payload = await api(API.adminSettings)
      fillSettingsForm(payload.settings)
    } catch (err) {
      error.set(err.message)
    } finally {
      settingsLoading = false
    }
  }

  async function saveSettings() {
    saving = true
    error.set('')
    success.set('')
    try {
      const payload = await api(API.adminSettings, {
        method: 'PUT',
        body: settingsForm,
      })
      fillSettingsForm(payload.settings)
      success.set('Einstellungen gespeichert.')
    } catch (err) {
      error.set(err.message)
    } finally {
      saving = false
    }
  }
</script>

<section class="admin-shell">
  <div class="admin-tabs">
    <button type="button" class:active={activeTab === 'users'} on:click={() => (activeTab = 'users')}>Benutzer</button>
    <button type="button" class:active={activeTab === 'settings'} on:click={() => (activeTab = 'settings')}>Einstellungen</button>
  </div>

  {#if activeTab === 'users'}
    <div class="admin-page">
      <aside class="card user-list">
        <div class="panel-header">
          <div>
            <h2>Benutzer</h2>
            <p>{users.length} Konten</p>
          </div>
          <div class="panel-actions">
            <button class="btn ghost" type="button" on:click={() => { showCreateForm = true; selectedUser = null; createForm = emptyCreateForm(); error.set(''); success.set('') }}>
              + Neu
            </button>
            <button class="btn ghost" type="button" on:click={loadUsers} disabled={loading}>Aktualisieren</button>
          </div>
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
        {#if showCreateForm}
          <div class="panel-header">
            <div>
              <h2>Neuen Benutzer erstellen</h2>
              <p>Der Benutzer kann sich danach einloggen.</p>
            </div>
          </div>

          <form class="stack" on:submit|preventDefault={createUser}>
            <label>
              Name
              <input bind:value={createForm.name} placeholder="Vollständiger Name" autocomplete="off" required />
            </label>
            <label>
              E-Mail
              <input bind:value={createForm.email} type="email" placeholder="name@example.com" autocomplete="off" required />
            </label>
            <label>
              Passwort
              <input bind:value={createForm.password} type="password" placeholder="Mindestens 8 Zeichen..." autocomplete="new-password" required />
            </label>
            <label>
              Rolle
              <select bind:value={createForm.role}>
                <option value="user">Benutzer</option>
                <option value="admin">Admin</option>
              </select>
            </label>
            <button class="btn secondary" type="submit" disabled={saving}>Benutzer erstellen</button>
            <button class="btn cancel" type="button" on:click={() => { showCreateForm = false; createForm = emptyCreateForm() }}>Abbrechen</button>
          </form>
        {:else if selectedUser}
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
    </div>
  {:else}
    <div class="settings-page">
      <div class="card admin-form settings-card stack">
        <div class="panel-header">
          <div>
            <h2>Globale Einstellungen</h2>
          </div>
          <button class="btn ghost" type="button" on:click={loadSettings} disabled={settingsLoading}>Aktualisieren</button>
        </div>

        {#if settingsLoading}
          <p class="muted list-state">Einstellungen werden geladen...</p>
        {:else}
          <form class="stack" on:submit|preventDefault={saveSettings}>
            <label>
              App Domain
              <input bind:value={settingsForm.appDomain} placeholder="machino.example.com" />
            </label>

            <div class="toggle-row">
              <div class="toggle-text">
                <span class="toggle-label">Registrierung erlauben</span>
                <span class="toggle-desc">Neue Benutzer können sich über die Login-Seite registrieren.</span>
              </div>
              <button
                type="button"
                class="toggle-btn"
                class:on={settingsForm.registrationEnabled}
                aria-label="Registrierung umschalten"
                aria-pressed={settingsForm.registrationEnabled}
                on:click={() => (settingsForm.registrationEnabled = !settingsForm.registrationEnabled)}
              >
                <span class="toggle-knob"></span>
              </button>
            </div>

            <div class="section-divider"></div>

            <div class="settings-grid">
              <label>
                SMTP Host
                <input bind:value={settingsForm.smtpHost} placeholder="smtp.example.com" />
              </label>
              <label>
                SMTP Port
                <input bind:value={settingsForm.smtpPort} inputmode="numeric" placeholder="587" />
              </label>
            </div>
            <label>
              SMTP Benutzer
              <input bind:value={settingsForm.smtpUsername} autocomplete="username" placeholder="mailer@example.com" />
            </label>
            <label>
              SMTP Passwort
              <input bind:value={settingsForm.smtpPassword} type="password" autocomplete="new-password" placeholder="Leer lassen, um bestehendes Passwort zu behalten" />
            </label>
            <p class="muted hint">
              {settings?.smtpPasswordSet ? 'SMTP-Passwort ist gesetzt. Leer speichern behält das bestehende Passwort.' : 'SMTP-Passwort ist noch nicht gesetzt.'}
            </p>
            <label>
              SMTP Absender
              <input bind:value={settingsForm.smtpFrom} type="email" placeholder="noreply@example.com" />
            </label>

            <button class="btn secondary" type="submit" disabled={saving}>Einstellungen speichern</button>
          </form>
        {/if}
      </div>
    </div>
  {/if}
</section>

<style>
  .admin-shell {
    display: grid;
    gap: 16px;
  }

  .admin-tabs {
    display: inline-flex;
    gap: 4px;
    margin: 28px 28px 0;
    padding: 4px;
    border-radius: 12px;
    background: var(--bg-2);
    border: 1px solid var(--border);
    width: fit-content;
  }

  .admin-tabs button {
    border: 0;
    border-radius: 9px;
    padding: 8px 18px;
    background: transparent;
    color: var(--text-muted);
    font-size: 0.875rem;
    font-weight: 600;
    transition: background 0.15s, color 0.15s;
  }

  .admin-tabs button:hover {
    background: var(--glass-hover);
    color: var(--text);
  }

  .admin-tabs button.active {
    background: rgba(99,102,241,0.2);
    color: #818cf8;
    box-shadow: 0 1px 4px rgba(99,102,241,0.18);
  }

  [data-theme="light"] .admin-tabs button.active {
    background: rgba(99,102,241,0.14);
    color: #4338ca;
  }

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

  .settings-page {
    padding: 0 28px 40px;
  }

  .settings-card {
    max-width: 720px;
  }

  .settings-grid {
    display: grid;
    grid-template-columns: 1fr 120px;
    gap: 12px;
  }

  .panel-header {
    display: flex;
    justify-content: space-between;
    gap: 16px;
    align-items: flex-start;
  }

  .panel-actions {
    display: flex;
    gap: 8px;
    align-items: center;
    flex-shrink: 0;
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

  :global([data-theme="light"]) .danger-btn {
    color: #b91c1c;
    background: rgba(220,38,38,0.08);
    border-color: rgba(220,38,38,0.25);
  }

  .btn.cancel {
    background: transparent;
    border: 1px solid var(--border);
    color: var(--text-muted);
    box-shadow: none;
  }
  .btn.cancel:hover {
    background: var(--glass);
    border-color: var(--border-hover);
    color: var(--text);
    transform: none;
    box-shadow: none;
  }

  .list-state {
    padding: 16px 0 0;
  }

  @media (max-width: 900px) {
    .admin-tabs {
      margin: 18px 16px 0;
    }

    .admin-page {
      grid-template-columns: 1fr;
      padding: 18px 16px 32px;
    }

    .settings-page {
      padding: 0 16px 32px;
    }

    .settings-grid {
      grid-template-columns: 1fr;
    }

    .user-list,
    .admin-form {
      padding: 14px 12px;
      border-radius: 14px;
    }
  }
</style>
