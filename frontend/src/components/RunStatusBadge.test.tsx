import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { RunStatusBadge } from "./RunStatusBadge";

describe("RunStatusBadge", () => {
  it("renders russian label for success status", () => {
    render(<RunStatusBadge status="success" />);
    expect(screen.getByText("Успех")).toBeInTheDocument();
  });

  it("renders russian label for pending status", () => {
    render(<RunStatusBadge status="pending" />);
    expect(screen.getByText("В процессе")).toBeInTheDocument();
  });

  it("renders russian label for failed status", () => {
    render(<RunStatusBadge status="failed" />);
    expect(screen.getByText("Ошибка")).toBeInTheDocument();
  });
});
