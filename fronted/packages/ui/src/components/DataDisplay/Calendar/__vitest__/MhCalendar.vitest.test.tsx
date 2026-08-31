import { render } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { MhCalendar } from "./../index";

describe("MhCalendar (vitest)", () => {
  it("renders calendar component", () => {
    const { container } = render(<MhCalendar />);
    expect(container.querySelector(".ant-picker-calendar")).toBeInTheDocument();
  });

  it("renders in fullscreen mode by default", () => {
    const { container } = render(<MhCalendar />);
    expect(container.querySelector(".ant-picker-calendar-full")).toBeInTheDocument();
  });

  it("renders in card mode when fullscreen is false", () => {
    const { container } = render(<MhCalendar fullscreen={false} />);
    expect(container.querySelector(".ant-picker-calendar-full")).not.toBeInTheDocument();
  });
});
