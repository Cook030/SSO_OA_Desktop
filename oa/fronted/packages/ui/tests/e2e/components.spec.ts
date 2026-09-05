import { expect, test } from "@playwright/test";

test.describe("Component Library E2E Tests", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto("/");
  });

  test("page loads successfully", async ({ page }) => {
    await expect(page.locator("h1")).toContainText("枫岚 Ant Design 组件库开发预览");
  });

  test("MhButton components render", async ({ page }) => {
    // Check if MhButtond exists
    await expect(page.locator("text=MhButton 示例")).toBeVisible();
    // Check if buttons are rendered
    await expect(page.locator("text=默认按钮")).toBeVisible();
    await expect(page.locator("text=👉 带前缀按钮")).toBeVisible();
    await expect(page.locator("text=带后缀按钮 👈")).toBeVisible();
  });

  test("MhButton be clicked", async ({ page }) => {
    const button = page.locator('button:has-text("默认按钮")');
    await expect(button).toBeVisible();
    await button.click();
  });

  test("MhInput components render", async ({ page }) => {
    // Check if MhInput card exists
    await expect(page.locator("text=MhInput 示例")).toBeVisible();

    // Check if inputs are rendered
    const inputs = page.locator('input[type="text"]');
    await expect(inputs.first()).toBeVisible();
  });

  test("MhInput accepts text input", async ({ page }) => {
    const input = page.locator('input[placeholder="请输入内容"]');
    await input.fill("Test input value");
    await expect(input).toHaveValue("Test input value");

    // Check if the value is displayed
    await expect(page.locator("text=当前输入值")).toBeVisible();
    await expect(page.locator("text=Test input value")).toBeVisible();
  });

  test("MhCard components render", async ({ page }) => {
    // Check if MhCard examples exist
    await expect(page.locator("text=MhCard 示例")).toBeVisible();

    // Check card content
    await expect(page.locator("text=这是一个带有阴影和悬浮效果的自定义卡片组件")).toBeVisible();
  });

  test("all three card sections are visible", async ({ page }) => {
    const cards = page.locator(".ant-card");
    await expect(cards).toHaveCount(3);
  });

  test("buttons have correct types", async ({ page }) => {
    // Primary button
    const primaryButton = page.locator("button.ant-btn-primary").first();
    await expect(primaryButton).toBeVisible();

    // Dashed button
    // const dashedButton = page.locator('button:has-text("🎨 虚线按钮")');
    // await expect(dashedButton).toBeVisible();
  });
});
