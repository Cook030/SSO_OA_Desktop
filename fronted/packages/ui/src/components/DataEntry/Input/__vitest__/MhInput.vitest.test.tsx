import { fireEvent, render, screen } from "@testing-library/react";
import { beforeAll, describe, expect, it } from "vitest";
import { MhInput } from "./../index";

describe("MhInput (vitest)", () => {
  beforeAll(() => {
    global.ResizeObserver = class ResizeObserver {
      observe() {}
      unobserve() {}
      disconnect() {}
    };
  });
  it("renders input element", () => {
    render(<MhInput placeholder="Enter text" />);
    expect(screen.getByPlaceholderText("Enter text")).toBeInTheDocument();
  });

  it("renders with value", () => {
    render(<MhInput value="Test value" />);
    expect(screen.getByDisplayValue("Test value")).toBeInTheDocument();
  });

  it("renders TextArea variant", () => {
    render(<MhInput.TextArea placeholder="Enter text" />);
    expect(screen.getByPlaceholderText("Enter text")).toBeInTheDocument();
  });

  it("renders Search variant", () => {
    render(<MhInput.Search placeholder="Search" />);
    expect(screen.getByPlaceholderText("Search")).toBeInTheDocument();
  });

  it("renders Password variant", () => {
    render(<MhInput.Password placeholder="Password" />);
    expect(screen.getByPlaceholderText("Password")).toBeInTheDocument();
  });

  it("trims leading and trailing spaces on blur", () => {
    render(<MhInput defaultValue="  hello world  " />);

    const input = screen.getByDisplayValue("  hello world  ");
    fireEvent.blur(input);

    expect(screen.getByDisplayValue("hello world")).toBeInTheDocument();
  });

  it("emits trimmed value through onChange when blur trims input", () => {
    let changedValue = "";

    render(
      <MhInput
        defaultValue="  hello world  "
        onChange={event => {
          changedValue = event.target.value;
        }}
      />
    );

    const input = screen.getByDisplayValue("  hello world  ");
    fireEvent.blur(input);

    expect(changedValue).toBe("hello world");
  });

  it("does not trim on blur when trimOnBlur is false", () => {
    render(<MhInput defaultValue="  hello world  " trimOnBlur={false} />);

    const input = screen.getByDisplayValue("  hello world  ");
    fireEvent.blur(input);

    expect(screen.getByDisplayValue("  hello world  ")).toBeInTheDocument();
  });
});
