import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { MhUpload } from "./../index";

describe("MhUpload (vitest)", () => {
  it("renders upload component", () => {
    const { container } = render(
      <MhUpload>
        <button>Upload</button>
      </MhUpload>
    );
    expect(container.querySelector(".ant-upload")).toBeInTheDocument();
    expect(screen.getByText("Upload")).toBeInTheDocument();
  });

  it("renders with file list", () => {
    const fileList = [{ uid: "1", name: "file1.txt", status: "done" as const }];
    const { container } = render(<MhUpload fileList={fileList} />);
    expect(container.querySelector(".ant-upload-list")).toBeInTheDocument();
  });
});
