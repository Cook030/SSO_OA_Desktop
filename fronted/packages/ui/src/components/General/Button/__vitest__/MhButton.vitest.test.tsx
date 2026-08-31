import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { MhButton } from "./../index";

describe("MhButton (vitest)", () => {
  it("renders children", () => {
    render(<MhButton>Click me</MhButton>);
    expect(screen.getByRole("button")).toBeInTheDocument();
    expect(screen.getByText("Click me")).toBeInTheDocument();
  });

  it("renders prefix and suffix when provided", () => {
    render(<MhButton mhProps={{ prefix: "P", suffix: "S" }}>Button</MhButton>);

    expect(screen.getByText("P")).toBeInTheDocument();
    expect(screen.getByText("Button")).toBeInTheDocument();
    expect(screen.getByText("S")).toBeInTheDocument();
  });

  it("applies antd type class", () => {
    render(<MhButton type="primary">Primary</MhButton>);
    expect(screen.getByRole("button")).toHaveClass("ant-btn-primary");
  });
});
