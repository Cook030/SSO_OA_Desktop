import { render } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { MhInputNumber } from "./../index";

describe("MhInputNumber (vitest)", () => {
  it("renders input number", () => {
    const { container } = render(<MhInputNumber />);
    expect(container.querySelector(".ant-input-number")).toBeInTheDocument();
  });

  it("renders with default value", () => {
    const { container } = render(<MhInputNumber defaultValue={10} />);
    expect(container.querySelector(".ant-input-number")).toBeInTheDocument();
  });

  it("renders with min and max", () => {
    const { container } = render(<MhInputNumber min={0} max={100} />);
    expect(container.querySelector(".ant-input-number")).toBeInTheDocument();
  });
});
