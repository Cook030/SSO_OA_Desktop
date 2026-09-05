import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { MhLayout } from "./../index";

describe("MhLayout (vitest)", () => {
  it("renders layout with children", () => {
    render(<MhLayout>Layout Content</MhLayout>);
    expect(screen.getByText("Layout Content")).toBeInTheDocument();
  });

  it("renders Header component", () => {
    render(<MhLayout.Header>Header Content</MhLayout.Header>);
    expect(screen.getByText("Header Content")).toBeInTheDocument();
  });

  it("renders Sider component", () => {
    render(<MhLayout.Sider>Sider Content</MhLayout.Sider>);
    expect(screen.getByText("Sider Content")).toBeInTheDocument();
  });

  it("renders Content component", () => {
    render(<MhLayout.Content>Main Content</MhLayout.Content>);
    expect(screen.getByText("Main Content")).toBeInTheDocument();
  });

  it("renders Footer component", () => {
    render(<MhLayout.Footer>Footer Content</MhLayout.Footer>);
    expect(screen.getByText("Footer Content")).toBeInTheDocument();
  });
});
