import { render } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { MhQRCode } from "./../index";

describe("MhQRCode (vitest)", () => {
  it("renders QR code with value", () => {
    const { container } = render(<MhQRCode value="https://example.com" />);
    expect(container.querySelector(".ant-qrcode")).toBeInTheDocument();
  });

  it("renders with custom size", () => {
    const { container } = render(<MhQRCode value="test" size={200} />);
    expect(container.querySelector(".ant-qrcode")).toBeInTheDocument();
  });
});
