import commonjs from "@rollup/plugin-commonjs";
import resolve from "@rollup/plugin-node-resolve";
import esbuild from "rollup-plugin-esbuild";
import livereload from "rollup-plugin-livereload";
import postcss from "rollup-plugin-postcss";
import serve from "rollup-plugin-serve";

const port = Number(process.env.PORT) || 3000;
const sourcemap = process.env.SOURCEMAP === "true";

export default {
  input: "dev/index.tsx",
  onwarn(warning, warn) {
    if (warning.code === "MODULE_LEVEL_DIRECTIVE" && /use client/i.test(warning.message)) {
      return;
    }

    if (warning.code === "THIS_IS_UNDEFINED") {
      return;
    }

    warn(warning);
  },
  output: {
    file: "dev/bundle.js",
    format: "iife",
    sourcemap,
    banner:
      'globalThis.process = globalThis.process || { env: { NODE_ENV: "development" } };\n' +
      'globalThis.process.env = globalThis.process.env || { NODE_ENV: "development" };'
  },
  plugins: [
    resolve({
      browser: true,
      extensions: [".js", ".jsx", ".ts", ".tsx"]
    }),
    esbuild({
      include: /\.[jt]sx?$/,
      minify: false,
      jsx: "automatic",
      tsconfig: "./tsconfig.dev.json",
      platform: "browser",
      define: {
        "process.env.NODE_ENV": '"development"',
        "process.env": "{}"
      }
    }),
    commonjs({
      include: /node_modules/
    }),
    postcss({
      extract: false,
      modules: false,
      use: ["sass"]
    }),
    serve({
      open: false,
      contentBase: ["dev"],
      host: "localhost",
      port
    }),
    livereload({
      watch: "dev"
    })
  ]
};
