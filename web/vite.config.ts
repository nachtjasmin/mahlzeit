import { defineConfig } from "vite";

export default defineConfig({
  build: {
    manifest: true,
    rollupOptions: {
      input: ["./assets/entrypoint.ts"],
    },
  },

  // Although we don't need CORS for local development, it's necessary for GitHub Codespaces.
  server: {
    cors: { origin: "*" },
  },
});
