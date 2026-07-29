import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// Müştəri saytı kökdə (/) verilir. Dev-də /api backend-ə (:8080) proxy olunur.
export default defineConfig({
  plugins: [react()],
  server: {
    port: 5174,
    proxy: {
      '/api': 'http://localhost:8080',
      '/uploads': 'http://localhost:8080',
    },
  },
})
