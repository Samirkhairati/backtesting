import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

import path from "path"

import { TanStackRouterVite } from '@tanstack/router-plugin/vite'

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    TanStackRouterVite(),
    react()
  ],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
  server: {
    // The client calls the API on its own origin (/api/v1); forward it in dev.
    proxy: {
      "/api": {
        target: process.env.API_PROXY_TARGET ?? "http://api:8080",
        changeOrigin: true,
      },
    },
  },
})
