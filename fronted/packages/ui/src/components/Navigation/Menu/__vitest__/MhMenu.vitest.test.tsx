import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { MhMenu } from "./../index";

// @rc-component/resize-observer expects ResizeObserver to exist.
// jsdom does not provide it by default.
if (typeof globalThis.ResizeObserver === "undefined") {
  class ResizeObserver {
    observe() {}
    unobserve() {}
    disconnect() {}
  }

  globalThis.ResizeObserver = ResizeObserver as unknown as typeof globalThis.ResizeObserver;
}

describe("MhMenu (vitest)", () => {
  it("renders menu with items", () => {
    const items = [
      { key: "1", label: "Menu Item 1" },
      { key: "2", label: "Menu Item 2" }
    ];
    render(<MhMenu items={items} />);
    expect(screen.getByText("Menu Item 1")).toBeInTheDocument();
    expect(screen.getByText("Menu Item 2")).toBeInTheDocument();
  });

  it("applies mode prop", () => {
    const items = [{ key: "1", label: "Item" }];
    const { container } = render(<MhMenu mode="horizontal" items={items} />);
    expect(container.querySelector(".ant-menu-horizontal")).toBeInTheDocument();
  });
});
