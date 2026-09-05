import { type Browser, chromium, type Page } from "@playwright/test";
import { createServer, type ViteDevServer } from "vite";
import { afterAll, beforeAll, describe, expect, it } from "vitest";

describe("E2E (vitest + playwright)", () => {
  let server: ViteDevServer;
  let baseURL: string;
  let browser: Browser;
  let page: Page;

  beforeAll(async () => {
    server = await createServer({
      logLevel: "silent",
      server: {
        port: 0,
        strictPort: true
      }
    });

    await server.listen();

    const address = server.httpServer?.address();
    if (address && typeof address !== "string") {
      baseURL = `http://localhost:${address.port}`;
    } else {
      throw new Error("Failed to determine dev server address");
    }

    browser = await chromium.launch({ headless: true });
    page = await browser.newPage();
  });

  afterAll(async () => {
    await page?.close();
    await browser?.close();
    await server?.close();
  });

  it("loads the preview page", async () => {
    await page.goto(baseURL, { waitUntil: "networkidle" });
    const title = await page.textContent("h1");
    expect(title).toContain("Ant Design 6 二次封装组件库");
  });

  it("renders MhButton demo content", async () => {
    await page.goto(baseURL, { waitUntil: "networkidle" });
    await page.waitForSelector("text=MhButton");
    await page.waitForSelector("text=默认按钮");
  });
});
