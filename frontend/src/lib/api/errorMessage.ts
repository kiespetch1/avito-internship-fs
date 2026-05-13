import { ApiError } from "@/lib/api";
import { toast } from "sonner";

const DEFAULT_PROD_ERROR_MESSAGE = "Что-то пошло не так. Попробуйте ещё раз.";
const PROD_RUN_ERROR_MESSAGE = "Запуск завершился ошибкой. Попробуйте ещё раз.";
const PROD_LLM_ERROR_MESSAGE =
  "Не удалось получить ответ ассистента. Попробуйте ещё раз.";

type ErrorMessageOptions = {
  fallback?: string;
  unauthorizedMessage?: string;
};

type RunErrorMessageOptions = {
  showTechnical?: boolean;
};

export function getErrorMessage(
  error: unknown,
  options: ErrorMessageOptions = {},
): string {
  return getErrorMessageForEnv(error, import.meta.env.DEV, options);
}

export function getErrorMessageForEnv(
  error: unknown,
  isDev: boolean,
  {
    fallback = DEFAULT_PROD_ERROR_MESSAGE,
    unauthorizedMessage = "Нужно войти в аккаунт.",
  }: ErrorMessageOptions = {},
): string {
  if (isDev) {
    if (error instanceof Error) return error.message;
    return String(error);
  }

  if (error instanceof ApiError) {
    if (error.code === "LLM_PROVIDER_ERROR") {
      return PROD_LLM_ERROR_MESSAGE;
    }
    if (error.code === "EMAIL_ALREADY_EXISTS") {
      return "Почта уже зарегистрирована.";
    }
    switch (error.status) {
      case 400:
        return error.message;
      case 401:
        return unauthorizedMessage;
      case 403:
        return "Недостаточно прав.";
      case 404:
        return "Данные не найдены.";
      default:
        return fallback;
    }
  }

  return fallback;
}

export function getRunErrorMessage(
  error?: string | null,
  options: RunErrorMessageOptions = {},
): string {
  return getRunErrorMessageForEnv(error, import.meta.env.DEV, options);
}

export function getRunErrorMessageForEnv(
  error: string | null | undefined,
  isDev: boolean,
  { showTechnical = false }: RunErrorMessageOptions = {},
): string {
  if (isDev || showTechnical) {
    return error || "Запуск завершился ошибкой";
  }

  return PROD_RUN_ERROR_MESSAGE;
}

export function showErrorToast(
  error: unknown,
  options: ErrorMessageOptions = {},
): void {
  console.error(error);

  toast.error(getErrorMessage(error, options));
}
