<script>
  import { createEventDispatcher } from 'svelte'
  import { api } from '../lib/api.js'
  import { API } from '../lib/constants.js'
  import { theme } from '../lib/stores.js'

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
      <img class="visual-logo" src={$theme === 'light' ? '/logo-white.png' : '/logo-dark.png'} alt="" />
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
    grid-template-columns: 1fr 460px;
    min-height: 100vh;
    background: var(--bg);
  }

  /* === Visual Side === */
  .auth-visual {
    position: relative;
    overflow: hidden;
    background: var(--bg-2);
  }

  .visual-bg {
    position: absolute;
    inset: 0;
    background:
      radial-gradient(ellipse at 20% 50%, rgba(99,102,241,0.28) 0%, transparent 58%),
      radial-gradient(ellipse at 78% 12%, rgba(139,92,246,0.22) 0%, transparent 48%),
      radial-gradient(ellipse at 60% 82%, rgba(56,189,248,0.14) 0%, transparent 48%);
    animation: aurora 9s ease-in-out infinite alternate;
  }

  @keyframes aurora {
    0%   { opacity: 1;   transform: scale(1)    rotate(0deg); }
    100% { opacity: 0.75; transform: scale(1.06) rotate(2deg); }
  }

  .visual-grid {
    position: absolute;
    inset: 0;
    background-image:
      linear-gradient(rgba(255,255,255,0.025) 1px, transparent 1px),
      linear-gradient(90deg, rgba(255,255,255,0.025) 1px, transparent 1px);
    background-size: 48px 48px;
  }

  .visual-content {
    position: relative;
    z-index: 1;
    height: 100%;
    display: flex;
    flex-direction: column;
    justify-content: center;
    padding: 64px 60px;
    gap: 28px;
  }

  .visual-logo {
    width: min(260px, 70%);
    height: auto;
  }

  .visual-title {
    margin: 0;
    font-size: clamp(3.2rem, 6vw, 5.5rem);
    font-weight: 900;
    letter-spacing: -0.04em;
    line-height: 0.92;
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .word-mach { color: var(--text); }
  .word-i    { color: #818cf8; font-style: italic; margin-left: 0.12em; }
  .word-no   { color: var(--text-faint); }

  .visual-sub {
    margin: 0;
    color: var(--text-muted);
    font-size: 1rem;
    font-weight: 400;
  }

  .floating-cards {
    display: flex;
    flex-direction: column;
    gap: 9px;
    margin-top: 16px;
  }

  .fc {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 11px 15px;
    border-radius: 14px;
    background: rgba(255,255,255,0.04);
    border: 1px solid rgba(255,255,255,0.08);
    backdrop-filter: blur(12px);
    color: var(--text-muted);
    font-size: 0.86rem;
    font-weight: 500;
    animation: float-in 0.55s cubic-bezier(0.16,1,0.3,1) both;
  }
  .fc1 { animation-delay: 0.15s; }
  .fc2 { animation-delay: 0.28s; }
  .fc3 { animation-delay: 0.41s; }
  .fc4 { animation-delay: 0.54s; }

  @keyframes float-in {
    from { opacity: 0; transform: translateX(-18px); }
    to   { opacity: 1; transform: translateX(0); }
  }

  .fc-check {
    width: 22px;
    height: 22px;
    border-radius: 7px;
    background: linear-gradient(135deg, #4ade80, #16a34a);
    display: grid;
    place-items: center;
    font-size: 0.68rem;
    color: #fff;
    font-weight: 900;
    flex-shrink: 0;
    box-shadow: 0 0 10px rgba(74,222,128,0.4);
  }

  .fc-dot {
    width: 8px;
    height: 8px;
    border-radius: 999px;
    background: var(--text-faint);
    flex-shrink: 0;
  }
  .fc-dot.high { background: #f87171; box-shadow: 0 0 8px rgba(248,113,113,0.6); }

  .fc-date {
    margin-left: auto;
    font-size: 0.72rem;
    color: var(--text-faint);
  }

  /* === Form Side === */
  .auth-form-panel {
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--bg-3);
    border-left: 1px solid var(--border);
    padding: 48px 44px;
  }

  .auth-form-inner {
    width: 100%;
    max-width: 350px;
    display: grid;
    gap: 24px;
  }

  .auth-logo {
    display: flex;
    align-items: center;
  }

  .auth-logo-img {
    width: 124px;
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

  @media (max-width: 900px) {
    .auth-layout { grid-template-columns: 1fr; }
    .auth-visual { min-height: 260px; }
    .visual-content { padding: 40px 32px; justify-content: flex-end; }
    .floating-cards { display: none; }
    .auth-form-panel { border-left: none; border-top: 1px solid var(--border); }
  }
</style>
