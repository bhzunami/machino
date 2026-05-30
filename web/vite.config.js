import { defineConfig } from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'
import { readFileSync, writeFileSync } from 'fs'
import { resolve } from 'path'

// Replaces __SW_VERSION__ in the built service-worker.js with the current
// build timestamp so the browser always picks up a fresh cache on deploy.
function swVersionPlugin() {
  return {
    name: 'sw-version',
    closeBundle() {
      const swPath = resolve(__dirname, 'dist/service-worker.js')
      const version = Date.now()
      const content = readFileSync(swPath, 'utf-8').replace('__SW_VERSION__', String(version))
      writeFileSync(swPath, content)
    },
  }
}

export default defineConfig({
  plugins: [svelte(), swVersionPlugin()],
  server: {
    host: '0.0.0.0',
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: false,
        ws: true,
      },
    },
  },
})
