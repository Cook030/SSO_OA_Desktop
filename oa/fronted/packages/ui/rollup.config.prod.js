// // import resolve from "@rollup/plugin-node-resolve";
// // import commonjs from "@rollup/plugin-commonjs";
// // import typescript from "@rollup/plugin-typescript";
// // import postcss from "rollup-plugin-postcss";
// // import dts from "rollup-plugin-dts";

// // const config = [
// //   {
// //     input: "src/index.ts",
// //     output: [
// //       {
// //         file: "dist/index.cjs.js",
// //         format: "cjs",
// //         sourcemap: true,
// //         exports: "named"
// //       },
// //       {
// //         file: "dist/index.esm.js",
// //         format: "esm",
// //         sourcemap: true
// //       }
// //     ],
// //     external: ["react", "react-dom", "antd"],
// //     plugins: [
// //       resolve(),
// //       commonjs(),
// //       typescript({
// //         // tsconfig: "./tsconfig.json",
// //         // declaration: true,
// //         // declarationDir: "./dist/types",
// //         // outDir: "./dist"
// //         tsconfig: "./tsconfig.json",
// //         declaration: true,
// //         declarationDir: "./dist",
// //         declarationMap: true, // 可选：生成声明文件的sourcemap
// //         emitDeclarationOnly: false // 确保不只是生成声明文件
// //       }),
// //       postcss({
// //         extract: false,
// //         modules: false
// //       })
// //     ]
// //   },
// //   {
// //     input: "dist/types/index.d.ts",
// //     output: [{ file: "dist/index.d.ts", format: "es" }],
// //     external: [/\.css$/],
// //     plugins: [dts()]
// //   }
// // ];

// // export default config;
// import resolve from "@rollup/plugin-node-resolve";
// import commonjs from "@rollup/plugin-commonjs";
// import typescript from "@rollup/plugin-typescript";
// import postcss from "rollup-plugin-postcss";

// const config = [
//   {
//     input: "src/index.ts",
//     output: [
//       {
//         file: "dist/index.cjs.js",
//         format: "cjs",
//         sourcemap: true,
//         exports: "named"
//       },
//       {
//         file: "dist/index.esm.js",
//         format: "esm",
//         sourcemap: true
//       }
//     ],
//     external: ["react", "react-dom", "antd"],
//     plugins: [
//       resolve(),
//       commonjs(),
//       typescript({
//         tsconfig: "./tsconfig.json",
//         declaration: true,
//         declarationDir: "./dist",
//         declarationMap: true, // 可选：生成声明文件的sourcemap
//         emitDeclarationOnly: false // 确保不只是生成声明文件
//       }),
//       postcss({
//         extract: false,
//         modules: false
//       })
//     ]
//   }
// ];

// export default config;

import { readFileSync } from "node:fs";
import commonjs from "@rollup/plugin-commonjs";
import resolve from "@rollup/plugin-node-resolve";
import typescript from "@rollup/plugin-typescript";
import dts from "rollup-plugin-dts";
import postcss from "rollup-plugin-postcss";

/** Rollup 将 SVG 内联为 data URL，兼容 Vite 开发态的 URL 导入 */
function svgDataUrl() {
  return {
    name: "svg-data-url",
    load(id) {
      if (!id.endsWith(".svg")) {
        return null;
      }

      const dataUrl = `data:image/svg+xml;base64,${readFileSync(id, "base64")}`;

      return {
        code: `export default ${JSON.stringify(dataUrl)};`,
        map: { mappings: "" }
      };
    }
  };
}

const config = [
  {
    input: "src/index.ts",
    output: [
      {
        file: "dist/index.cjs.js",
        format: "cjs",
        sourcemap: true,
        exports: "named"
      },
      {
        file: "dist/index.esm.js",
        format: "esm",
        sourcemap: true
      }
    ],
    external: ["react", "react-dom", "antd"],
    plugins: [
      svgDataUrl(),
      resolve(),
      commonjs(),
      typescript({
        tsconfig: "./tsconfig.json",
        declaration: false, // 不在第一个配置中生成声明文件
        outDir: "./dist"
        // tsconfig: "./tsconfig.json",
        // declaration: true,
        // declarationDir: "./dist/types",
        // outDir: "./dist"
      }),
      postcss({
        extract: false,
        modules: false
      })
    ]
  },
  {
    input: "src/index.ts", // 改为从源码读取
    output: [{ file: "dist/index.d.ts", format: "es" }],
    external: [/\.css$/],
    plugins: [dts()]
  }
];

export default config;
