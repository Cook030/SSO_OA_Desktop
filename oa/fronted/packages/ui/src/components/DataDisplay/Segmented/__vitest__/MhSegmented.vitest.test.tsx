import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { MhSegmented } from "./../index";

describe("MhSegmented (vitest)", () => {
  it("renders with options", () => {
    render(<MhSegmented options={["Option 1", "Option 2", "Option 3"]} />);
    expect(screen.getByText("Option 1")).toBeInTheDocument();
    expect(screen.getByText("Option 2")).toBeInTheDocument();
    expect(screen.getByText("Option 3")).toBeInTheDocument();
  });

  it("renders with default value", () => {
    const { container } = render(<MhSegmented options={["A", "B", "C"]} defaultValue="B" />);
    expect(container.querySelector(".ant-segmented")).toBeInTheDocument();
  });
});
