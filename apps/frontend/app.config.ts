/// <reference types="@tanstack/start/config" />
import { defineConfig } from "@tanstack/start/config";
import tailwindcss from "@tailwindcss/vite";

export default defineConfig({
  server: {
    port: 3000,
  },
  vite: {
    plugins: [
      tailwindcss(),
    ],
    resolve: {
      alias: {
        "@": "/src",
      },
    },
  },
});
