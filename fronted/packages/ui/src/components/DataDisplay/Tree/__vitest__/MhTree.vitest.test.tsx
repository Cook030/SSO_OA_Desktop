import { render, screen } from "@testing-library/react";
import { beforeAll, describe, expect, it } from "vitest";
import { MhTree } from "./../index";

describe("MhTree (vitest)", () => {
  beforeAll(() => {
    global.ResizeObserver = class ResizeObserver {
      observe() {}
      unobserve() {}
      disconnect() {}
    };
  });
  it("renders tree with data", () => {
    const treeData = [
      {
        title: "Parent",
        key: "0-0",
        children: [
          { title: "Child 1", key: "0-0-0" },
          { title: "Child 2", key: "0-0-1" }
        ]
      }
    ];
    render(<MhTree treeData={treeData} />);
    expect(screen.getByText("Parent")).toBeInTheDocument();
  });

  it("renders with checkable prop", () => {
    const treeData = [{ title: "Node", key: "0" }];
    const { container } = render(<MhTree checkable treeData={treeData} />);
    expect(container.querySelector(".ant-tree-checkbox")).toBeInTheDocument();
  });
});
