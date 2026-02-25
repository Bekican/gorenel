import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    react({
      babel: {
        plugins: [['babel-plugin-react-compiler']],
      },
    }),
  ],
  // TODO [K8S]: Local dev proxy'si burasıdır. 
  // Production için (K8s) bu yönlendirmeleri nginx.conf içinde yapmalısınız.
  server: {
    port: 5176,
    host: '127.0.0.1',
    proxy: {
      '/health': 'http://127.0.0.1:9091',
      '/metrics': 'http://127.0.0.1:9091',
      '/info': 'http://127.0.0.1:9091',
      '/analytics': 'http://127.0.0.1:9091',
      '/api': 'http://127.0.0.1:9091',
    },
  },
})
