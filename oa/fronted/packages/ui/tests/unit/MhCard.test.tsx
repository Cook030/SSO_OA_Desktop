// import React from "react";
// import { render, screen } from "@testing-library/react";
// import { MhCard } from "../../src/components/MhCard";

// describe("MhCard", () => {
//   it("renders with title and children", () => {
//     render(
//       <MhCard title="Test Card">
//         <p>Card content</p>
//       </MhCard>
//     );
//     expect(screen.getByText("Test Card")).toBeInTheDocument();
//     expect(screen.getByText("Card content")).toBeInTheDocument();
//   });

//   it("applies shadow style when shadow prop is true", () => {
//     const { container } = render(
//       <MhCard shadow>
//         <p>Content</p>
//       </MhCard>
//     );
//     const card = container.querySelector(".ant-card");
//     expect(card).toHaveStyle({ boxShadow: "0 4px 12px rgba(0, 0, 0, 0.15)" });
//   });

//   it("applies hoverable prop", () => {
//     const { container } = render(
//       <MhCard hoverable>
//         <p>Content</p>
//       </MhCard>
//     );
//     const card = container.querySelector(".ant-card");
//     expect(card).toHaveClass("ant-card-hoverable");
//   });

//   it("applies custom title color", () => {
//     const { container } = render(
//       <MhCard title="Test" titleColor="#ff0000">
//         <p>Content</p>
//       </MhCard>
//     );
//     const cardHead = container.querySelector(".ant-card-head");
//     expect(cardHead).toHaveStyle({ color: "#ff0000" });
//   });

//   it("renders without title", () => {
//     render(
//       <MhCard>
//         <p>Content only</p>
//       </MhCard>
//     );
//     expect(screen.getByText("Content only")).toBeInTheDocument();
//   });
// });
