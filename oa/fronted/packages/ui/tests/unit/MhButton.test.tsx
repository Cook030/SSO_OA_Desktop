import { render, screen } from "@testing-library/react";
import { MhButton } from "../../src/components/General/Button";

describe("MhButton", () => {
  it("renders with children", () => {
    render(<MhButton>Click me</MhButton>);
    expect(screen.getByText("Click me")).toBeInTheDocument();
  });

  it("renders with prefix", () => {
    render(<MhButton>Click me</MhButton>);
    expect(screen.getByText("👉")).toBeInTheDocument();
    expect(screen.getByText("Button")).toBeInTheDocument();
  });

  it("renders with suffix", () => {
    render(<MhButton mhProps={{ prefix: "👈" }}>Button</MhButton>);
    expect(screen.getByText("👈")).toBeInTheDocument();
    expect(screen.getByText("Button")).toBeInTheDocument();
  });

  it("renders with both prefix and suffix", () => {
    render(<MhButton mhProps={{ prefix: "✨", suffix: "🚀" }}>Button</MhButton>);
    expect(screen.getByText("✨")).toBeInTheDocument();
    expect(screen.getByText("Button")).toBeInTheDocument();
    expect(screen.getByText("🚀")).toBeInTheDocument();
  });

  it("renders with loading state", () => {
    render(<MhButton loading>Loading</MhButton>);
    expect(screen.getByRole("button")).toBeInTheDocument();
  });

  it("applies correct button type", () => {
    render(<MhButton type="primary">Primary</MhButton>);
    const button = screen.getByRole("button");
    expect(button).toHaveClass("ant-btn-primary");
  });
});
