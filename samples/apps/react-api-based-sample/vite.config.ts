import fs from 'fs'
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    https: {
      key: fs.readFileSync('./server.key'),
      cert: fs.readFileSync('./server.cert'),
    },
    port: 3000,
    host: 'localhost',
  },
  preview: {
    https: {
      key: fs.readFileSync('./server.key'),
      cert: fs.readFileSync('./server.cert'),
    },
    port: 3000,
    host: 'localhost',
  },
})
