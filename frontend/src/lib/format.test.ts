import { describe, expect, it } from "vitest";
import { formatDateTime, runStatusLabel } from "./format";

describe("formatDateTime", () => {
  it("returns em-dash for empty values", () => {
    expect(formatDateTime(null)).toBe("—");
    expect(formatDateTime(undefined)).toBe("—");
    expect(formatDateTime("")).toBe("—");
  });

  it("returns em-dash for invalid date strings", () => {
    expect(formatDateTime("not a date")).toBe("—");
  });

  it("formats a valid ISO string using ru-RU locale", () => {
    const formatted = formatDateTime("2026-05-12T09:30:00Z");
    expect(formatted).toMatch(/\d{2}\.\d{2}\.\d{4}/);
    expect(formatted).toMatch(/\d{2}:\d{2}/);
  });
});

describe("runStatusLabel", () => {
  it("covers all three run statuses with russian labels", () => {
    expect(runStatusLabel.pending).toBe("В процессе");
    expect(runStatusLabel.success).toBe("Успех");
    expect(runStatusLabel.failed).toBe("Ошибка");
  });
});
