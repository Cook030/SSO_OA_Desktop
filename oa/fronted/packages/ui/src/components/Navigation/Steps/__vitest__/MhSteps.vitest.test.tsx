import { render, screen } from "@testing-library/react";
import { beforeAll, describe, expect, it } from "vitest";
import { MhSteps } from "./../index";

describe("MhSteps (vitest)", () => {
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
  it("renders steps with items", () => {
    const items = [{ title: "Step 1" }, { title: "Step 2" }, { title: "Step 3" }];
    render(<MhSteps items={items} />);
    expect(screen.getByText("Step 1")).toBeInTheDocument();
    expect(screen.getByText("Step 2")).toBeInTheDocument();
  });

  it("renders with current step", () => {
    const items = [{ title: "Step 1" }, { title: "Step 2" }];
    const { container } = render(<MhSteps current={1} items={items} />);
    expect(container.querySelector(".ant-steps")).toBeInTheDocument();
  });
});
