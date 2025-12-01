import { defineConfig } from 'vite';
import { resolve } from 'path';
import tailwindcss from '@tailwindcss/vite';

export default defineConfig({
  plugins: [
    tailwindcss(),
  ],

  server: {
    host: '127.0.0.1', // Use IPv4 to avoid Windows permission issues
    port: 3000,
    origin: 'http://localhost:3000',
  },
  
  build: {
    outDir: 'static',
    emptyOutDir: true,
    manifest: true,
    rollupOptions: {
      input: {
        main: resolve(__dirname, 'assets/js/main.js'),
        style: resolve(__dirname, 'assets/css/input.css'),
      },
    },
  },
});