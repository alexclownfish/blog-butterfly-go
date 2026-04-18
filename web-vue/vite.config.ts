import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import path from 'node:path'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, 'src')
    }
  },
  server: {
    host: '0.0.0.0',
    port: 18086,
    proxy: {
      '/api': {
        target: 'http://172.28.74.191:31083',
        changeOrigin: true
      }
    }
  }
})
