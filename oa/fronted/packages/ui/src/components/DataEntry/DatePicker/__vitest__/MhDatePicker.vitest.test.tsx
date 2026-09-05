import { render } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { MhDatePicker } from "./../index";

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

describe("MhDatePicker (vitest)", () => {
  it("renders date picker", () => {
    const { container } = render(<MhDatePicker />);
    expect(container.querySelector(".ant-picker")).toBeInTheDocument();
  });

  it("renders RangePicker", () => {
    const { container } = render(<MhDatePicker.RangePicker />);
    expect(container.querySelector(".ant-picker-range")).toBeInTheDocument();
  });

  it("renders with placeholder", () => {
    const { container } = render(<MhDatePicker placeholder="Select date" />);
    expect(container.querySelector(".ant-picker")).toBeInTheDocument();
  });
});
