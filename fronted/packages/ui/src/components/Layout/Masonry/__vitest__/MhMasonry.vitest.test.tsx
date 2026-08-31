import { render } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { MhMasonry } from "./../index";

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

describe("MhMasonry (vitest)", () => {
  it("renders masonry with children", () => {
    const { container } = render(
      <MhMasonry
        itemRender={() => (
          <>
            <div>Item 1</div>
            <div>Item 2</div>
          </>
        )}
      />
    );
    expect(container.querySelector(".ant-masonry")).toBeInTheDocument();
  });

  it("applies columns prop", () => {
    const { container } = render(<MhMasonry columns={3} itemRender={() => <div>Item 1</div>} />);
    expect(container.querySelector(".ant-masonry")).toBeInTheDocument();
  });
});
