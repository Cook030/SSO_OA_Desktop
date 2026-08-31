import { render } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { MhIcon } from "./../index";

describe("MhIcon (vitest)", () => {
  it("renders icon component", () => {
    const { container } = render(<MhIcon component={() => <span>Icon</span>} />);
    expect(container.querySelector(".anticon")).toBeInTheDocument();
  });

  it("applies style prop", () => {
    const { container } = render(<MhIcon style={{ fontSize: "24px" }} />);
    expect(container.querySelector(".anticon")).toBeInTheDocument();
  });
});
