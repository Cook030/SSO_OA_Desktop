import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { MhRadio } from "./../index";

describe("MhRadio (vitest)", () => {
  it("renders radio with label", () => {
    render(<MhRadio>Radio Label</MhRadio>);
    expect(screen.getByText("Radio Label")).toBeInTheDocument();
  });

  it("renders Radio.Group with options", () => {
    const options = [
      { label: "Option 1", value: "1" },
      { label: "Option 2", value: "2" }
    ];
    render(<MhRadio.Group options={options} />);
    expect(screen.getByText("Option 1")).toBeInTheDocument();
    expect(screen.getByText("Option 2")).toBeInTheDocument();
  });

  it("renders checked radio", () => {
    const { container } = render(<MhRadio checked>Checked</MhRadio>);
    expect(container.querySelector(".ant-radio-checked")).toBeInTheDocument();
  });
});
