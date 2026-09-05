import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { MhCollapse } from "./../index";

describe("MhCollapse (vitest)", () => {
  it("renders collapse items", () => {
    const items = [
      { key: "1", label: "Panel 1", children: "Content 1" },
      { key: "2", label: "Panel 2", children: "Content 2" }
    ];
    render(<MhCollapse items={items} />);
    expect(screen.getByText("Panel 1")).toBeInTheDocument();
    expect(screen.getByText("Panel 2")).toBeInTheDocument();
  });

  it("applies accordion prop", () => {
    const items = [{ key: "1", label: "Panel", children: "Content" }];
    const { container } = render(<MhCollapse accordion items={items} />);
    expect(container.querySelector(".ant-collapse")).toBeInTheDocument();
  });
});
