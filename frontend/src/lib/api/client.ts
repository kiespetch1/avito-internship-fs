import { loadAuth } from "@/lib/auth/storage";
import type { components } from "./schema";

export type Schemas = components["schemas"];
export type ErrorResponse = Schemas["ErrorResponse"];
export type ApiErrorCode = ErrorResponse["error"]["code"];

const BASE_URL = (import.meta.env.VITE_API_URL ?? "http://localhost:8080").replace(/\/$/, "");

export class ApiError extends Error {
  readonly status: number;
  readonly code: ApiErrorCode | "UNKNOWN";

  constructor(status: number, code: ApiErrorCode | "UNKNOWN", message: string) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.code = code;
  }
}

type QueryValue = string | number | boolean | undefined | null;

export type RequestOptions = {
  method?: "GET" | "POST" | "PUT" | "DELETE";
  query?: Record<string, QueryValue>;
  body?: unknown;
  signal?: AbortSignal;
};

export type StreamRequestOptions = {
  method?: "POST";
  body?: unknown;
  signal?: AbortSignal;
  onEvent: (event: string, data: unknown) => void;
};

function buildUrl(path: string, query?: Record<string, QueryValue>): string {
  const url = new URL(BASE_URL + path);
  if (query) {
    for (const [k, v] of Object.entries(query)) {
      if (v === undefined || v === null || v === "") continue;
      url.searchParams.set(k, String(v));
    }
  }
  return url.toString();
}

export async function apiRequest<TResponse>(
  path: string,
  options: RequestOptions = {},
): Promise<TResponse> {
  const headers: Record<string, string> = { Accept: "application/json" };
  const token = loadAuth()?.token;
  if (token) headers.Authorization = `Bearer ${token}`;

  let body: BodyInit | undefined;
  if (options.body !== undefined) {
    headers["Content-Type"] = "application/json";
    body = JSON.stringify(options.body);
  }

  const res = await fetch(buildUrl(path, options.query), {
    method: options.method ?? "GET",
    headers,
    body,
    signal: options.signal,
  });

  if (res.status === 204) return undefined as TResponse;

  const text = await res.text();
  const data: unknown = text ? safeJson(text) : null;

  if (!res.ok) {
    const parsed = parseError(data);
    throw new ApiError(res.status, parsed.code, parsed.message);
  }

  return data as TResponse;
}

export async function apiStreamRequest(
  path: string,
  options: StreamRequestOptions,
): Promise<void> {
  const headers: Record<string, string> = { Accept: "text/event-stream" };
  const token = loadAuth()?.token;
  if (token) headers.Authorization = `Bearer ${token}`;

  let body: BodyInit | undefined;
  if (options.body !== undefined) {
    headers["Content-Type"] = "application/json";
    body = JSON.stringify(options.body);
  }

  const res = await fetch(buildUrl(path), {
    method: options.method ?? "POST",
    headers,
    body,
    signal: options.signal,
  });

  if (!res.ok) {
    const text = await res.text();
    const parsed = parseError(text ? safeJson(text) : null);
    throw new ApiError(res.status, parsed.code, parsed.message);
  }
  if (!res.body) {
    throw new ApiError(0, "UNKNOWN", "Поток ответа недоступен");
  }

  const reader = res.body.getReader();
  const decoder = new TextDecoder();
  let buffer = "";

  const dispatchFrame = (frame: string) => {
    const parsed = parseSSEFrame(frame);
    if (!parsed) return;
    const data = parseJson(parsed.data);
    options.onEvent(parsed.event, data.ok ? data.value : parsed.data);
  };

  while (true) {
    const { value, done } = await reader.read();
    if (done) break;
    buffer += decoder.decode(value, { stream: true });

    let boundary = findSSEBoundary(buffer);
    while (boundary) {
      const frame = buffer.slice(0, boundary.index);
      buffer = buffer.slice(boundary.index + boundary.length);
      dispatchFrame(frame);
      boundary = findSSEBoundary(buffer);
    }
  }

  buffer += decoder.decode();
  if (buffer.trim() !== "") {
    dispatchFrame(buffer);
  }
}

function safeJson(text: string): unknown {
  try {
    return JSON.parse(text);
  } catch {
    return null;
  }
}

function parseError(data: unknown): { code: ApiErrorCode | "UNKNOWN"; message: string } {
  if (
    data &&
    typeof data === "object" &&
    "error" in data &&
    data.error &&
    typeof data.error === "object" &&
    "code" in data.error &&
    "message" in data.error
  ) {
    const err = data.error as { code: ApiErrorCode; message: string };
    return { code: err.code, message: err.message };
  }
  return { code: "UNKNOWN", message: "Неизвестная ошибка" };
}

function parseJson(text: string): { ok: true; value: unknown } | { ok: false } {
  try {
    return { ok: true, value: JSON.parse(text) };
  } catch {
    return { ok: false };
  }
}

function parseSSEFrame(frame: string): { event: string; data: string } | null {
  let event = "message";
  const dataLines: string[] = [];

  for (const line of frame.split(/\r?\n/)) {
    if (line === "" || line.startsWith(":")) continue;

    const colonIndex = line.indexOf(":");
    const field = colonIndex >= 0 ? line.slice(0, colonIndex) : line;
    let value = colonIndex >= 0 ? line.slice(colonIndex + 1) : "";
    if (value.startsWith(" ")) {
      value = value.slice(1);
    }

    if (field === "event") {
      event = value;
    }
    if (field === "data") {
      dataLines.push(value);
    }
  }

  if (dataLines.length === 0) return null;

  return { event, data: dataLines.join("\n") };
}

function findSSEBoundary(buffer: string): { index: number; length: number } | null {
  const lfIndex = buffer.indexOf("\n\n");
  const crlfIndex = buffer.indexOf("\r\n\r\n");
  if (lfIndex === -1 && crlfIndex === -1) return null;
  if (lfIndex === -1) return { index: crlfIndex, length: 4 };
  if (crlfIndex === -1) return { index: lfIndex, length: 2 };

  return lfIndex < crlfIndex
    ? { index: lfIndex, length: 2 }
    : { index: crlfIndex, length: 4 };
}
