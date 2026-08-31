import type { CodegenConfig } from "@graphql-codegen/cli";

// const config: CodegenConfig = {
//   schema: "https://admin-api-dmp.maplehaze.cn/graphql",

//   documents: "./src/generated/generated-graphql/*/*.graphql",

//   generates: {
//     "./src/generated/generated-ts/sdk.ts": {
//       plugins: ["typescript", "typescript-operations", "typescript-graphql-request"]
//     }
//   }
// };
const config: CodegenConfig = {
  generates: {
    // 服务 A 的 SDK
    "./src/generated/generated-ts/service-dmp/sdk.ts": {
      schema: "https://admin-api-dmp.maplehaze.cn/graphql",
      documents: "./src/generated/generated-graphql/service-dmp/**/*.graphql",
      plugins: ["typescript", "typescript-operations", "typescript-graphql-request"],
      config: {
        useTypeImports: true,
        // 关键点：映射自定义标量
        scalars: {
          JSON: "Record<string, any>" // 或者直接使用 'any'
        }
      }
    },
    // 服务 B 的 SDK
    "./src/generated/generated-ts/service-sso2/sdk.ts": {
      schema: "https://admin-api-sso2.maplehaze.cn/graphql",
      documents: "./src/generated/generated-graphql/service-sso2/**/*.graphql",
      plugins: ["typescript", "typescript-operations", "typescript-graphql-request"],
      config: {
        useTypeImports: true,
        // 关键点：映射自定义标量
        scalars: {
          JSON: "Record<string, any>" // 或者直接使用 'any'
        }
      }
    }
  }
};

export default config;
