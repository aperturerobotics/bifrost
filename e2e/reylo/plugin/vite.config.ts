import { resolve } from 'node:path'
import reylo from 'reylo/vite'
import { defineConfig } from 'vite'

// The plugin bundle: main.ts and everything it imports in one file, its .sql
// queries compiled by Reylo against DATABASE_URL.
export default defineConfig({
  plugins: [reylo()],
  resolve: {
    alias: {
      '@aptre/bldr-sdk': resolve(import.meta.dirname, '../../../bldr/sdk'),
    },
  },
  ssr: { noExternal: true },
  build: {
    ssr: 'main.ts',
    target: 'esnext',
    emptyOutDir: true,
  },
})
