<script>
  import { api } from '../lib/api.js'
  import { user, error, success } from '../lib/stores.js'

  let profileForm = { email: $user?.email || '', name: $user?.name || '', password: '' }

  $: if ($user) {
    profileForm = { email: $user.email, name: $user.name || '', password: '' }
  }

  async function updateProfile() {
    error.set('')
    success.set('')
    try {
      const payload = await api('/api/profile', {
        method: 'PUT',
        body: { email: profileForm.email, name: profileForm.name },
      })
      user.set(payload.user)
      profileForm = { email: payload.user.email, name: payload.user.name || '', password: '' }
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
  <label>Name <input bind:value={profileForm.name} /></label>
  <label>E-Mail <input bind:value={profileForm.email} type="email" /></label>
  <button class="btn secondary" on:click={updateProfile}>Profil speichern</button>
  <label>Neues Passwort <input bind:value={profileForm.password} type="password" /></label>
  <button class="btn secondary" on:click={updatePassword}>Passwort ändern</button>
</section>

<style>
  .profile-page {
    max-width: 620px;
    padding: 14px 20px;
  }

  .stack {
    display: grid;
    gap: 14px;
  }
</style>
