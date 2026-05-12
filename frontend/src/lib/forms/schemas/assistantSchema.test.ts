import { describe, expect, it } from "vitest";
import { assistantSchema } from "./assistantSchema";
import { categorySchema } from "./categorySchema";

const validAssistant = {
  name: "Повар",
  description: "Помогает с рецептами",
  categoryId: "00000000-0000-4000-8000-000000000000",
  model: "gpt-4o-mini",
  systemPrompt: "Ты повар",
  exampleUserPrompt: "курица, рис",
  isActive: true,
};

describe("assistantSchema", () => {
  it("accepts a fully valid assistant payload", () => {
    const result = assistantSchema.safeParse(validAssistant);
    expect(result.success).toBe(true);
  });

  it("rejects an empty system prompt", () => {
    const result = assistantSchema.safeParse({ ...validAssistant, systemPrompt: "   " });
    expect(result.success).toBe(false);
    if (!result.success) {
      const issue = result.error.issues.find((i) => i.path[0] === "systemPrompt");
      expect(issue?.message).toBe("Системный промпт обязателен");
    }
  });

  it("rejects a non-uuid categoryId", () => {
    const result = assistantSchema.safeParse({ ...validAssistant, categoryId: "not-a-uuid" });
    expect(result.success).toBe(false);
    if (!result.success) {
      const issue = result.error.issues.find((i) => i.path[0] === "categoryId");
      expect(issue?.message).toBe("Выберите категорию");
    }
  });

  it("allows an empty example user prompt", () => {
    const result = assistantSchema.safeParse({ ...validAssistant, exampleUserPrompt: "" });
    expect(result.success).toBe(true);
  });
});

describe("categorySchema", () => {
  it("requires a non-empty name", () => {
    const result = categorySchema.safeParse({ name: "  ", description: "" });
    expect(result.success).toBe(false);
  });

  it("accepts a valid category with empty description", () => {
    const result = categorySchema.safeParse({ name: "Еда", description: "" });
    expect(result.success).toBe(true);
  });
});
