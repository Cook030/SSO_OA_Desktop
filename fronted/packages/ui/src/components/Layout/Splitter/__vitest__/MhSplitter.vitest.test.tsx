import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { MhSplitter } from "./../index";

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

describe("MhSplitter (vitest)", () => {
  it("renders splitter with panels", () => {
    const { container } = render(
      <MhSplitter>
        <MhSplitter.Panel>Panel 1</MhSplitter.Panel>
        <MhSplitter.Panel>Panel 2</MhSplitter.Panel>
      </MhSplitter>
    );
    expect(container.querySelector(".ant-splitter")).toBeInTheDocument();
    expect(screen.getByText("Panel 1")).toBeInTheDocument();
    expect(screen.getByText("Panel 2")).toBeInTheDocument();
  });

  it("renders with layout prop", () => {
    const { container } = render(
      <MhSplitter vertical>
        <div>1212</div>
        <MhSplitter.Panel>Content</MhSplitter.Panel>
      </MhSplitter>
    );
    expect(container.querySelector(".ant-splitter")).toBeInTheDocument();
  });
});
