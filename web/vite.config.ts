import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from "node:path"
import tailwindcss from '@tailwindcss/vite'
import { viteStaticCopy } from 'vite-plugin-static-copy'

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    react(),
    tailwindcss(),
    viteStaticCopy({
      targets: [
        {
          src: 'node_modules/three/examples/jsm/libs/basis/*',
          dest: 'basis'
        }
      ]
    })
  ],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
  server:{
    proxy:{
         '/api': 'http://localhost:8080',
         '/scans': 'http://localhost:8080',
    }
  }
})
