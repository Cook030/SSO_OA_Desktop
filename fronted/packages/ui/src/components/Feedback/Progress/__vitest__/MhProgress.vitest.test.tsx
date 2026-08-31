import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { MhProgress } from "./../index";

describe("MhProgress (vitest)", () => {
  it("renders progress with percent", () => {
    render(<MhProgress percent={50} />);
    expect(screen.getByText("50%")).toBeInTheDocument();
  });

  it("applies type prop", () => {
    const { container } = render(<MhProgress percent={75} type="circle" />);
    expect(container.querySelector(".ant-progress-circle")).toBeInTheDocument();
  });

  it("applies status prop", () => {
    const { container } = render(<MhProgress percent={100} status="success" />);
    expect(container.querySelector(".ant-progress-status-success")).toBeInTheDocument();
  });
});
