import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import MhTypography from "../index";

// 模拟 clipboard API
Object.assign(navigator, {
  clipboard: {
    writeText: vi.fn().mockResolvedValue(undefined)
  }
});

describe("MhTypography (copyable) — using fireEvent", () => {
  it("renders copyable text and triggers copy", async () => {
    render(<MhTypography.Paragraph copyable>可复制的文本</MhTypography.Paragraph>);

    // 定位复制按钮（Ant Design 默认会渲染一个带复制图标的按钮）
    const copyButton = screen.getByRole("button", { name: /copy/i });
    expect(copyButton).toBeInTheDocument();

    // 使用 fireEvent 模拟点击
    fireEvent.click(copyButton);

    // 等待异步复制完成
    await waitFor(() => {
      expect(navigator.clipboard.writeText).toHaveBeenCalledWith("可复制的文本");
    });
  });

  it("uses custom copy text", async () => {
    render(<MhTypography.Paragraph copyable={{ text: "自定义复制内容" }}>显示内容</MhTypography.Paragraph>);

    const copyButton = screen.getByRole("button", { name: /copy/i });
    fireEvent.click(copyButton);

    await waitFor(() => {
      expect(navigator.clipboard.writeText).toHaveBeenCalledWith("自定义复制内容");
    });
  });

  it("renders custom copy icon", () => {
    render(
      <MhTypography.Paragraph
        copyable={{
          icon: [<span key="1">📋</span>, <span key="2">✅</span>]
        }}
      >
        自定义图标
      </MhTypography.Paragraph>
    );
    expect(screen.getByText("📋")).toBeInTheDocument();
  });

  it("hides tooltips when tooltips set to false", () => {
    render(<MhTypography.Paragraph copyable={{ tooltips: false }}>无提示</MhTypography.Paragraph>);
    expect(screen.getByText("无提示")).toBeInTheDocument();
  });

  it("handles async copy text", async () => {
    render(
      <MhTypography.Paragraph
        copyable={{
          text: async () =>
            new Promise(resolve => {
              setTimeout(() => resolve("异步文本"), 100);
            })
        }}
      >
        异步复制
      </MhTypography.Paragraph>
    );

    const copyButton = screen.getByRole("button", { name: /copy/i });
    fireEvent.click(copyButton);

    await waitFor(() => {
      expect(navigator.clipboard.writeText).toHaveBeenCalledWith("异步文本");
    });
  });

  it("works with MhTypography.Text copyable", async () => {
    render(<MhTypography.Text copyable>文本组件复制</MhTypography.Text>);
    const copyButton = screen.getByRole("button", { name: /copy/i });
    fireEvent.click(copyButton);
    await waitFor(() => {
      expect(navigator.clipboard.writeText).toHaveBeenCalledWith("文本组件复制");
    });
  });
});
