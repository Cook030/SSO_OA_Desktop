import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { MhDropdown } from "./../index";

describe("MhDropdown (vitest)", () => {
  it("renders dropdown with children", () => {
    const items = [
      { key: "1", label: "Menu Item 1" },
      { key: "2", label: "Menu Item 2" }
    ];
    render(
      <MhDropdown menu={{ items }}>
        <button>Dropdown</button>
      </MhDropdown>
    );
    expect(screen.getByText("Dropdown")).toBeInTheDocument();
  });

  it("renders with trigger prop", () => {
    const items = [{ key: "1", label: "Item" }];
    render(
      <MhDropdown menu={{ items }} trigger={["click"]}>
        <span>Click me</span>
      </MhDropdown>
    );
    expect(screen.getByText("Click me")).toBeInTheDocument();
  });
});
