import { render } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { MhTransfer } from "./../index";

describe("MhTransfer (vitest)", () => {
  it("renders transfer component", () => {
    const dataSource = [
      { key: "1", title: "Item 1" },
      { key: "2", title: "Item 2" }
    ];
    const { container } = render(<MhTransfer dataSource={dataSource} />);
    expect(container.querySelector(".ant-transfer")).toBeInTheDocument();
  });

  it("renders with target keys", () => {
    const dataSource = [{ key: "1", title: "Item 1" }];
    const { container } = render(<MhTransfer dataSource={dataSource} targetKeys={["1"]} />);
    expect(container.querySelector(".ant-transfer")).toBeInTheDocument();
  });
});
