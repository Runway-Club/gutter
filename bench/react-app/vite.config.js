import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

// base './' so the built bundle resolves under the /react/ mount path used by
// the benchmark static server.
export default defineConfig({
  base: './',
  plugins: [react()],
  build: {
    target: 'es2020',
    // single chunk keeps the bundle-size comparison simple and is what a small
    // app ships in practice.
    rollupOptions: { output: { manualChunks: undefined } },
  },
});
