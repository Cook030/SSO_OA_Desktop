import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { MhSpin } from "./../index";

describe("MhSpin (vitest)", () => {
  it("renders spin component", () => {
    const { container } = render(<MhSpin />);
    expect(container.querySelector(".ant-spin")).toBeInTheDocument();
  });

  it("renders with tip text", () => {
    const { container } = render(<MhSpin tip="Loading..." />);
    const spin = container.querySelector(".ant-spin");
    expect(spin).toBeTruthy();
    expect(spin).toHaveAttribute("aria-busy", "true");
    expect(spin).toHaveClass("ant-spin-show-text");
  });

  it("renders with children", () => {
    render(
      <MhSpin>
        <div>Content</div>
      </MhSpin>
    );
    expect(screen.getByText("Content")).toBeInTheDocument();
  });

  it("applies size prop", () => {
    const { container } = render(<MhSpin size="large" />);
    expect(container.querySelector(".ant-spin-lg")).toBeInTheDocument();
  });
});
