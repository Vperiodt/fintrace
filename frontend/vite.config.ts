import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

const apiTarget = process.env.API_BASE_URL || "http://localhost:8080";

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      "/users": {
        target: apiTarget,
        changeOrigin: true
      },
      "/transactions": {
        target: apiTarget,
        changeOrigin: true
      },
      "/relationships": {
        target: apiTarget,
        changeOrigin: true
      }
    }
  }
});

