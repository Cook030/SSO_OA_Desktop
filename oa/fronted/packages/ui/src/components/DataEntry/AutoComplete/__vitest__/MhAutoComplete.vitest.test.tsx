import { render } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { MhAutoComplete } from "./../index";

describe("MhAutoComplete (vitest)", () => {
  it("renders autocomplete component", () => {
    const options = [{ value: "Option 1" }, { value: "Option 2" }];
    const { container } = render(<MhAutoComplete options={options} />);
    expect(container.querySelector(".ant-select")).toBeInTheDocument();
  });

  it("renders with placeholder", () => {
    const { container } = render(<MhAutoComplete placeholder="Search..." />);
    expect(container.querySelector(".ant-select")).toBeInTheDocument();
  });
});
