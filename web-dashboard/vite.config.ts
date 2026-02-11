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
    proxy: {
      '/health': 'http://localhost:9090',
      '/metrics': 'http://localhost:9090',
      '/info': 'http://localhost:9090',
      '/analytics': 'http://localhost:9090',
      '/api': 'http://localhost:9090',
    },
  },
})
