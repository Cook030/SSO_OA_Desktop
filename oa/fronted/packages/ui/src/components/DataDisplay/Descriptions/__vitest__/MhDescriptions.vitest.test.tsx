import { render, screen } from "@testing-library/react";
import { beforeAll, describe, expect, it } from "vitest";
import { MhDescriptions } from "./../index";

describe("MhDescriptions (vitest)", () => {
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
  it("renders with items", () => {
    const items = [
      { key: "1", label: "Name", children: "John Doe" },
      { key: "2", label: "Age", children: "30" }
    ];
    render(<MhDescriptions items={items} />);
    expect(screen.getByText("Name")).toBeInTheDocument();
    expect(screen.getByText("John Doe")).toBeInTheDocument();
  });

  it("renders with title", () => {
    const items = [{ key: "1", label: "Label", children: "Value" }];
    render(<MhDescriptions title="User Info" items={items} />);
    expect(screen.getByText("User Info")).toBeInTheDocument();
  });
});
