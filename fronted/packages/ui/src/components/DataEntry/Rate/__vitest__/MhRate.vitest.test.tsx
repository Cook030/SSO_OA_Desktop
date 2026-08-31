import { render } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { MhRate } from "./../index";

describe("MhRate (vitest)", () => {
  it("renders rate component", () => {
    const { container } = render(<MhRate />);
    expect(container.querySelector(".ant-rate")).toBeInTheDocument();
  });

  it("renders with default value", () => {
    const { container } = render(<MhRate defaultValue={3} />);
    expect(container.querySelector(".ant-rate")).toBeInTheDocument();
  });

  it("renders with custom count", () => {
    const { container } = render(<MhRate count={10} />);
    expect(container.querySelectorAll(".ant-rate-star").length).toBe(10);
  });
});
