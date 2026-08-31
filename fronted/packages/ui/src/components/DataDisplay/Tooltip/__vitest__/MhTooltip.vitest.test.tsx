import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { MhTooltip } from "./../index";

describe("MhTooltip (vitest)", () => {
  it("renders children", () => {
    render(
      <MhTooltip title="Tooltip text">
        <span>Hover me</span>
      </MhTooltip>
    );
    expect(screen.getByText("Hover me")).toBeInTheDocument();
  });

  it("renders with placement prop", () => {
    render(
      <MhTooltip title="Top tooltip" placement="top">
        <button>Button</button>
      </MhTooltip>
    );
    expect(screen.getByText("Button")).toBeInTheDocument();
  });
});
