import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { MhTypography } from "./../index";

describe("MhTypography (vitest)", () => {
  it("renders typography text", () => {
    render(<MhTypography.Text>Text content</MhTypography.Text>);
    expect(screen.getByText("Text content")).toBeInTheDocument();
  });

  it("renders typography title", () => {
    render(<MhTypography.Title>Title content</MhTypography.Title>);
    expect(screen.getByText("Title content")).toBeInTheDocument();
  });

  it("renders typography paragraph", () => {
    render(<MhTypography.Paragraph>Paragraph content</MhTypography.Paragraph>);
    expect(screen.getByText("Paragraph content")).toBeInTheDocument();
  });

  it("renders typography link", () => {
    render(<MhTypography.Link href="#">Link text</MhTypography.Link>);
    expect(screen.getByText("Link text")).toBeInTheDocument();
  });
});
