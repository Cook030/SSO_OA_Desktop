import { render } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { MhFloatButton } from "./../index";

describe("MhFloatButton (vitest)", () => {
  it("renders float button", () => {
    const { container } = render(<MhFloatButton />);
    expect(container.querySelector(".ant-float-btn")).toBeInTheDocument();
  });

  it("renders with icon", () => {
    const { container } = render(<MhFloatButton icon={<span>Icon</span>} />);
    expect(container.querySelector(".ant-float-btn")).toBeInTheDocument();
  });

  it("renders FloatButton.Group", () => {
    const { container } = render(
      <MhFloatButton.Group>
        <MhFloatButton />
        <MhFloatButton />
      </MhFloatButton.Group>
    );
    expect(container.querySelector(".ant-float-btn-group")).toBeInTheDocument();
  });
});
