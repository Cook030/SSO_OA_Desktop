import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { MhDivider } from "./../index";

describe("MhDivider (vitest)", () => {
  it("renders divider component", () => {
    const { container } = render(<MhDivider />);
    expect(container.querySelector(".ant-divider")).toBeInTheDocument();
  });

  it("renders with text", () => {
    render(<MhDivider>Divider Text</MhDivider>);
    expect(screen.getByText("Divider Text")).toBeInTheDocument();
  });

  it("applies orientation prop", () => {
    const { container } = render(<MhDivider orientation="horizontal">Left</MhDivider>);
    expect(container.querySelector(".ant-divider-with-text")).toBeInTheDocument();
    expect(screen.getByText("Left")).toBeInTheDocument();
  });

  it("renders vertical divider", () => {
    const { container } = render(<MhDivider orientation="vertical" />);
    expect(container.querySelector(".ant-divider-vertical")).toBeInTheDocument();
  });
});
