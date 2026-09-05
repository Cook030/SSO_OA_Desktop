import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { MhResult } from "./../index";

describe("MhResult (vitest)", () => {
  it("renders result with title", () => {
    render(<MhResult title="Success" />);
    expect(screen.getByText("Success")).toBeInTheDocument();
  });

  it("applies status prop", () => {
    const { container } = render(<MhResult status="success" title="Done" />);
    expect(container.querySelector(".ant-result-success")).toBeInTheDocument();
  });

  it("renders with subtitle", () => {
    render(<MhResult title="Title" subTitle="Subtitle text" />);
    expect(screen.getByText("Subtitle text")).toBeInTheDocument();
  });
});
