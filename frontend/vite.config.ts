import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// Prod: admin panel /sirab altında verilir (base '/sirab/').
// Dev: kökdə qalır ('/') və /api, /uploads backend-ə (:8080) proxy olunur.
export default defineConfig(({ command }) => ({
  base: command === 'build' ? '/sirab/' : '/',
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      '/api': 'http://localhost:8080',
      '/uploads': 'http://localhost:8080',
    },
  },
}))
