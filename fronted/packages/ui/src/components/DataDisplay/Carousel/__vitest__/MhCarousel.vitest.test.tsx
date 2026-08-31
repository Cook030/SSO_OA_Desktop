import { render, screen } from "@testing-library/react";
import { beforeAll, describe, expect, it } from "vitest";
import { MhCarousel } from "./../index";

describe("MhCarousel (vitest)", () => {
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
  it("renders carousel with children", () => {
    render(
      <MhCarousel>
        <div>Slide 1</div>
        <div>Slide 2</div>
      </MhCarousel>
    );
    expect(screen.getAllByText("Slide 1").length).toBeGreaterThan(0);
  });

  it("renders with autoplay prop", () => {
    const { container } = render(
      <MhCarousel autoplay>
        <div>Slide 1</div>
      </MhCarousel>
    );
    expect(container.querySelector(".slick-slider")).toBeInTheDocument();
  });
});
