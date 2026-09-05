import { render } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { MhTour } from "./../index";

describe("MhTour (vitest)", () => {
  it("renders tour component", () => {
    const steps = [
      { title: "Step 1", description: "Description 1" },
      { title: "Step 2", description: "Description 2" }
    ];
    const { container } = render(<MhTour open={false} steps={steps} />);
    expect(container).toBeInTheDocument();
  });
});
