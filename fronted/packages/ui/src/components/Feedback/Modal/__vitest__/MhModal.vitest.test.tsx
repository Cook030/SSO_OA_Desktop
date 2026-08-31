import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { MhModal } from "./../index";

describe("MhModal (vitest)", () => {
  it("renders modal when open", () => {
    render(
      <MhModal open title="Modal Title">
        Modal Content
      </MhModal>
    );
    expect(screen.getByText("Modal Title")).toBeInTheDocument();
    expect(screen.getByText("Modal Content")).toBeInTheDocument();
  });

  it("does not render when closed", () => {
    const { container } = render(
      <MhModal open={false} title="Hidden">
        Content
      </MhModal>
    );
    expect(container.querySelector(".ant-modal")).not.toBeInTheDocument();
  });

  it("renders with footer", () => {
    render(
      <MhModal open footer={<button>Custom Footer</button>}>
        Content
      </MhModal>
    );
    expect(screen.getByText("Custom Footer")).toBeInTheDocument();
  });
});
