import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { MhAlert } from "./../index";

describe("MhAlert (vitest)", () => {
  it("renders alert with message", () => {
    render(<MhAlert message="Alert message" />);
    expect(screen.getByText("Alert message")).toBeInTheDocument();
  });

  it("applies type prop", () => {
    const { container } = render(<MhAlert message="Success" type="success" />);
    expect(container.querySelector(".ant-alert-success")).toBeInTheDocument();
  });

  it("renders with description", () => {
    render(<MhAlert message="Title" description="Description text" />);
    expect(screen.getByText("Description text")).toBeInTheDocument();
  });

  it("renders closable alert", () => {
    const { container } = render(<MhAlert message="Closable" closable />);
    expect(container.querySelector(".ant-alert-close-icon")).toBeInTheDocument();
  });
});
