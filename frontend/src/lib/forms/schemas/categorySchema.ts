import { z } from "zod";

export const categorySchema = z.object({
  name: z.string().trim().min(1, "Введите название"),
  description: z.string().trim(),
});

export type CategoryFormValues = z.input<typeof categorySchema>;
