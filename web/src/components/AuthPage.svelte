<script>
  import { createEventDispatcher } from 'svelte'
  import { api } from '../lib/api.js'
  import { API } from '../lib/constants.js'

  const dispatch = createEventDispatcher()

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
  <section class="auth-visual" aria-hidden="true">
    <div class="visual-bg"></div>
    <div class="visual-grid"></div>
    <div class="visual-content">
      <div class="brand-mark">
        <span class="brand-dot"></span>
        <span class="brand-dot"></span>
        <span class="brand-dot"></span>
      </div>
      <h1 class="visual-title">
        <span class="word-mach">Mach</span>
        <span class="word-i">I</span>
        <span class="word-no">No</span>
      </h1>
      <p class="visual-sub">Deine Aufgaben. Immer dabei.</p>
      <div class="floating-cards">
        <div class="fc fc1"><span class="fc-check">✓</span> Präsentation vorbereiten</div>
        <div class="fc fc2"><span class="fc-dot high"></span> API deployen <span class="fc-date">heute</span></div>
        <div class="fc fc3"><span class="fc-dot"></span> Team Meeting planen</div>
        <div class="fc fc4"><span class="fc-check">✓</span> Design Review</div>
      </div>
    </div>
  </section>

  <section class="auth-form-panel">
    <div class="auth-form-inner">
      <div class="auth-logo">
        <svg width="36" height="36" viewBox="0 0 36 36" fill="none">
          <rect width="36" height="36" rx="10" fill="#4f46e5"/>
          <path d="M10 24 L18 12 L26 24" stroke="#fff" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" fill="none"/>
          <circle cx="18" cy="12" r="2.5" fill="#a5b4fc"/>
        </svg>
        <span>Mach I No</span>
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
            <button class="link-btn" on:click={() => { authMode = 'register'; error = ''; success = '' }}>
              Noch kein Account? <strong>Registrieren</strong>
            </button>
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
    display: grid;
    grid-template-columns: 1fr 480px;
    min-height: 100vh;
  }

  .auth-visual {
    position: relative;
    overflow: hidden;
    background: #0f0e17;
  }

  .visual-bg {
    position: absolute;
    inset: 0;
    background:
      radial-gradient(ellipse 80% 60% at 30% -10%, #4f46e5 0%, transparent 60%),
      radial-gradient(ellipse 60% 50% at 100% 80%, #7c3aed 0%, transparent 55%),
      radial-gradient(ellipse 50% 40% at 0% 100%, #0ea5e9 0%, transparent 50%);
  }

  .visual-grid {
    position: absolute;
    inset: 0;
    background-image:
      linear-gradient(rgba(255,255,255,0.04) 1px, transparent 1px),
      linear-gradient(90deg, rgba(255,255,255,0.04) 1px, transparent 1px);
    background-size: 48px 48px;
    mask-image: radial-gradient(ellipse at 40% 40%, black 30%, transparent 75%);
  }

  .visual-content {
    position: relative;
    z-index: 1;
    display: flex;
    flex-direction: column;
    justify-content: center;
    height: 100%;
    padding: 64px;
  }

  .brand-mark {
    display: flex;
    gap: 8px;
    margin-bottom: 36px;
  }

  .brand-dot {
    width: 10px;
    height: 10px;
    border-radius: 999px;
  }
  .brand-dot:nth-child(1) { background: #a5b4fc; }
  .brand-dot:nth-child(2) { background: #7c3aed; }
  .brand-dot:nth-child(3) { background: #38bdf8; }

  .visual-title {
    margin: 0 0 16px;
    font-size: clamp(3.5rem, 6vw, 5.5rem);
    font-weight: 950;
    letter-spacing: -0.06em;
    line-height: 0.95;
    display: flex;
    flex-direction: column;
  }

  .word-mach { color: #fff; }
  .word-i    { color: #a5b4fc; font-style: italic; margin-left: 0.12em; }
  .word-no   { color: #38bdf8; }

  .visual-sub {
    margin: 0 0 56px;
    color: rgba(255,255,255,0.5);
    font-size: 1.1rem;
    letter-spacing: 0.02em;
  }

  .floating-cards {
    display: flex;
    flex-direction: column;
    gap: 12px;
    max-width: 360px;
  }

  .fc {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 12px 16px;
    border-radius: 14px;
    background: rgba(255,255,255,0.07);
    border: 1px solid rgba(255,255,255,0.1);
    backdrop-filter: blur(12px);
    color: rgba(255,255,255,0.85);
    font-size: 0.88rem;
    font-weight: 600;
    transform: translateX(0);
    animation: float-in 0.6s ease both;
  }

  .fc1 { animation-delay: 0.1s; transform: translateX(-8px); opacity: 0.7; }
  .fc2 { animation-delay: 0.2s; }
  .fc3 { animation-delay: 0.3s; opacity: 0.6; transform: translateX(12px); }
  .fc4 { animation-delay: 0.4s; opacity: 0.45; transform: translateX(-4px); }

  @keyframes float-in {
    from { opacity: 0; transform: translateY(16px); }
    to   { opacity: var(--fc-opacity, 1); transform: var(--fc-tx, translateX(0)); }
  }

  .fc-check {
    width: 22px;
    height: 22px;
    border-radius: 999px;
    background: rgba(165,180,252,0.2);
    color: #a5b4fc;
    font-size: 0.72rem;
    display: grid;
    place-items: center;
    flex-shrink: 0;
  }

  .fc-dot {
    width: 8px;
    height: 8px;
    border-radius: 999px;
    background: rgba(255,255,255,0.3);
    flex-shrink: 0;
  }
  .fc-dot.high { background: #f87171; }

  .fc-date {
    margin-left: auto;
    font-size: 0.74rem;
    color: rgba(255,255,255,0.4);
  }

  .auth-form-panel {
    display: flex;
    align-items: center;
    justify-content: center;
    background: #fff;
    padding: 48px 40px;
  }

  .auth-form-inner {
    width: 100%;
    max-width: 360px;
    display: grid;
    gap: 20px;
  }

  .auth-logo {
    display: flex;
    align-items: center;
    gap: 12px;
    font-size: 1.25rem;
    font-weight: 900;
    color: #111827;
    letter-spacing: -0.04em;
    margin-bottom: 8px;
  }

  .form-header h2 {
    margin: 0 0 4px;
    font-size: 1.6rem;
    font-weight: 900;
    color: #111827;
    letter-spacing: -0.04em;
  }

  .form-header p {
    margin: 0;
    color: #6b7280;
    font-size: 0.9rem;
  }

  .auth-footer {
    display: flex;
    flex-direction: column;
    gap: 6px;
    align-items: center;
  }

  .link-btn {
    background: none;
    border: none;
    cursor: pointer;
    color: #4f46e5;
    font-size: 0.88rem;
    font-weight: 600;
    padding: 4px 0;
    text-align: center;
  }
  .link-btn:hover { text-decoration: underline; }
  .link-btn.muted { color: #9ca3af; font-weight: 500; }
  .link-btn.muted:hover { color: #6b7280; }

  .demo-token code {
    display: block;
    margin-top: 6px;
    padding: 6px 10px;
    background: #f3f4f6;
    border-radius: 8px;
    font-size: 0.8rem;
    word-break: break-all;
  }

  form {
    display: grid;
    gap: 14px;
  }

  @media (max-width: 900px) {
    .auth-layout {
      grid-template-columns: 1fr;
    }
    .auth-visual {
      min-height: 280px;
    }
    .visual-content {
      padding: 40px 32px;
      justify-content: flex-end;
    }
    .floating-cards { display: none; }
  }
</style>
