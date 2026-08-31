import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { MhBreadcrumb } from "./../index";

describe("MhBreadcrumb (vitest)", () => {
  it("renders breadcrumb with items", () => {
    const items = [{ title: "Home" }, { title: "Products" }, { title: "Details" }];
    render(<MhBreadcrumb items={items} />);
    expect(screen.getByText("Home")).toBeInTheDocument();
    expect(screen.getByText("Products")).toBeInTheDocument();
  });

  it("renders with separator", () => {
    const items = [{ title: "Home" }, { title: "Page" }];
    const { container } = render(<MhBreadcrumb separator=">" items={items} />);
    expect(container.querySelector(".ant-breadcrumb")).toBeInTheDocument();
  });
});
