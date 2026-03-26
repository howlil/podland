// app.config.ts
import { defineConfig } from "@tanstack/start/config";
import tailwindcss from "@tailwindcss/vite";
var app_config_default = defineConfig({
  server: {
    port: 3e3
  },
  vite: {
    plugins: [
      tailwindcss()
    ],
    resolve: {
      alias: {
        "@": "/src"
      }
    }
  }
});
export {
  app_config_default as default
};
