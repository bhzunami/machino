<script>
  import { createEventDispatcher, tick } from 'svelte'
  import {
    selectedProject,
    currentView,
    user,
    avatarInitial,
    openTodoCount,
    theme,
  } from '../lib/stores.js'

  const dispatch = createEventDispatcher()

  export let menuOpen = false
  export let sidebarOpen = false

  let avatarEl
  let dropdownPos = { top: 0, right: 0 }

  async function toggleMenu() {
    if (!menuOpen) {
      const rect = avatarEl.getBoundingClientRect()
      dropdownPos = { top: rect.bottom + 8, right: window.innerWidth - rect.right }
    }
    dispatch('toggle-menu')
  }
</script>

<header class="topbar card">
  <div class="topbar-left">
    <button class="hamburger" aria-label="Menü öffnen" on:click|stopPropagation={() => dispatch('toggle-sidebar')}>
      <span class="bar" class:open={sidebarOpen}></span>
      <span class="bar" class:open={sidebarOpen}></span>
      <span class="bar" class:open={sidebarOpen}></span>
    </button>
    <div class="topbar-info">
      <div class="topbar-title">
        <span class="topbar-badge">{$currentView === 'profile' ? 'Konto' : 'Projekt'}</span>
        <h2>{$currentView === 'profile' ? 'Profil' : $selectedProject?.title || 'Noch kein Projekt'}</h2>
      </div>
      {#if ($currentView !== 'profile' && $selectedProject?.description) || $currentView === 'profile'}
        <p class="topbar-desc muted">
          {$currentView === 'profile' ? 'E-Mail und Passwort verwalten.' : $selectedProject.description}
        </p>
      {/if}
    </div>
  </div>
  <div class="actions">
    <span>{$currentView === 'profile' ? $user.email : `${$openTodoCount} offen`}</span>
    <button
      class="theme-btn"
      on:click={theme.toggle}
      aria-label="Farbschema wechseln"
      title={$theme === 'dark' ? 'Light Mode aktivieren' : 'Dark Mode aktivieren'}
    >
      {$theme === 'dark' ? '☀' : '🌙'}
    </button>
    <div class="profile-menu">
      <button class="avatar" bind:this={avatarEl} aria-label="Profilmenü" on:click|stopPropagation={toggleMenu}>
        {$avatarInitial}
      </button>
    </div>
  </div>
</header>

{#if menuOpen}
  <!-- svelte-ignore a11y-no-static-element-interactions -->
  <div
    class="dropdown-portal"
    style="top:{dropdownPos.top}px;right:{dropdownPos.right}px"
    on:click|stopPropagation
    role="menu"
    aria-label="Profilmenü"
  >
    <button on:click={() => dispatch('open-todos')}>Todos</button>
    <button on:click={() => dispatch('open-profile')}>Profil</button>
    <button on:click={() => dispatch('logout')}>Logout</button>
  </div>
{/if}

<style>
  .topbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    position: relative;
    overflow: visible;
    background: var(--glass);
    border-color: var(--border);
    backdrop-filter: blur(24px);
    -webkit-backdrop-filter: blur(24px);
    padding: 12px 18px;
  }

  .topbar-left {
    display: flex;
    align-items: center;
    gap: 10px;
    min-width: 0;
    flex: 1;
  }

  /* Hamburger — nur auf Mobile sichtbar */
  .hamburger {
    display: none;
    flex-direction: column;
    justify-content: center;
    gap: 5px;
    width: 32px;
    height: 32px;
    background: transparent;
    border: none;
    padding: 4px;
    flex-shrink: 0;
    cursor: pointer;
  }
  .bar {
    display: block;
    width: 20px;
    height: 2px;
    background: var(--text-muted);
    border-radius: 2px;
    transition: transform 0.2s, opacity 0.2s, width 0.2s;
  }
  .bar.open:nth-child(1) { transform: translateY(7px) rotate(45deg); }
  .bar.open:nth-child(2) { opacity: 0; width: 0; }
  .bar.open:nth-child(3) { transform: translateY(-7px) rotate(-45deg); }

  @media (max-width: 900px) {
    .hamburger { display: flex; }
  }

  .topbar::after {
    content: '';
    position: absolute;
    bottom: -1px;
    left: 20px;
    right: 20px;
    height: 1px;
    background: linear-gradient(90deg, transparent, color-mix(in srgb, var(--accent-color), transparent 50%), transparent);
    pointer-events: none;
  }

  .topbar-info {
    display: flex;
    flex-direction: column;
    gap: 2px;
    min-width: 0;
  }

  .topbar-title {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .topbar-title h2 {
    margin: 0;
    font-size: 1rem;
    font-weight: 700;
    color: var(--text);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .topbar-badge {
    flex-shrink: 0;
    padding: 2px 10px;
    border-radius: 999px;
    background: color-mix(in srgb, var(--accent-color), transparent 80%);
    color: color-mix(in srgb, var(--accent-color), white 35%);
    font-size: 0.66rem;
    font-weight: 800;
    letter-spacing: 0.09em;
    text-transform: uppercase;
    border: 1px solid color-mix(in srgb, var(--accent-color), transparent 62%);
  }

  .topbar-desc {
    margin: 0;
    font-size: 0.8rem;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    color: var(--text-muted);
  }

  .actions {
    display: flex;
    align-items: center;
    gap: 12px;
    flex-shrink: 0;
  }

  .actions > span {
    font-size: 0.8rem;
    font-weight: 600;
    color: var(--text-faint);
    white-space: nowrap;
  }

  .theme-btn {
    width: 32px;
    height: 32px;
    border-radius: 10px;
    background: var(--glass-md);
    border: 1px solid var(--border);
    color: var(--text-muted);
    font-size: 1rem;
    display: grid;
    place-items: center;
    transition: background 0.15s, border-color 0.15s, color 0.15s, transform 0.15s;
    flex-shrink: 0;
  }
  .theme-btn:hover {
    background: var(--glass-hover);
    border-color: var(--border-hover);
    color: var(--text);
    transform: scale(1.1);
  }

  .profile-menu { position: relative; }

  .avatar {
    width: 36px;
    height: 36px;
    border-radius: 999px;
    background: linear-gradient(135deg, var(--accent-color), color-mix(in srgb, var(--accent-color), #7c3aed 55%));
    color: #fff;
    font-size: 0.82rem;
    font-weight: 800;
    display: grid;
    place-items: center;
    box-shadow: 0 0 0 1px rgba(255,255,255,0.1), 0 0 18px color-mix(in srgb, var(--accent-color), transparent 58%);
    transition: box-shadow 0.2s, transform 0.15s;
  }
  .avatar:hover {
    box-shadow: 0 0 0 2px color-mix(in srgb, var(--accent-color), transparent 40%), 0 0 28px color-mix(in srgb, var(--accent-color), transparent 38%);
    transform: scale(1.06);
  }

  .dropdown-portal {
    position: fixed;
    min-width: 164px;
    padding: 5px;
    display: grid;
    gap: 2px;
    background: var(--bg-2);
    border: 1px solid var(--border-hover);
    border-radius: 14px;
    box-shadow: 0 24px 64px var(--shadow-dropdown), 0 0 0 1px rgba(255,255,255,0.04);
    backdrop-filter: blur(24px);
    animation: drop-in 0.13s cubic-bezier(0.16,1,0.3,1);
    z-index: 9999;
  }

  @keyframes drop-in {
    from { opacity: 0; transform: translateY(-8px) scale(0.96); }
    to   { opacity: 1; transform: translateY(0)  scale(1); }
  }

  .dropdown-portal button {
    text-align: left;
    padding: 9px 12px;
    border-radius: 10px;
    background: transparent;
    color: var(--text);
    font-size: 0.88rem;
    font-weight: 600;
    transition: background 0.12s;
  }
  .dropdown-portal button:hover { background: var(--glass-hover); }
  .dropdown-portal button:last-child { color: #f87171; }
  .dropdown-portal button:last-child:hover { background: rgba(248,113,113,0.12); }
</style>
