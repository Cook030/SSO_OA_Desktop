import { render } from "@testing-library/react";
import { beforeAll, describe, expect, it } from "vitest";
import { MhColorPicker } from "./../index";

describe("MhColorPicker (vitest)", () => {
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
  it("renders color picker component", () => {
    const { container } = render(<MhColorPicker />);
    expect(container.firstChild).toBeTruthy();
  });

  it("renders with default value", () => {
    const { container } = render(<MhColorPicker defaultValue="#1890ff" />);
    expect(container.firstChild).toBeTruthy();
  });
});
