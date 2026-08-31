import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { MhEmpty } from "./../index";

describe("MhEmpty (vitest)", () => {
  it("renders empty component", () => {
    const { container } = render(<MhEmpty />);
    expect(container.querySelector(".ant-empty")).toBeInTheDocument();
  });

  it("renders with custom description", () => {
    render(<MhEmpty description="No data available" />);
    expect(screen.getByText("No data available")).toBeInTheDocument();
  });

  it("renders with image", () => {
    const { container } = render(<MhEmpty image="https://example.com/empty.png" />);
    expect(container.querySelector(".ant-empty-image")).toBeInTheDocument();
  });
});
