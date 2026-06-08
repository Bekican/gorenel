import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig(({ isSsrBuild }) => ({
  plugins: [
    react({
      babel: {
        plugins: [['babel-plugin-react-compiler']],
      },
    }),
  ],
  build: {
    target: 'es2022',
    modulePreload: false,
    rollupOptions: isSsrBuild
      ? undefined
      : {
          output: {
            manualChunks(id) {
              if (!id.includes('node_modules')) return
              if (id.includes('recharts')) return 'charts'
              if (id.includes('leaflet') || id.includes('react-leaflet')) return 'maps'
              if (id.includes('axios')) return 'http-vendor'
              if (id.includes('i18next') || id.includes('react-i18next')) return 'i18n-vendor'
              if (id.includes('react-router')) return 'router-vendor'
              if (id.includes('react-dom')) return 'vendor-react'
              if (id.includes('node_modules/react/')) return 'vendor-react'
            },
          },
        },
  },
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
  ssgOptions: {
    crittersOptions: {
      // Inline critical CSS and keep CSS preloads efficient.
      preload: 'swap',
    },
  },
}))
