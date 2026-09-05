import { render } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { MhSkeleton } from "./../index";

describe("MhSkeleton (vitest)", () => {
  it("renders skeleton component", () => {
    const { container } = render(<MhSkeleton />);
    expect(container.querySelector(".ant-skeleton")).toBeInTheDocument();
  });

  it("renders with avatar", () => {
    const { container } = render(<MhSkeleton avatar />);
    expect(container.querySelector(".ant-skeleton-avatar")).toBeInTheDocument();
  });

  it("renders active skeleton", () => {
    const { container } = render(<MhSkeleton active />);
    expect(container.querySelector(".ant-skeleton-active")).toBeInTheDocument();
  });
});
