import vue from '@vitejs/plugin-vue'
import { defineConfig } from 'vite'
import { createHtmlPlugin } from 'vite-plugin-html'
import vueDevTools from 'vite-plugin-vue-devtools'

export default defineConfig({
  plugins: [
    vue(),
    vueDevTools(),
    createHtmlPlugin({})
  ],
  server: {
    host: '0.0.0.0',      // слушать на всех интерфейсах внутри контейнера
    port: 5173,
    allowedHosts: 'all',
    strictPort: true,
    hmr: {
      clientPort: 8000,   // браузер подключается к HMR через порт nginx
    },
  },
})