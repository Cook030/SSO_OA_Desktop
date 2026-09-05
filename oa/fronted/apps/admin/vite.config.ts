import path from "node:path";
import tailwindcss from "@tailwindcss/vite";
import { defineConfig, loadEnv } from "vite";
import qiankun from "vite-plugin-qiankun";
import svgr from "vite-plugin-svgr";

const APP_NAME = "admin";
const DEV_HOST = "oa.maplehaze.cn";
const DEV_PORT = 5173;

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, __dirname, "");

  return {
    plugins: [
      qiankun(APP_NAME, {
        useDevMode: true
      }),
      svgr(),
      tailwindcss()
    ],

    resolve: {
      alias: {
        "@": path.resolve(__dirname, "./src"),
        "@assets": path.resolve(__dirname, "./src/assets"),
        "@mh-repo/ui": path.resolve(__dirname, "../../packages/ui/src"),
        "@mh-repo/utils": path.resolve(__dirname, "../../packages/utils/src"),
        "@mh-repo/types": path.resolve(__dirname, "../../packages/types/src"),
        "@mh-repo/apis": path.resolve(__dirname, "../../packages/apis/src")
      }
    },

    server: {
      host: true,
      port: DEV_PORT,
      cors: true,
      // 允许本地开发时的域名访问
      allowedHosts: ["maplehaze.cn", ".maplehaze.cn"],
      open: `http://${DEV_HOST}:${DEV_PORT}/${APP_NAME}/`,

      // ✅ 核心修改：为 qiankun 添加跨域头，并修复代理路径
      headers: {
        "Access-Control-Allow-Origin": "*"
      },

      proxy: {
        // 1. 处理带 /admin 前缀的 API 请求（由 base 路径引起）
        "/admin/api": {
          target: "https://oa.maplehaze.cn",
          changeOrigin: true,
          secure: false,
          // 转发时去掉 /admin 前缀，变成 /api/xxx
          rewrite: path => path.replace(/^\/admin/, ""),
          // 重写 Origin/Referer，避免后端 CSRF 校验拒绝 POST/PUT/DELETE
          configure: proxy => {
            proxy.on("proxyReq", proxyReq => {
              proxyReq.setHeader("Origin", "https://oa.maplehaze.cn");
            });
          }
        },

        "/api/v1/auth/": {
          target: "https://sso2.maplehaze.cn",
          changeOrigin: true,
          secure: false,
          // 重写 Origin/Referer，避免后端 CSRF 校验拒绝 POST/PUT/DELETE
          configure: proxy => {
            proxy.on("proxyReq", proxyReq => {
              proxyReq.setHeader("Origin", "https://sso2.maplehaze.cn");
            });
          }
        },
        // 2. 兜底：处理直接以 /api 开头的请求（防止其他地方直接写绝对路径）
        "/api": {
          target: "https://oa.maplehaze.cn",
          changeOrigin: true,
          secure: false,
          // 重写 Origin/Referer，避免后端 CSRF 校验拒绝 POST/PUT/DELETE
          configure: proxy => {
            proxy.on("proxyReq", proxyReq => {
              proxyReq.setHeader("Origin", "https://oa.maplehaze.cn");
            });
          }
        }
      }
    },

    // 根据环境和模式动态设置 base 路径
    base: env.VITE_APP_BASE || (mode === "production" ? `https://oa.maplehaze.cn/${APP_NAME}/` : `/${APP_NAME}/`),

    build: {
      outDir: "dist"
    },

    css: {
      postcss: "./../../postcss.config.js",
      preprocessorOptions: {
        less: {
          javascriptEnabled: true
          // 如果有全局变量，可以在这里引入
          // additionalData: `@import "./src/styles/variables.less";`,
        }
      }
    }
  };
});
