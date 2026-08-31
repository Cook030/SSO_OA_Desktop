import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { MhSpace } from "./../index";

describe("MhSpace (vitest)", () => {
  it("renders space with children", () => {
    render(
      <MhSpace>
        <button>Button 1</button>
        <button>Button 2</button>
      </MhSpace>
    );
    expect(screen.getByText("Button 1")).toBeInTheDocument();
    expect(screen.getByText("Button 2")).toBeInTheDocument();
  });

  it("applies orientation prop", () => {
    const { container } = render(
      <MhSpace orientation="vertical">
        <div>Item</div>
      </MhSpace>
    );
    expect(container.querySelector(".ant-space-vertical")).toBeInTheDocument();
  });

  it("applies size prop", () => {
    const { container } = render(
      <MhSpace size="large">
        <div>Item</div>
      </MhSpace>
    );
    expect(container.querySelector(".ant-space")).toBeInTheDocument();
  });
});
