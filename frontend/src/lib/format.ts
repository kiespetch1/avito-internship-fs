import type { RunStatus } from "@/lib/api";

const dateTimeFormatter = new Intl.DateTimeFormat("ru-RU", {
  day: "2-digit",
  month: "2-digit",
  year: "numeric",
  hour: "2-digit",
  minute: "2-digit",
});

export function formatDateTime(value: string | null | undefined): string {
  if (!value) return "—";
  const d = new Date(value);
  if (Number.isNaN(d.getTime())) return "—";
  return dateTimeFormatter.format(d);
}

export const runStatusLabel: Record<RunStatus, string> = {
  pending: "В процессе",
  success: "Успех",
  failed: "Ошибка",
};

export const runStatusTone: Record<RunStatus, "default" | "success" | "destructive"> = {
  pending: "default",
  success: "success",
  failed: "destructive",
};
