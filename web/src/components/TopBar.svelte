<script>
  import { createEventDispatcher } from 'svelte'
  import {
    selectedProject,
    currentView,
    user,
    avatarInitial,
    openTodoCount,
  } from '../lib/stores.js'

  const dispatch = createEventDispatcher()

  export let menuOpen = false
</script>

<header class="topbar card">
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
  <div class="actions">
    <span>{$currentView === 'profile' ? $user.email : `${$openTodoCount} offen`}</span>
    <div class="profile-menu">
      <button class="avatar" aria-label="Profilmenü" on:click|stopPropagation={() => dispatch('toggle-menu')}>
        {$avatarInitial}
      </button>
      {#if menuOpen}
        <div class="dropdown card" on:click|stopPropagation role="menu" aria-label="Profilmenü">
          <button on:click={() => dispatch('open-todos')}>Todos</button>
          <button on:click={() => dispatch('open-profile')}>Profil</button>
          <button on:click={() => dispatch('logout')}>Logout</button>
        </div>
      {/if}
    </div>
  </div>
</header>

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

  .dropdown {
    position: absolute;
    top: calc(100% + 8px);
    right: 0;
    min-width: 164px;
    padding: 5px;
    display: grid;
    gap: 2px;
    background: var(--bg-2);
    border-color: var(--border-hover);
    box-shadow: 0 24px 64px rgba(0,0,0,0.7);
    backdrop-filter: blur(24px);
    animation: drop-in 0.13s cubic-bezier(0.16,1,0.3,1);
    z-index: 100;
  }

  @keyframes drop-in {
    from { opacity: 0; transform: translateY(-8px) scale(0.96); }
    to   { opacity: 1; transform: translateY(0)  scale(1); }
  }

  .dropdown button {
    text-align: left;
    padding: 9px 12px;
    border-radius: 10px;
    background: transparent;
    color: var(--text);
    font-size: 0.88rem;
    font-weight: 600;
    transition: background 0.12s;
  }
  .dropdown button:hover { background: var(--glass-hover); }
  .dropdown button:last-child { color: #f87171; }
  .dropdown button:last-child:hover { background: rgba(248,113,113,0.12); }
</style>
