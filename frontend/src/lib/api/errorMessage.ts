import { ApiError } from "@/lib/api";
import { toast } from "sonner";

const DEFAULT_PROD_ERROR_MESSAGE = "Что-то пошло не так. Попробуйте ещё раз.";

export function getErrorMessage(error: unknown): string {
  if (import.meta.env.DEV) {
    if (error instanceof Error) return error.message;
    return String(error);
  }

  if (error instanceof ApiError) {
    switch (error.status) {
      case 400:
        return error.message;
      case 401:
        return "Нужно войти в аккаунт.";
      case 403:
        return "Недостаточно прав.";
      case 404:
        return "Данные не найдены.";
      default:
        return DEFAULT_PROD_ERROR_MESSAGE;
    }
  }

  return DEFAULT_PROD_ERROR_MESSAGE;
}

export function getRunErrorMessage(error?: string | null): string {
  if (import.meta.env.DEV) {
    return error || "Запуск завершился ошибкой";
  }

  return "Запуск завершился ошибкой. Попробуйте ещё раз.";
}

export function showErrorToast(error: unknown): void {
  console.error(error);

  toast.error(getErrorMessage(error));
}
