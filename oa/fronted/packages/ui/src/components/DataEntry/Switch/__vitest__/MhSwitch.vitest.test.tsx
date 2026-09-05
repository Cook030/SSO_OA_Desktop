import { render } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { MhSwitch } from "./../index";

describe("MhSwitch (vitest)", () => {
  it("renders switch component", () => {
    const { container } = render(<MhSwitch />);
    expect(container.querySelector(".ant-switch")).toBeInTheDocument();
  });

  it("renders checked switch", () => {
    const { container } = render(<MhSwitch checked />);
    expect(container.querySelector(".ant-switch-checked")).toBeInTheDocument();
  });

  it("renders disabled switch", () => {
    const { container } = render(<MhSwitch disabled />);
    expect(container.querySelector(".ant-switch-disabled")).toBeInTheDocument();
  });
});
