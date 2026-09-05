// import React from "react";
// import { render, screen, fireEvent } from "@testing-library/react";
// import { MhInput } from "../../src/components/MhInput";

// describe("MhInput", () => {
//   it("renders correctly", () => {
//     render(<MhInput placeholder="Enter text" />);
//     expect(screen.getByPlaceholderText("Enter text")).toBeInTheDocument();
//   });

//   it("handles value change", () => {
//     const handleChange = jest.fn();
//     render(<MhInput onValueChange={handleChange} />);
//     const input = screen.getByRole("textbox");

//     fireEvent.change(input, { target: { value: "test" } });
//     expect(handleChange).toHaveBeenCalledWith("test");
//   });

//   it("trims value when autoTrim is true", () => {
//     const handleChange = jest.fn();
//     render(<MhInput onValueChange={handleChange} autoTrim />);
//     const input = screen.getByRole("textbox");

//     fireEvent.change(input, { target: { value: "  test  " } });
//     expect(handleChange).toHaveBeenCalledWith("test");
//   });

//   it("shows character count when showCount is true", () => {
//     render(<MhInput showCount maxLength={50} />);
//     expect(screen.getByRole("textbox")).toBeInTheDocument();
//   });

//   it("respects maxLength prop", () => {
//     render(<MhInput maxLength={10} />);
//     const input = screen.getByRole("textbox") as HTMLInputElement;
//     expect(input.maxLength).toBe(10);
//   });
// });
