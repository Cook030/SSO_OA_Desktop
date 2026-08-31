import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { MhAnchor } from "./../index";

if (typeof globalThis.ResizeObserver === "undefined") {
  class ResizeObserver {
    observe() {}
    unobserve() {}
    disconnect() {}
  }

  globalThis.ResizeObserver = ResizeObserver as unknown as typeof globalThis.ResizeObserver;
}

if (typeof window !== "undefined" && typeof window.matchMedia === "undefined") {
  window.matchMedia = () => {
    return {
      matches: false,
      media: "",
      onchange: null,
      addListener: () => {},
      removeListener: () => {},
      addEventListener: () => {},
      removeEventListener: () => {},
      dispatchEvent: () => false
    } as unknown as MediaQueryList;
  };
}

describe("MhAnchor (vitest)", () => {
  it("renders anchor with items", () => {
    const items = [
      { key: "1", href: "#section1", title: "Section 1" },
      { key: "2", href: "#section2", title: "Section 2" }
    ];
    render(<MhAnchor items={items} />);
    expect(screen.getByText("Section 1")).toBeInTheDocument();
    expect(screen.getByText("Section 2")).toBeInTheDocument();
  });

  it("renders with affix prop", () => {
    const items = [{ key: "1", href: "#test", title: "Test" }];
    const { container } = render(<MhAnchor affix={false} items={items} />);
    expect(container.querySelector(".ant-anchor-wrapper")).toBeInTheDocument();
  });
});
