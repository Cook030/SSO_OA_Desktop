import { render, screen } from "@testing-library/react";
import { beforeAll, describe, expect, it } from "vitest";
import { MhAffix } from "./../index";

describe("MhAffix (vitest)", () => {
  beforeAll(() => {
    global.ResizeObserver = class ResizeObserver {
      observe() {}
      unobserve() {}
      disconnect() {}
    };
  });
  it("renders affix with children", () => {
    render(
      <MhAffix>
        <button>Fixed Button</button>
      </MhAffix>
    );
    expect(screen.getByText("Fixed Button")).toBeInTheDocument();
  });

  it("renders with offsetTop prop", () => {
    const { container } = render(
      <MhAffix offsetTop={10}>
        <div>Content</div>
      </MhAffix>
    );
    expect(container.querySelector(".ant-affix")).toBeDefined();
  });
});
