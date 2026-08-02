import { fileURLToPath } from "node:url";
import tailwindcss from "@tailwindcss/vite";
import react from "@vitejs/plugin-react";
import { defineConfig } from "vitest/config";

export default defineConfig({
  root: fileURLToPath(new URL(".", import.meta.url)),
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      "@": fileURLToPath(new URL("./src", import.meta.url)),
    },
  },
  server: {
    host: "0.0.0.0",
    port: 3000,
    watch: {
      usePolling: true,
    },
  },
  build: {
    rollupOptions: {
      output: {
        manualChunks: {
          "react-vendor": ["react", "react-dom", "react-router-dom"],
          "state-vendor": ["@reduxjs/toolkit", "react-redux", "@tanstack/react-query"],
          "forms-vendor": ["@hookform/resolvers", "react-hook-form", "zod"],
          "i18n-vendor": [
            "i18next",
            "i18next-browser-languagedetector",
            "react-i18next",
          ],
        },
      },
    },
  },
  test: { environment: "jsdom", globals: true, setupFiles: "./src/test-setup.ts" },
});
