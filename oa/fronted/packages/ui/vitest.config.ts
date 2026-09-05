import { defineConfig } from "vitest/config";

export default defineConfig({
  test: {
    environment: "jsdom",
    globals: true,
    setupFiles: ["./tests/setup-vitest.ts"],
    include: ["src/components/**/__vitest__/*.{test,spec}.{ts,tsx}"]
  }
});
