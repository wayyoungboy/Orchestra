import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'

const orchestraAPIURL = process.env.ORCHESTRA_API_URL || 'http://127.0.0.1:8080'
const orchestraWebSocketURL = orchestraAPIURL.replace(/^http/, 'ws')

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src')
    }
  },
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: orchestraAPIURL,
        changeOrigin: true
      },
      '/ws': {
        target: orchestraWebSocketURL,
        ws: true
      }
    }
  },
  preview: {
    proxy: {
      '/api': {
        target: orchestraAPIURL,
        changeOrigin: true
      },
      '/ws': {
        target: orchestraWebSocketURL,
        ws: true
      }
    }
  },
  test: {
    environment: 'jsdom',
    globals: true,
    include: ['src/**/*.test.ts'],
  }
})
