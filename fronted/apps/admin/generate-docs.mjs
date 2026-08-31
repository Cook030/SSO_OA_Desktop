/* global console, process */

import fs from "node:fs";
import path from "node:path";
import { loadSchema } from "@graphql-tools/load";
import { UrlLoader } from "@graphql-tools/url-loader";
import { buildOperationNodeForField } from "@graphql-tools/utils";
import { print } from "graphql";

// 配置项：定义需要生成文档的 GraphQL 服务
const generateConfigs = [
  {
    url: "https://admin-api-dmp.maplehaze.cn/graphql",
    outputDir: "./src/generated/generated-graphql/service-dmp"
  },
  {
    url: "https://admin-api-sso2.maplehaze.cn/graphql",
    outputDir: "./src/generated/generated-graphql/service-sso2"
  }
];

// 操作类型配置
const OPERATION_TYPES = ["query", "mutation"];
const DEPTH_LIMIT = 2; // 限制深度，防止生成的 query 太大

/**
 * 为指定的 GraphQL 服务生成文档
 * @param {Object} config - 生成配置
 * @param {string} config.url - GraphQL 服务的 URL
 * @param {string} config.outputDir - 输出目录路径
 */
async function generateDocs(config) {
  const { url, outputDir } = config;

  try {
    console.log(`正在从 ${url} 加载 schema...`);

    // 加载 GraphQL schema
    const schema = await loadSchema(url, {
      loaders: [new UrlLoader()]
    });

    console.log(`Schema 加载成功，开始生成文档到 ${outputDir}`);

    // 遍历每种操作类型（query, mutation）
    for (const operationType of OPERATION_TYPES) {
      await generateOperationDocs(schema, operationType, outputDir);
    }

    console.log(`✓ ${url} 的文档生成完成`);
  } catch (error) {
    console.error(`✗ 生成 ${url} 的文档时出错:`, error.message);
    throw error;
  }
}

/**
 * 为指定的操作类型生成文档
 * @param {Object} schema - GraphQL schema
 * @param {string} operationType - 操作类型 (query 或 mutation)
 * @param {string} outputDir - 输出目录路径
 */
async function generateOperationDocs(schema, operationType, outputDir) {
  // 获取对应的根类型
  const rootType = operationType === "query" ? schema.getQueryType() : schema.getMutationType();

  if (!rootType) {
    console.warn(`警告: Schema 中没有 ${operationType} 类型`);
    return;
  }

  const fields = rootType.getFields();
  const fieldNames = Object.keys(fields);

  if (fieldNames.length === 0) {
    console.warn(`警告: ${operationType} 类型中没有字段`);
    return;
  }

  // 创建输出目录（queries 或 mutations）
  const typeDir = path.join(outputDir, `${operationType}s`);
  if (!fs.existsSync(typeDir)) {
    fs.mkdirSync(typeDir, { recursive: true });
  }

  console.log(`  生成 ${fieldNames.length} 个 ${operationType} 文档...`);

  // 为每个字段生成 GraphQL 文档
  fieldNames.forEach(fieldName => {
    try {
      // 自动构建 Operation Node
      const operation = buildOperationNodeForField({
        schema,
        kind: operationType,
        field: fieldName,
        depthLimit: DEPTH_LIMIT
      });

      // 将操作转换为 GraphQL 字符串并写入文件
      const graphqlContent = print(operation);
      const filePath = path.join(typeDir, `${fieldName}.graphql`);
      fs.writeFileSync(filePath, graphqlContent);
    } catch (error) {
      console.error(`  ✗ 生成 ${fieldName} 时出错:`, error.message);
    }
  });
}

/**
 * 主函数：批量生成所有配置的文档
 */
async function main() {
  console.log("开始生成 GraphQL 文档...\n");

  try {
    // 并行生成所有配置的文档
    await Promise.all(generateConfigs.map(config => generateDocs(config)));

    console.log("\n所有文档生成完成！");
  } catch (error) {
    console.error("\n文档生成失败:", error.message);
    process.exit(1);
  }
}

// 执行主函数
main();
