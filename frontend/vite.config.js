import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueDevTools from 'vite-plugin-vue-devtools'

const proxyTarget = process.env.VITE_PROXY_TARGET || 'http://localhost:8080'

const authEndpointRedirectPlugin = () => ({
  name: 'auth-endpoint-redirect',
  configureServer(server) {
    server.middlewares.use((req, res, next) => {
      const path = req.url?.split('?')[0]
      const isAuthEndpointPath = path === '/api/v1/auth/login' || path === '/api/v1/auth/register'

      if (req.method === 'GET' && isAuthEndpointPath) {
        res.statusCode = 302
        res.setHeader('Location', '/login')
        res.end()
        return
      }

      next()
    })
  },
})

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    vue(),
    vueDevTools(),
    authEndpointRedirectPlugin(),
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    },
  },
  server: {
    host: true,
    port: 5173,
    watch: {
      usePolling: true,
    },
    proxy: {
      '/api': {
        target: proxyTarget,
        changeOrigin: true
      }
    }
  }
})
