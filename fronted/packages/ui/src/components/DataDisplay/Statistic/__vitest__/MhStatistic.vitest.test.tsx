import { render, screen } from "@testing-library/react";
import { beforeAll, describe, expect, it } from "vitest";
import { MhStatistic } from "./../index";

describe("MhStatistic (vitest)", () => {
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
  it("renders with title and value", () => {
    render(<MhStatistic title="Active Users" value={112893} />);
    expect(screen.getByText("Active Users")).toBeInTheDocument();
    expect(screen.getByText("112,893")).toBeInTheDocument();
  });

  it("renders with prefix and suffix", () => {
    render(<MhStatistic title="Price" value={99.99} prefix="$" suffix="USD" />);
    expect(screen.getByText("$")).toBeInTheDocument();
    expect(screen.getByText("USD")).toBeInTheDocument();
  });
});
