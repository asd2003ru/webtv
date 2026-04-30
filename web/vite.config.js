import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  build: {
    outDir: '../internal/webui/dist',
    emptyOutDir: true,
    chunkSizeWarningLimit: 700,
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (id.includes('node_modules/hls.js')) return 'hls'
          if (id.includes('node_modules')) return 'vendor'
        }
      }
    }
  }
})
