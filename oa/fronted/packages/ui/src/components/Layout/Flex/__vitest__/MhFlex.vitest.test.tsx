import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { MhFlex } from "./../index";

describe("MhFlex (vitest)", () => {
  it("renders flex container with children", () => {
    render(
      <MhFlex>
        <div>Child 1</div>
        <div>Child 2</div>
      </MhFlex>
    );
    expect(screen.getByText("Child 1")).toBeInTheDocument();
    expect(screen.getByText("Child 2")).toBeInTheDocument();
  });

  it("applies gap prop", () => {
    const { container } = render(
      <MhFlex gap="large">
        <div>Content</div>
      </MhFlex>
    );
    expect(container.querySelector(".ant-flex")).toBeInTheDocument();
  });

  it("applies vertical prop", () => {
    const { container } = render(
      <MhFlex vertical>
        <div>Content</div>
      </MhFlex>
    );
    expect(container.querySelector(".ant-flex-vertical")).toBeInTheDocument();
  });
});
