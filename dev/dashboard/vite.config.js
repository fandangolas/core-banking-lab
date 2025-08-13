import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  server: {
    host: true,
    proxy: {
      '/metrics': {
        target: process.env.VITE_API_URL || 'http://banking-api-service:8080',
        changeOrigin: true
      }
    }
  }
});
