import { render, screen } from "@testing-library/react";
import { beforeAll, describe, expect, it } from "vitest";
import { MhAvatar } from "./../index";

describe("MhAvatar (vitest)", () => {
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
  it("renders with text", () => {
    render(<MhAvatar>U</MhAvatar>);
    expect(screen.getByText("U")).toBeInTheDocument();
  });

  it("renders with src prop", () => {
    render(<MhAvatar src="https://example.com/avatar.jpg" alt="User Avatar" />);
    const img = screen.getByRole("img");
    expect(img).toBeInTheDocument();
    expect(img).toHaveAttribute("src", "https://example.com/avatar.jpg");
  });

  it("applies size prop", () => {
    const { container } = render(<MhAvatar size="large">L</MhAvatar>);
    expect(container.querySelector(".ant-avatar-lg")).toBeInTheDocument();
  });

  it("renders Avatar.Group", () => {
    const { container } = render(
      <MhAvatar.Group>
        <MhAvatar>A</MhAvatar>
        <MhAvatar>B</MhAvatar>
      </MhAvatar.Group>
    );
    expect(container.querySelector(".ant-avatar-group")).toBeInTheDocument();
  });
});
