import { render, screen } from "@testing-library/react";
import { beforeAll, describe, expect, it } from "vitest";
import { MhGrid } from "./../index";

describe("MhGrid (vitest)", () => {
  beforeAll(() => {
    global.ResizeObserver = class ResizeObserver {
      observe() {}
      unobserve() {}
      disconnect() {}
    };
    Object.defineProperty(window, "matchMedia", {
      writable: true,
      value: (query: string) => ({
        matches: false,
        media: query,
        onchange: null,
        addListener: () => {},
        removeListener: () => {},
        addEventListener: () => {},
        removeEventListener: () => {},
        dispatchEvent: () => {}
      })
    });
  });
  it("renders Row and Col", () => {
    render(
      <MhGrid.Row>
        <MhGrid.Col span={12}>Column 1</MhGrid.Col>
        <MhGrid.Col span={12}>Column 2</MhGrid.Col>
      </MhGrid.Row>
    );
    expect(screen.getByText("Column 1")).toBeInTheDocument();
    expect(screen.getByText("Column 2")).toBeInTheDocument();
  });

  it("applies gutter prop to Row", () => {
    const { container } = render(
      <MhGrid.Row gutter={16}>
        <MhGrid.Col>Content</MhGrid.Col>
      </MhGrid.Row>
    );
    expect(container.querySelector(".ant-row")).toBeInTheDocument();
  });

  it("applies span prop to Col", () => {
    const { container } = render(
      <MhGrid.Row>
        <MhGrid.Col span={24}>Full Width</MhGrid.Col>
      </MhGrid.Row>
    );
    expect(container.querySelector(".ant-col-24")).toBeInTheDocument();
  });
});
