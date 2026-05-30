<script>
  import { createEventDispatcher } from 'svelte'
  import { api } from '../lib/api.js'
  import { API } from '../lib/constants.js'
  import { theme } from '../lib/stores.js'

  export let defaults = {}

  const dispatch = createEventDispatcher()

  let step = 1
  let saving = false
  let error = ''
  let adminForm = { name: '', email: '', password: '' }
  let settingsForm = {
    appDomain: defaults.appDomain || location.hostname,
    registrationEnabled: defaults.registrationEnabled ?? true,
    smtpHost: defaults.smtpHost || '',
    smtpPort: defaults.smtpPort || '587',
    smtpUsername: defaults.smtpUsername || '',
    smtpPassword: '',
    smtpFrom: defaults.smtpFrom || '',
  }
  $: passwordHint = defaults.smtpPasswordSet ? 'SMTP-Passwort ist bereits aus der Umgebung gesetzt. Leer lassen, um es zu behalten.' : 'Optional, falls Passwort-Reset per E-Mail aktiv sein soll.'

  function nextStep() {
    error = ''
    if (!adminForm.name.trim() || !adminForm.email.trim() || !adminForm.password) {
      error = 'Bitte Name, E-Mail und Passwort ausfüllen.'
      return
    }
    if (adminForm.password.length < 8) {
      error = 'Passwort muss mindestens 8 Zeichen lang sein.'
      return
    }
    step = 2
  }

  async function completeSetup() {
    saving = true
    error = ''
    try {
      const payload = await api(API.setupComplete, {
        method: 'POST',
        body: {
          ...adminForm,
          ...settingsForm,
        },
      })
      dispatch('authenticated', { user: payload.user })
    } catch (err) {
      error = err.message
    } finally {
      saving = false
    }
  }
</script>

<main class="setup-layout">
  <section class="setup-visual" aria-hidden="true">
    <div class="visual-bg"></div>
    <div class="visual-grid"></div>
    <div class="visual-content">
      <img class="visual-logo" src={$theme === 'light' ? '/logo-white.png' : '/logo-dark.png'} alt="" />
      <div>
        <p class="eyebrow">Ersteinrichtung</p>
        <h1 class="visual-title">
          <span>Mach</span>
          <span class="accent">I</span>
          <span>No</span>
        </h1>
      </div>
      <p class="visual-sub">Lege den ersten Admin an und mache die App startklar.</p>
      <div class="setup-steps">
        <span class:active={step === 1}>1 Admin</span>
        <span class:active={step === 2}>2 Einstellungen</span>
      </div>
    </div>
  </section>

  <section class="setup-form-panel">
    <div class="setup-form-inner">
      <div class="auth-logo">
        <img class="auth-logo-img" src={$theme === 'light' ? '/logo-white.png' : '/logo-dark.png'} alt="Machino" />
      </div>

      <div class="form-header">
        <p class="step-label">Schritt {step} von 2</p>
        <h2>{step === 1 ? 'Admin erstellen' : 'Globale Einstellungen'}</h2>
        <p>{step === 1 ? 'Dieser Benutzer erhält automatisch Admin-Rechte.' : 'Diese Werte kannst du später im Adminbereich ändern.'}</p>
      </div>

      {#if error}<p class="error">{error}</p>{/if}

      {#if step === 1}
        <form on:submit|preventDefault={nextStep}>
          <label>Name <input bind:value={adminForm.name} autocomplete="name" placeholder="Admin Name" required /></label>
          <label>E-Mail <input bind:value={adminForm.email} type="email" autocomplete="email" placeholder="admin@example.com" required /></label>
          <label>Passwort <input bind:value={adminForm.password} type="password" autocomplete="new-password" placeholder="Mindestens 8 Zeichen" required /></label>
          <button class="btn" type="submit">Weiter zu Einstellungen</button>
        </form>
      {:else}
        <form on:submit|preventDefault={completeSetup}>
          <label>App Domain <input bind:value={settingsForm.appDomain} placeholder="machino.example.com" /></label>

          <div class="toggle-row">
            <div class="toggle-text">
              <span class="toggle-label">Registrierung erlauben</span>
              <span class="toggle-desc">Neue Benutzer können sich selbst registrieren.</span>
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

          <div class="smtp-grid">
            <label>SMTP Host <input bind:value={settingsForm.smtpHost} placeholder="smtp.example.com" /></label>
            <label>SMTP Port <input bind:value={settingsForm.smtpPort} inputmode="numeric" placeholder="587" /></label>
          </div>
          <label>SMTP Benutzer <input bind:value={settingsForm.smtpUsername} autocomplete="username" placeholder="mailer@example.com" /></label>
          <label>SMTP Passwort <input bind:value={settingsForm.smtpPassword} type="password" autocomplete="new-password" placeholder="Leer lassen, falls nicht benötigt" /></label>
          <p class="hint">{passwordHint}</p>
          <label>SMTP Absender <input bind:value={settingsForm.smtpFrom} type="email" placeholder="noreply@example.com" /></label>

          <div class="actions">
            <button class="link-btn" type="button" on:click={() => { step = 1; error = '' }}>Zurück</button>
            <button class="btn" type="submit" disabled={saving}>{saving ? 'Speichern...' : 'Setup abschließen'}</button>
          </div>
        </form>
      {/if}
    </div>
  </section>
</main>

<style>
  .setup-layout {
    display: grid;
    grid-template-columns: 1fr 500px;
    min-height: 100vh;
    background: var(--bg);
  }

  .setup-visual {
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

  .eyebrow,
  .step-label {
    margin: 0 0 8px;
    color: #818cf8;
    font-size: 0.78rem;
    font-weight: 900;
    letter-spacing: 0.12em;
    text-transform: uppercase;
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

  .visual-title span { color: var(--text); }
  .visual-title .accent { color: #818cf8; font-style: italic; margin-left: 0.12em; }

  .visual-sub {
    margin: 0;
    color: var(--text-muted);
    font-size: 1rem;
    font-weight: 400;
    max-width: 380px;
  }

  .setup-steps {
    display: flex;
    gap: 10px;
    flex-wrap: wrap;
  }

  .setup-steps span {
    border-radius: 999px;
    padding: 8px 12px;
    background: rgba(255,255,255,0.04);
    border: 1px solid rgba(255,255,255,0.08);
    color: var(--text-muted);
    font-size: 0.78rem;
    font-weight: 800;
  }

  .setup-steps span.active {
    background: rgba(99,102,241,0.18);
    border-color: rgba(99,102,241,0.36);
    color: #c7d2fe;
  }

  .setup-form-panel {
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--bg-3);
    border-left: 1px solid var(--border);
    padding: 48px 44px;
  }

  .setup-form-inner {
    width: 100%;
    max-width: 390px;
    display: grid;
    gap: 22px;
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

  .form-header p:not(.step-label) {
    margin: 0;
    color: var(--text-muted);
    font-size: 0.88rem;
  }

  form {
    display: grid;
    gap: 14px;
  }

  .smtp-grid {
    display: grid;
    grid-template-columns: 1fr 100px;
    gap: 10px;
  }

  .hint {
    margin: -4px 0 0;
    color: var(--text-muted);
    font-size: 0.78rem;
    line-height: 1.45;
  }

  .actions {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 14px;
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
  }

  .link-btn:hover { color: #a5b4fc; }

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

  @media (max-width: 900px) {
    .setup-layout { grid-template-columns: 1fr; }
    .setup-visual { min-height: 260px; }
    .visual-content { padding: 40px 32px; justify-content: flex-end; }
    .setup-form-panel { border-left: none; border-top: 1px solid var(--border); }
    .smtp-grid { grid-template-columns: 1fr; }
  }
</style>
