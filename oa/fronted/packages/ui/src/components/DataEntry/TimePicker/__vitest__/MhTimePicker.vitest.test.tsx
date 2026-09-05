import { render } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { MhTimePicker } from "./../index";

describe("MhTimePicker (vitest)", () => {
  it("renders time picker", () => {
    const { container } = render(<MhTimePicker />);
    expect(container.querySelector(".ant-picker")).toBeInTheDocument();
  });

  it("renders with placeholder", () => {
    const { container } = render(<MhTimePicker placeholder="Select time" />);
    expect(container.querySelector(".ant-picker")).toBeInTheDocument();
  });
});
