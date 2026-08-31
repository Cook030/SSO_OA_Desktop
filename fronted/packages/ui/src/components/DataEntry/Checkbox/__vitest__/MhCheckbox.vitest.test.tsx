import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { MhCheckbox } from "./../index";

describe("MhCheckbox (vitest)", () => {
  it("renders checkbox with label", () => {
    render(<MhCheckbox>Checkbox Label</MhCheckbox>);
    expect(screen.getByText("Checkbox Label")).toBeInTheDocument();
  });

  it("renders checked checkbox", () => {
    const { container } = render(<MhCheckbox checked>Checked</MhCheckbox>);
    expect(container.querySelector(".ant-checkbox-checked")).toBeInTheDocument();
  });

  it("renders disabled checkbox", () => {
    const { container } = render(<MhCheckbox disabled>Disabled</MhCheckbox>);
    expect(container.querySelector(".ant-checkbox-disabled")).toBeInTheDocument();
  });

  it("renders Checkbox.Group", () => {
    const options = ["Option 1", "Option 2"];
    render(<MhCheckbox.Group options={options} />);
    expect(screen.getByText("Option 1")).toBeInTheDocument();
    expect(screen.getByText("Option 2")).toBeInTheDocument();
  });
});
