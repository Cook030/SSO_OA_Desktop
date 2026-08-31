import { render } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { MhMessage } from "../index";

// Mock antd 的 message 模块，只保留我们需要的静态方法
vi.mock("antd", () => ({
  message: {
    success: vi.fn(),
    error: vi.fn(),
    info: vi.fn(),
    warning: vi.fn(),
    config: vi.fn(),
    destroy: vi.fn()
  }
}));

describe("MhMessage (方案二: 空组件挂载静态方法)", () => {
  it("should be a React component that renders null without crashing", () => {
    // 渲染组件（虽然返回 null，但不应报错）
    const { container } = render(<MhMessage />);
    expect(container.firstChild).toBeNull();
  });

  it("should have all static methods attached", () => {
    expect(MhMessage.success).toBeTypeOf("function");
    expect(MhMessage.error).toBeTypeOf("function");
    expect(MhMessage.info).toBeTypeOf("function");
    expect(MhMessage.warning).toBeTypeOf("function");
    expect(MhMessage.config).toBeTypeOf("function");
    expect(MhMessage.destroy).toBeTypeOf("function");
  });

  it("should call antd's message.success when MhMessage.success is invoked", () => {
    const { message } = require("antd");
    MhMessage.success("test success");
    expect(message.success).toHaveBeenCalledWith("test success");
  });

  it("should call antd's message.error when MhMessage.error is invoked", () => {
    const { message } = require("antd");
    MhMessage.error("test error");
    expect(message.error).toHaveBeenCalledWith("test error");
  });

  it("should call antd's message.info when MhMessage.info is invoked", () => {
    const { message } = require("antd");
    MhMessage.info("test info");
    expect(message.info).toHaveBeenCalledWith("test info");
  });

  it("should call antd's message.warning when MhMessage.warning is invoked", () => {
    const { message } = require("antd");
    MhMessage.warning("test warning");
    expect(message.warning).toHaveBeenCalledWith("test warning");
  });

  it("should call antd's message.config when MhMessage.config is invoked", () => {
    const { message } = require("antd");
    const config = { duration: 2 };
    MhMessage.config(config);
    expect(message.config).toHaveBeenCalledWith(config);
  });

  it("should call antd's message.destroy when MhMessage.destroy is invoked", () => {
    const { message } = require("antd");
    MhMessage.destroy();
    expect(message.destroy).toHaveBeenCalled();
  });
});
