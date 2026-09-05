import { render, screen } from "@testing-library/react";
import { beforeAll, describe, expect, it } from "vitest";
import { MhTable } from "./../index";

describe("MhTable (vitest)", () => {
  beforeAll(() => {
    if (typeof globalThis.ResizeObserver === "undefined") {
      class ResizeObserver {
        observe() {}
        unobserve() {}
        disconnect() {}
      }

      globalThis.ResizeObserver = ResizeObserver as unknown as typeof globalThis.ResizeObserver;
    }

    if (typeof window !== "undefined" && typeof window.matchMedia === "undefined") {
      window.matchMedia = () => {
        return {
          matches: false,
          media: "",
          onchange: null,
          addListener: () => {},
          removeListener: () => {},
          addEventListener: () => {},
          removeEventListener: () => {},
          dispatchEvent: () => false
        } as unknown as MediaQueryList;
      };
    }
  });
  it("renders table with columns and data", () => {
    const columns = [
      { title: "Name", dataIndex: "name", key: "name" },
      { title: "Age", dataIndex: "age", key: "age" }
    ];
    const dataSource = [
      { key: "1", name: "John", age: 30 },
      { key: "2", name: "Jane", age: 25 }
    ];
    render(<MhTable columns={columns} dataSource={dataSource} />);
    expect(screen.getByText("Name")).toBeInTheDocument();
    expect(screen.getByText("John")).toBeInTheDocument();
  });

  it("renders empty table", () => {
    const columns = [{ title: "Name", dataIndex: "name", key: "name" }];
    const { container } = render(<MhTable columns={columns} dataSource={[]} />);
    expect(container.querySelector(".ant-table")).toBeInTheDocument();
  });
});
