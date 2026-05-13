import { z } from "zod";

export const assistantSchema = z.object({
  name: z.string().trim().min(1, "Введите название"),
  description: z.string().trim().min(1, "Введите описание"),
  categoryId: z.string().uuid("Выберите категорию"),
  model: z.string().trim().min(1, "Укажите модель"),
  tags: z.string().trim(),
  systemPrompt: z.string().trim().min(1, "Системный промпт обязателен"),
  exampleUserPrompt: z.string().trim(),
  isActive: z.boolean(),
});

export type AssistantFormValues = z.input<typeof assistantSchema>;
