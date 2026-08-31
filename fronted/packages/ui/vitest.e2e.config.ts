import { defineConfig } from "vitest/config";

export default defineConfig({
  test: {
    environment: "node",
    globals: true,
    include: ["tests/e2e-vitest/**/*.{test,spec}.{ts,tsx}"],
    testTimeout: 60000,
    hookTimeout: 60000
  }
});
