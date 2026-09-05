import { render, screen } from "@testing-library/react";
import { beforeAll, describe, expect, it } from "vitest";
import { MhTimeline } from "./../index";

describe("MhTimeline (vitest)", () => {
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
  it("renders timeline with items", () => {
    const items = [{ children: "Event 1" }, { children: "Event 2" }, { children: "Event 3" }];
    render(<MhTimeline items={items} />);
    expect(screen.getByText("Event 1")).toBeInTheDocument();
    expect(screen.getByText("Event 2")).toBeInTheDocument();
  });

  it("renders with mode prop", () => {
    const items = [{ children: "Event" }];
    const { container } = render(<MhTimeline mode="alternate" items={items} />);
    expect(container.querySelector(".ant-timeline")).toBeInTheDocument();
  });
});
