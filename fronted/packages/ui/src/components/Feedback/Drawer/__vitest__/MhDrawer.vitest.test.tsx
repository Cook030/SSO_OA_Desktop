import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { MhDrawer } from "./../index";

describe("MhDrawer (vitest)", () => {
  it("renders drawer when open", () => {
    render(
      <MhDrawer open title="Drawer Title">
        Drawer Content
      </MhDrawer>
    );
    expect(screen.getByText("Drawer Title")).toBeInTheDocument();
    expect(screen.getByText("Drawer Content")).toBeInTheDocument();
  });

  it("does not render when closed", () => {
    const { container } = render(<MhDrawer open={false}>Content</MhDrawer>);
    expect(container.querySelector(".ant-drawer-open")).not.toBeInTheDocument();
  });

  it("renders with placement prop", () => {
    render(
      <MhDrawer open placement="right">
        Content
      </MhDrawer>
    );
    expect(screen.getByText("Content")).toBeInTheDocument();
  });
});
