import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { MhSelect } from "./../index";

describe("MhSelect (vitest)", () => {
  it("renders select with options", () => {
    const options = [
      { value: "1", label: "Option 1" },
      { value: "2", label: "Option 2" }
    ];
    const { container } = render(<MhSelect options={options} />);
    expect(container.querySelector(".ant-select")).toBeInTheDocument();
  });

  it("renders with placeholder", () => {
    render(<MhSelect placeholder="Select an option" />);
    expect(screen.getByText("Select an option")).toBeInTheDocument();
  });

  it("renders in multiple mode", () => {
    const { container } = render(<MhSelect mode="multiple" />);
    expect(container.querySelector(".ant-select-multiple")).toBeInTheDocument();
  });
});
