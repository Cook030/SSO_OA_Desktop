import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { MhPopconfirm } from "./../index";

describe("MhPopconfirm (vitest)", () => {
  it("renders children", () => {
    render(
      <MhPopconfirm title="Are you sure?">
        <button>Delete</button>
      </MhPopconfirm>
    );
    expect(screen.getByText("Delete")).toBeInTheDocument();
  });

  it("renders with description", () => {
    render(
      <MhPopconfirm title="Confirm" description="This action cannot be undone">
        <button>Action</button>
      </MhPopconfirm>
    );
    expect(screen.getByText("Action")).toBeInTheDocument();
  });
});
