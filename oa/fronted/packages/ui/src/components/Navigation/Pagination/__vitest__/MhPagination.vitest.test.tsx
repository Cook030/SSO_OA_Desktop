import { render } from "@testing-library/react";
import { beforeAll, describe, expect, it } from "vitest";
import { MhPagination } from "./../index";

describe("MhPagination (vitest)", () => {
  beforeAll(() => {
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
  });
  it("renders pagination component", () => {
    const { container } = render(<MhPagination total={50} />);
    expect(container.querySelector(".ant-pagination")).toBeInTheDocument();
  });

  it("renders with current page", () => {
    const { container } = render(<MhPagination total={100} current={2} />);
    expect(container.querySelector(".ant-pagination-item-active")).toBeInTheDocument();
  });

  it("renders with page size", () => {
    const { container } = render(<MhPagination total={100} pageSize={20} />);
    expect(container.querySelector(".ant-pagination")).toBeInTheDocument();
  });
});
