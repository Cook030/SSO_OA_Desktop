import { render } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { MhTreeSelect } from "./../index";

describe("MhTreeSelect (vitest)", () => {
  it("renders tree select component", () => {
    const treeData = [
      {
        title: "Parent",
        value: "parent",
        children: [{ title: "Child", value: "child" }]
      }
    ];
    const { container } = render(<MhTreeSelect treeData={treeData} />);
    expect(container.querySelector(".ant-select")).toBeInTheDocument();
  });

  it("renders with placeholder", () => {
    const { container } = render(<MhTreeSelect placeholder="Select" />);
    expect(container.querySelector(".ant-select")).toBeInTheDocument();
  });
});
