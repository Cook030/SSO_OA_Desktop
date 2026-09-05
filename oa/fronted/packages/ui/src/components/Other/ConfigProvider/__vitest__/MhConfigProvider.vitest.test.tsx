import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { MhConfigProvider } from "./../index";

describe("MhConfigProvider (vitest)", () => {
  it("renders children", () => {
    render(
      <MhConfigProvider>
        <div>App Content</div>
      </MhConfigProvider>
    );
    expect(screen.getByText("App Content")).toBeInTheDocument();
  });

  it("renders with theme config", () => {
    render(
      <MhConfigProvider theme={{ token: { colorPrimary: "#00b96b" } }}>
        <div>Themed Content</div>
      </MhConfigProvider>
    );
    expect(screen.getByText("Themed Content")).toBeInTheDocument();
  });
});
