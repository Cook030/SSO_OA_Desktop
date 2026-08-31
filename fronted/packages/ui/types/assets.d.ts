// global.d.ts 或 assets.d.ts
declare module "*.css" {
  const content: string;
  export default content;
}

// 如果也导入了 less 等文件，可以一并声明
declare module "*.less" {
  const content: string;
  export default content;
}
