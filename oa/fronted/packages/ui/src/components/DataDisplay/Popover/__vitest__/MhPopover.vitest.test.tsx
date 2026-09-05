import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { MhPopover } from "./../index";

describe("MhPopover (vitest)", () => {
  it("renders children", () => {
    render(
      <MhPopover content="Popover content">
        <button>Hover me</button>
      </MhPopover>
    );
    expect(screen.getByText("Hover me")).toBeInTheDocument();
  });

  it("renders with title", () => {
    render(
      <MhPopover title="Title" content="Content">
        <span>Trigger</span>
      </MhPopover>
    );
    expect(screen.getByText("Trigger")).toBeInTheDocument();
  });
});
