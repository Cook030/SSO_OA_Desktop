import { render, screen } from "@testing-library/react";
import { beforeAll, describe, expect, it } from "vitest";
import { MhTabs } from "./../index";

describe("MhTabs (vitest)", () => {
  beforeAll(() => {
    global.ResizeObserver = class ResizeObserver {
      observe() {}
      unobserve() {}
      disconnect() {}
    };
  });
  it("renders tabs with items", () => {
    const items = [
      { key: "1", label: "Tab 1", children: "Content 1" },
      { key: "2", label: "Tab 2", children: "Content 2" }
    ];
    render(<MhTabs items={items} />);
    expect(screen.getByText("Tab 1")).toBeInTheDocument();
    expect(screen.getByText("Tab 2")).toBeInTheDocument();
  });

  it("renders with default active key", () => {
    const items = [
      { key: "1", label: "Tab 1", children: "Content 1" },
      { key: "2", label: "Tab 2", children: "Content 2" }
    ];
    render(<MhTabs defaultActiveKey="2" items={items} />);
    expect(screen.getByText("Content 2")).toBeInTheDocument();
  });
});
