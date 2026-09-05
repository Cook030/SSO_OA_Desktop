import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { MhBadge } from "./../index";

describe("MhBadge (vitest)", () => {
  it("renders with count", () => {
    render(
      <MhBadge count={5}>
        <span>Badge</span>
      </MhBadge>
    );
    expect(screen.getByText("5")).toBeInTheDocument();
  });

  it("renders with dot", () => {
    const { container } = render(
      <MhBadge dot>
        <span>Badge</span>
      </MhBadge>
    );
    expect(container.querySelector(".ant-badge-dot")).toBeInTheDocument();
  });

  it("renders children", () => {
    render(
      <MhBadge count={1}>
        <span>Content</span>
      </MhBadge>
    );
    expect(screen.getByText("Content")).toBeInTheDocument();
  });

  it("applies status prop", () => {
    const { container } = render(<MhBadge status="success" text="Success" />);
    expect(container.querySelector(".ant-badge-status-success")).toBeInTheDocument();
  });
});
