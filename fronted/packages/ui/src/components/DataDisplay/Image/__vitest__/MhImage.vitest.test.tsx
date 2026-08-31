import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { MhImage } from "./../index";

describe("MhImage (vitest)", () => {
  it("renders image with src", () => {
    render(<MhImage src="https://example.com/image.jpg" alt="Test Image" />);
    const img = screen.getByRole("img");
    expect(img).toBeInTheDocument();
    expect(img).toHaveAttribute("src", "https://example.com/image.jpg");
  });

  it("renders with width and height", () => {
    render(<MhImage src="test.jpg" width={200} height={200} alt="Test" />);
    const img = screen.getByRole("img");
    expect(img).toBeInTheDocument();
  });
});
