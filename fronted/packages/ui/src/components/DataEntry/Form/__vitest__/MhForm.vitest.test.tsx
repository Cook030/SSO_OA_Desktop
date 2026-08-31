import { render, screen } from "@testing-library/react";
import { beforeAll, describe, expect, it } from "vitest";
import { MhForm } from "./../index";

describe("MhForm (vitest)", () => {
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
  it("renders form component", () => {
    const { container } = render(
      <MhForm>
        <MhForm.Item label="Username">
          <input />
        </MhForm.Item>
      </MhForm>
    );
    expect(container.querySelector(".ant-form")).toBeInTheDocument();
    expect(screen.getByText("Username")).toBeInTheDocument();
  });

  it("renders with layout prop", () => {
    const { container } = render(<MhForm layout="horizontal" />);
    expect(container.querySelector(".ant-form-horizontal")).toBeInTheDocument();
  });

  it("renders Form.Item", () => {
    render(
      <MhForm>
        <MhForm.Item label="Email" name="email">
          <input />
        </MhForm.Item>
      </MhForm>
    );
    expect(screen.getByText("Email")).toBeInTheDocument();
  });
});
