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
      <button class="avatar" aria-label="Profilmenü" on:click={() => dispatch('toggle-menu')}>
        {$avatarInitial}
      </button>
      {#if menuOpen}
        <div class="dropdown card">
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
    border-color: rgba(226, 232, 240, 0.9);
    background:
      linear-gradient(120deg, color-mix(in srgb, var(--accent-color), white 92%), rgba(255, 255, 255, 0.98) 48%),
      #fff;
    padding: 14px 20px;
  }

  .topbar::before {
    content: '';
    position: absolute;
    inset: 0 auto 0 0;
    width: 5px;
    border-radius: 24px 0 0 24px;
    background: var(--accent-color);
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
    font-size: 1.05rem;
    font-weight: 750;
    color: #111827;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .topbar-badge {
    flex-shrink: 0;
    border-radius: 6px;
    padding: 2px 7px;
    background: color-mix(in srgb, var(--accent-color), white 86%);
    color: color-mix(in srgb, var(--accent-color), #111827 50%);
    font-size: 0.68rem;
    font-weight: 800;
    letter-spacing: 0.06em;
    text-transform: uppercase;
  }

  .topbar-desc {
    font-size: 0.8rem;
    margin: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .actions {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .actions span {
    border-radius: 999px;
    padding: 4px 10px;
    background: color-mix(in srgb, var(--accent-color), white 88%);
    color: color-mix(in srgb, var(--accent-color), #111827 60%);
    font-weight: 700;
    font-size: 0.78rem;
  }

  .profile-menu {
    position: relative;
  }

  .avatar {
    width: 38px;
    height: 38px;
    border-radius: 999px;
    background: linear-gradient(135deg, var(--accent-color), #111827);
    color: #fff;
    font-weight: 900;
    box-shadow: 0 4px 12px color-mix(in srgb, var(--accent-color), transparent 78%);
  }

  .dropdown {
    position: absolute;
    top: calc(100% + 10px);
    right: 0;
    z-index: 10;
    display: grid;
    min-width: 170px;
    overflow: hidden;
    padding: 6px;
  }

  .dropdown button {
    border-radius: 12px;
    padding: 10px 12px;
    background: transparent;
    color: #111827;
    text-align: left;
    font-weight: 700;
  }

  .dropdown button:hover {
    background: #f3f4f6;
  }

  @media (max-width: 900px) {
    .topbar {
      align-items: flex-start;
      flex-direction: column;
    }
  }
</style>
