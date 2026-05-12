import type { Role, User } from "@/lib/api";

const KEY = "avito.auth.v1";

export type AuthSnapshot = { token: string; user: User };

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}

function isRole(value: unknown): value is Role {
  return value === "admin" || value === "user";
}

function isUser(value: unknown): value is User {
  if (!isRecord(value)) return false;
  if (typeof value.id !== "string") return false;
  if (typeof value.email !== "string") return false;
  if (!isRole(value.role)) return false;
  if (
    value.createdAt !== undefined &&
    value.createdAt !== null &&
    typeof value.createdAt !== "string"
  ) {
    return false;
  }
  return true;
}

function isAuthSnapshot(value: unknown): value is AuthSnapshot {
  if (!isRecord(value)) return false;
  if (typeof value.token !== "string") return false;
  return isUser(value.user);
}

let cached: AuthSnapshot | null | undefined;

function readFromStorage(): AuthSnapshot | null {
  try {
    const raw = localStorage.getItem(KEY);
    if (!raw) return null;
    const parsed: unknown = JSON.parse(raw);
    return isAuthSnapshot(parsed) ? parsed : null;
  } catch {
    return null;
  }
}

export function loadAuth(): AuthSnapshot | null {
  if (cached === undefined) cached = readFromStorage();
  return cached;
}

const listeners = new Set<() => void>();

function emit(): void {
  for (const fn of listeners) fn();
}

export function subscribeAuth(listener: () => void): () => void {
  listeners.add(listener);
  return () => {
    listeners.delete(listener);
  };
}

export function saveAuth(snap: AuthSnapshot): void {
  localStorage.setItem(KEY, JSON.stringify(snap));
  cached = snap;
  emit();
}

export function clearAuth(): void {
  localStorage.removeItem(KEY);
  cached = null;
  emit();
}

if (typeof window !== "undefined") {
  window.addEventListener("storage", (e) => {
    if (e.key !== KEY && e.key !== null) return;
    cached = readFromStorage();
    emit();
  });
}
