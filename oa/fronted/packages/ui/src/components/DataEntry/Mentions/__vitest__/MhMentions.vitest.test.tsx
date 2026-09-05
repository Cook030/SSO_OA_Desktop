import { render } from "@testing-library/react";
import { beforeAll, describe, expect, it } from "vitest";
import { MhMentions } from "./../index";

describe("MhMentions (vitest)", () => {
  beforeAll(() => {
    global.ResizeObserver = class ResizeObserver {
      observe() {}
      unobserve() {}
      disconnect() {}
    };
  });
  it("renders mentions component", () => {
    const options = [
      { value: "user1", label: "User 1" },
      { value: "user2", label: "User 2" }
    ];
    const { container } = render(<MhMentions options={options} />);
    expect(container.querySelector(".ant-mentions")).toBeInTheDocument();
  });

  it("renders with placeholder", () => {
    const { container } = render(<MhMentions placeholder="Mention someone" />);
    expect(container.querySelector(".ant-mentions")).toBeInTheDocument();
  });
});
