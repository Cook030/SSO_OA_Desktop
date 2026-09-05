import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { MhTag } from "./../index";

describe("MhTag (vitest)", () => {
  it("renders children", () => {
    render(<MhTag>Tag Content</MhTag>);
    expect(screen.getByText("Tag Content")).toBeInTheDocument();
  });

  it("applies color prop", () => {
    const { container } = render(<MhTag color="blue">Blue Tag</MhTag>);
    expect(container.querySelector(".ant-tag-blue")).toBeInTheDocument();
  });

  it("renders closable tag", () => {
    const { container } = render(<MhTag closable>Closable</MhTag>);
    expect(container.querySelector(".ant-tag-close-icon")).toBeInTheDocument();
  });
});
