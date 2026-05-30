<script>
  import { createEventDispatcher } from 'svelte'
  import { api } from '../lib/api.js'
  import { API } from '../lib/constants.js'
  import { theme } from '../lib/stores.js'

  const dispatch = createEventDispatcher()

  export let registrationEnabled = true

  let authMode = 'login'
  let loginForm = { email: '', password: '', name: '' }
  let resetForm = { email: '', token: '', password: '' }
  let resetToken = ''      // set when API returns token (demo mode, no SMTP)
  let resetSent = false    // set when email was dispatched via SMTP
  let error = ''
  let success = ''

  async function submitAuth() {
    error = ''
    const endpoint = authMode === 'register' ? API.register : API.login
    const body =
      authMode === 'register' ? loginForm : { email: loginForm.email, password: loginForm.password }
    try {
      const payload = await api(endpoint, { method: 'POST', body })
      loginForm = { email: '', password: '', name: '' }
      dispatch('authenticated', { user: payload.user })
    } catch (err) {
      error = err.message
    }
  }

  async function requestReset() {
    error = ''
    resetToken = ''
    resetSent = false
    try {
      const payload = await api(API.passwordResetRequest, {
        method: 'POST',
        body: { email: resetForm.email },
      })
      if (payload.resetToken) {
        // Demo-Modus: kein SMTP konfiguriert
        resetToken = payload.resetToken
        resetForm.token = payload.resetToken
      } else {
        // SMTP konfiguriert: Token wurde per E-Mail gesendet
        resetSent = true
      }
    } catch (err) {
      error = err.message
    }
  }

  async function confirmReset() {
    error = ''
    success = ''
    try {
      await api(API.passwordResetConfirm, {
        method: 'POST',
        body: { token: resetForm.token, password: resetForm.password },
      })
      resetForm = { email: '', token: '', password: '' }
      resetToken = ''
      resetSent = false
      authMode = 'login'
      success = 'Passwort wurde zurückgesetzt.'
    } catch (err) {
      error = err.message
    }
  }
</script>

<main class="auth-layout">
  <section class="auth-form-panel">
    <div class="auth-form-inner">
      <div class="auth-logo">
        <img class="auth-logo-img" src={$theme === 'light' ? '/logo-white.png' : '/logo-dark.png'} alt="Machino" />
      </div>

      {#if authMode === 'reset'}
        <div class="form-header">
          <h2>Passwort zurücksetzen</h2>
          <p>Wir senden dir einen Reset-Link.</p>
        </div>
        {#if error}<p class="error">{error}</p>{/if}
        {#if success}<p class="success">{success}</p>{/if}
        <form on:submit|preventDefault={requestReset}>
          <label>E-Mail <input bind:value={resetForm.email} type="email" placeholder="du@beispiel.ch" required /></label>
          <button class="btn" type="submit">Reset-Token anfordern</button>
        </form>
        {#if resetSent}
          <p class="success">Prüfe deine E-Mails — wir haben dir den Reset-Token gesendet.</p>
          <form on:submit|preventDefault={confirmReset}>
            <label>Token <input bind:value={resetForm.token} placeholder="Token aus E-Mail einfügen" required /></label>
            <label>Neues Passwort <input bind:value={resetForm.password} type="password" placeholder="Neues Passwort" required /></label>
            <button class="btn" type="submit">Passwort setzen</button>
          </form>
        {:else if resetToken}
          <p class="success demo-token">Demo-Modus (kein SMTP): <code>{resetToken}</code></p>
          <form on:submit|preventDefault={confirmReset}>
            <label>Token <input bind:value={resetForm.token} placeholder="Token einfügen" required /></label>
            <label>Neues Passwort <input bind:value={resetForm.password} type="password" placeholder="Neues Passwort" required /></label>
            <button class="btn" type="submit">Passwort setzen</button>
          </form>
        {/if}
        <button class="link-btn" on:click={() => { authMode = 'login'; error = ''; success = '' }}>
          ← Zurück zum Login
        </button>

      {:else}
        <div class="form-header">
          <h2>{authMode === 'register' ? 'Account erstellen' : 'Willkommen zurück'}</h2>
          <p>{authMode === 'register' ? 'Starte kostenlos durch.' : 'Schön, dass du wieder da bist.'}</p>
        </div>
        {#if error}<p class="error">{error}</p>{/if}
        {#if success}<p class="success">{success}</p>{/if}
        <form on:submit|preventDefault={submitAuth}>
          {#if authMode === 'register'}
            <label>Name <input bind:value={loginForm.name} placeholder="Dein Name" autocomplete="name" /></label>
          {/if}
          <label>E-Mail <input bind:value={loginForm.email} type="email" placeholder="du@beispiel.ch" autocomplete="email" required /></label>
          <label>
            Passwort
            <input bind:value={loginForm.password} type="password" placeholder="••••••••" autocomplete="current-password" required />
          </label>
          <button class="btn" type="submit">
            {authMode === 'register' ? 'Account erstellen' : 'Einloggen'}
          </button>
        </form>

        <div class="auth-footer">
          {#if authMode === 'login'}
            {#if registrationEnabled}
              <button class="link-btn" on:click={() => { authMode = 'register'; error = ''; success = '' }}>
                Noch kein Account? <strong>Registrieren</strong>
              </button>
            {/if}
            <button class="link-btn muted" on:click={() => { authMode = 'reset'; error = ''; success = '' }}>
              Passwort vergessen?
            </button>
          {:else}
            <button class="link-btn" on:click={() => { authMode = 'login'; error = ''; success = '' }}>
              Bereits registriert? <strong>Einloggen</strong>
            </button>
          {/if}
        </div>
      {/if}
    </div>
  </section>
</main>

<style>
  .auth-layout {
    display: flex;
    min-height: 100vh;
    align-items: center;
    justify-content: center;
    background: var(--bg);
    padding: 32px 16px;
  }

  /* === Form Side === */
  .auth-form-panel {
    width: 100%;
    max-width: 420px;
    background: var(--bg-3);
    border: 1px solid var(--border);
    border-radius: 20px;
    padding: 48px 44px;
  }

  .auth-form-inner {
    width: 100%;
    display: grid;
    gap: 24px;
  }

  .auth-logo {
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .auth-logo-img {
    width: 180px;
    height: auto;
  }

  .form-header h2 {
    margin: 0 0 4px;
    font-size: 1.45rem;
    font-weight: 800;
    color: var(--text);
    letter-spacing: -0.03em;
  }
  .form-header p {
    margin: 0;
    color: var(--text-muted);
    font-size: 0.88rem;
  }

  form { display: grid; gap: 14px; }

  .auth-footer {
    display: flex;
    flex-direction: column;
    gap: 8px;
    align-items: center;
  }

  .link-btn {
    background: none;
    border: none;
    cursor: pointer;
    color: #818cf8;
    font-size: 0.86rem;
    font-weight: 600;
    padding: 4px 0;
    text-align: center;
    transition: color 0.15s;
  }
  .link-btn:hover { color: #a5b4fc; }

  .link-btn.muted { color: var(--text-faint); font-weight: 500; }
  .link-btn.muted:hover { color: var(--text-muted); }

  .demo-token code {
    display: block;
    margin-top: 6px;
    padding: 8px 12px;
    background: var(--glass);
    border: 1px solid var(--border);
    border-radius: 10px;
    font-size: 0.76rem;
    word-break: break-all;
    color: var(--text);
    font-family: 'SF Mono', 'Fira Code', 'Consolas', monospace;
  }

  @media (max-width: 500px) {
    .auth-form-panel { padding: 36px 24px; }
  }
</style>
