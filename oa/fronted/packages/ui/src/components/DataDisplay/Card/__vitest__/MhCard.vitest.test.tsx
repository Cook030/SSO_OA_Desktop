import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { MhCard } from "./../index";

describe("MhCard (vitest)", () => {
  it("renders children", () => {
    render(<MhCard>Card Content</MhCard>);
    expect(screen.getByText("Card Content")).toBeInTheDocument();
  });

  it("renders with title", () => {
    render(<MhCard title="Card Title">Content</MhCard>);
    expect(screen.getByText("Card Title")).toBeInTheDocument();
  });

  it("applies bordered prop", () => {
    const { container } = render(<MhCard bordered={false}>Content</MhCard>);
    expect(container.querySelector(".ant-card-bordered")).not.toBeInTheDocument();
  });

  it("renders Card.Meta", () => {
    render(
      <MhCard>
        <MhCard.Meta title="Meta Title" description="Meta Description" />
      </MhCard>
    );
    expect(screen.getByText("Meta Title")).toBeInTheDocument();
    expect(screen.getByText("Meta Description")).toBeInTheDocument();
  });
});
