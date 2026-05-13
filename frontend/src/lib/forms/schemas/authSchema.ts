import { z } from "zod";

const hasLetter = /\p{L}/u;
const hasNumberOrSpecial = /[\p{N}\p{P}\p{S}]/u;

export const authCredentialsSchema = z.object({
  email: z
    .string()
    .trim()
    .email("Введите корректную почту")
    .max(254, "Почта слишком длинная"),
  password: z
    .string()
    .min(8, "Пароль должен быть не короче 8 символов")
    .max(72, "Пароль должен быть не длиннее 72 символов")
    .refine((value) => hasLetter.test(value), "Пароль должен содержать букву")
    .refine(
      (value) => hasNumberOrSpecial.test(value),
      "Пароль должен содержать цифру или спецсимвол",
    ),
});

export type AuthCredentialsFormValues = z.input<typeof authCredentialsSchema>;
