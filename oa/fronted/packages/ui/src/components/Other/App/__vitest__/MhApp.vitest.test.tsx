import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { MhApp } from "./../index";

describe("MhApp (vitest)", () => {
  it("renders app container with children", () => {
    render(
      <MhApp>
        <div>App Content</div>
      </MhApp>
    );
    expect(screen.getByText("App Content")).toBeInTheDocument();
  });

  it("renders with message and notification", () => {
    const { container } = render(
      <MhApp>
        <div>Content</div>
      </MhApp>
    );
    expect(container.querySelector(".ant-app")).toBeInTheDocument();
  });
});
