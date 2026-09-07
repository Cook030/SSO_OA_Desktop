import { fileURLToPath, URL } from "node:url";

import tailwindcss from "@tailwindcss/vite";
import vue from "@vitejs/plugin-vue";
import { defineConfig } from "vite";

// 本地开发将 /api 代理到 mh-sso-svc，浏览器同源携带 Cookie，规避跨域与 Domain 配置
export default defineConfig({
  plugins: [vue(), tailwindcss()],
  resolve: {
    alias: {
      "@": fileURLToPath(new URL("./src", import.meta.url)),
    },
  },
  server: {
    // port: 5173, // 默认端口（与本地 oa 前端 5173 冲突）
    port: 8004, // 纯本地：SSO 前端端口，与 oa 前端（5173）区分
    proxy: {
      "/api": {
        // target: "http://127.0.0.1:8080", // 原 SSO 后端端口（与 oa 后端冲突）
        target: "http://127.0.0.1:8081", // 纯本地：sso/server 已改为 8081
        changeOrigin: true,
      },
    },
  },
});
