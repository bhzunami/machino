<script>
  import { api } from '../lib/api.js'
  import { user, error, success } from '../lib/stores.js'

  let profileForm = {
    email: $user?.email || '',
    name: $user?.name || '',
    searchable: $user?.searchable ?? true,
    password: '',
  }

  $: if ($user) {
    profileForm = {
      email: $user.email,
      name: $user.name || '',
      searchable: $user.searchable ?? true,
      password: '',
    }
  }

  async function updateProfile() {
    error.set('')
    success.set('')
    try {
      const payload = await api('/api/profile', {
        method: 'PUT',
        body: { email: profileForm.email, name: profileForm.name, searchable: profileForm.searchable },
      })
      user.set(payload.user)
      profileForm = { ...profileForm, email: payload.user.email, name: payload.user.name || '', password: '' }
      success.set('Profil gespeichert.')
    } catch (err) {
      error.set(err.message)
    }
  }

  async function updatePassword() {
    error.set('')
    success.set('')
    try {
      await api('/api/profile/password', { method: 'PUT', body: { password: profileForm.password } })
      profileForm = { ...profileForm, password: '' }
      success.set('Passwort gespeichert.')
    } catch (err) {
      error.set(err.message)
    }
  }
</script>

<section class="card profile-page stack">
  <h2>Profil</h2>
  <label>
    Name
    <input bind:value={profileForm.name} autocomplete="name" />
  </label>
  <label>
    E-Mail
    <input bind:value={profileForm.email} type="email" autocomplete="email" />
  </label>

  <div class="toggle-row">
    <div class="toggle-text">
      <span class="toggle-label">In Projektsuche auffindbar</span>
      <span class="toggle-desc">Andere Nutzer können dich beim Teilen von Projekten über Name oder E-Mail finden.</span>
    </div>
    <button
      type="button"
      class="toggle-btn"
      class:on={profileForm.searchable}
      aria-pressed={profileForm.searchable}
      on:click={() => (profileForm.searchable = !profileForm.searchable)}
    >
      <span class="toggle-knob"></span>
    </button>
  </div>

  <button class="btn secondary" on:click={updateProfile}>Profil speichern</button>

  <div class="section-divider"></div>

  <label>
    Neues Passwort
    <input bind:value={profileForm.password} type="password" autocomplete="new-password" placeholder="Mindestens 8 Zeichen…" />
  </label>
  <button class="btn secondary" on:click={updatePassword}>Passwort ändern</button>
</section>

<style>
  .profile-page {
    max-width: 580px;
    padding: 20px 24px;
    background: var(--glass);
    border-color: var(--border);
    width: 100%;
    box-sizing: border-box;
  }

  @media (max-width: 600px) {
    .profile-page {
      padding: 14px 12px;
      border-radius: 14px;
    }
  }

  h2 {
    margin: 0 0 4px;
    font-size: 1.15rem;
    font-weight: 800;
    color: var(--text);
    letter-spacing: -0.02em;
  }

  .section-divider {
    height: 1px;
    background: var(--border);
    margin: 4px 0;
  }

  .stack {
    display: grid;
    gap: 14px;
  }

  /* Toggle switch */
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
</style>
