import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { MhWatermark } from "./../index";

describe("MhWatermark (vitest)", () => {
  it("renders watermark with content", () => {
    render(
      <MhWatermark content="Watermark Text">
        <div>Protected Content</div>
      </MhWatermark>
    );
    expect(screen.getByText("Protected Content")).toBeInTheDocument();
  });

  it("renders with children", () => {
    render(
      <MhWatermark content="Mark">
        <div style={{ height: 500 }}>Content</div>
      </MhWatermark>
    );
    expect(screen.getByText("Content")).toBeInTheDocument();
  });
});
