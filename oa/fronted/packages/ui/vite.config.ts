import react from "@vitejs/plugin-react-swc";
import { defineConfig } from "vite";

const port = Number(process.env.PORT) || 3000;

export default defineConfig({
  plugins: [react()],
  server: {
    port,
    strictPort: true,
    open: true
  },
  build: {
    outDir: "distvite", // 打包目录
    emptyOutDir: true // 每次打包清空
  }
});
