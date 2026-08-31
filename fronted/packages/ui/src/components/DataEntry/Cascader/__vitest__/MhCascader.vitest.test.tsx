import { render } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { MhCascader } from "./../index";

describe("MhCascader (vitest)", () => {
  it("renders cascader component", () => {
    const options = [
      {
        value: "parent",
        label: "Parent",
        children: [{ value: "child", label: "Child" }]
      }
    ];
    const { container } = render(<MhCascader options={options} />);
    expect(container.querySelector(".ant-cascader")).toBeInTheDocument();
  });

  it("renders with placeholder", () => {
    const { container } = render(<MhCascader placeholder="Select" />);
    expect(container.querySelector(".ant-cascader")).toBeInTheDocument();
  });
});
