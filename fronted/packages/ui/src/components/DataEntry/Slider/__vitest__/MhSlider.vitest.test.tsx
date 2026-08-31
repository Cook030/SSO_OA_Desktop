import { render } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { MhSlider } from "./../index";

describe("MhSlider (vitest)", () => {
  it("renders slider component", () => {
    const { container } = render(<MhSlider />);
    expect(container.querySelector(".ant-slider")).toBeInTheDocument();
  });

  it("renders with default value", () => {
    const { container } = render(<MhSlider defaultValue={30} />);
    expect(container.querySelector(".ant-slider")).toBeInTheDocument();
  });

  it("renders range slider", () => {
    const { container } = render(<MhSlider range={true} defaultValue={[20, 50]} />);
    expect(container.querySelector(".ant-slider")).toBeInTheDocument();
  });
});
