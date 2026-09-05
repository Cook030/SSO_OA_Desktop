import { expect, test } from "@playwright/test";

test.describe("简单的 E2E 测试", () => {
  test("页面能够加载", async ({ page }) => {
    await page.goto("http://localhost:3000");

    // 等待页面加载
    await page.waitForLoadState("networkidle");

    // 检查标题是否存在
    const title = await page.textContent("h1");
    expect(title).toContain("Ant Design");

    console.log("页面标题:", title);
  });

  test("检查按钮是否存在", async ({ page }) => {
    await page.goto("http://localhost:3000");
    await page.waitForLoadState("networkidle");

    // 等待按钮出现
    const buttons = await page.locator("button").count();
    console.log("按钮数量:", buttons);
    expect(buttons).toBeGreaterThan(0);
  });
});
