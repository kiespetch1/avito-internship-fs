import {
  ApiError,
  apiRequest,
  apiStreamRequest,
  type ApiErrorCode,
  type Schemas,
} from "./client";

export type Role = Schemas["Role"];
export type User = Schemas["User"];
export type Token = Schemas["Token"];
export type AuthCredentialsIn = Schemas["AuthCredentialsIn"];
export type Category = Schemas["Category"];
export type CategoryCreateIn = Schemas["CategoryCreateIn"];
export type Assistant = Schemas["Assistant"];
export type AssistantCreateIn = Schemas["AssistantCreateIn"];
export type AssistantUpdateIn = Schemas["AssistantUpdateIn"];
export type AssistantRun = Schemas["AssistantRun"];
export type AssistantRunCreateIn = Schemas["AssistantRunCreateIn"];
export type AssistantRunStreamDelta = Schemas["AssistantRunStreamDelta"];
export type AssistantRunStreamFailure = Schemas["AssistantRunStreamFailure"];
export type RunFeedbackRating = Schemas["RunFeedbackRating"];
export type RunFeedbackUpsertIn = Schemas["RunFeedbackUpsertIn"];
export type RunStatus = Schemas["RunStatus"];
export type Pagination = Schemas["Pagination"];

export type AssistantsList = { assistants: Assistant[]; pagination: Pagination };
export type RunsList = { runs: AssistantRun[]; pagination: Pagination };

export type RunStreamCallbacks = {
  onRun?: (run: AssistantRun) => void;
  onDelta?: (delta: string) => void;
  onDone?: (run: AssistantRun) => void;
  onFailed?: (failure: AssistantRunStreamFailure) => void;
};

export const auth = {
  register: (input: AuthCredentialsIn) =>
    apiRequest<Token>("/auth/register", { method: "POST", body: input }),
  login: (input: AuthCredentialsIn) =>
    apiRequest<Token>("/auth/login", { method: "POST", body: input }),
  dummyLogin: (role: Role) =>
    apiRequest<Token>("/dummyLogin", { method: "POST", body: { role } }),
};

export const categories = {
  list: (signal?: AbortSignal) =>
    apiRequest<{ categories: Category[] }>("/categories", { signal }),
  create: (input: CategoryCreateIn) =>
    apiRequest<Category>("/categories", { method: "POST", body: input }),
};

export type AssistantsQuery = {
  categoryId?: string;
  q?: string;
  includeInactive?: boolean;
  favoriteOnly?: boolean;
  page?: number;
  pageSize?: number;
};

export const assistants = {
  list: (query: AssistantsQuery = {}, signal?: AbortSignal) =>
    apiRequest<AssistantsList>("/assistants", { query, signal }),
  get: (id: string, signal?: AbortSignal) =>
    apiRequest<Assistant>(`/assistants/${id}`, { signal }),
  create: (input: AssistantCreateIn) =>
    apiRequest<Assistant>("/assistants", { method: "POST", body: input }),
  update: (id: string, input: AssistantUpdateIn) =>
    apiRequest<Assistant>(`/assistants/${id}`, { method: "PUT", body: input }),
  addFavorite: (id: string) =>
    apiRequest<void>(`/assistants/${id}/favorite`, { method: "PUT" }),
  removeFavorite: (id: string) =>
    apiRequest<void>(`/assistants/${id}/favorite`, { method: "DELETE" }),
  run: (id: string, input: AssistantRunCreateIn) =>
    apiRequest<AssistantRun>(`/assistants/${id}/run`, { method: "POST", body: input }),
  runStream: async (
    id: string,
    input: AssistantRunCreateIn,
    callbacks: RunStreamCallbacks = {},
  ): Promise<AssistantRun> => {
    const state: { finalRun?: AssistantRun } = {};
    await apiStreamRequest(`/assistants/${id}/run/stream`, {
      method: "POST",
      body: input,
      onEvent: (event, data) => {
        if (event === "run") {
          if (!isAssistantRun(data)) throw malformedStreamEvent(event);
          callbacks.onRun?.(data);
          return;
        }
        if (event === "delta") {
          if (!isRunStreamDelta(data)) throw malformedStreamEvent(event);
          callbacks.onDelta?.(data.delta);
          return;
        }
        if (event === "done") {
          if (!isAssistantRun(data)) throw malformedStreamEvent(event);
          state.finalRun = data;
          callbacks.onDone?.(data);
          return;
        }
        if (event === "failed") {
          if (!isRunStreamFailure(data)) throw malformedStreamEvent(event);
          state.finalRun = data.run;
          callbacks.onFailed?.(data);
        }
      },
    });
    if (!state.finalRun) {
      throw new ApiError(0, "UNKNOWN", "Поток завершился без финального события");
    }

    return state.finalRun;
  },
};

const apiErrorCodes = [
  "INVALID_REQUEST",
  "UNAUTHORIZED",
  "FORBIDDEN",
  "NOT_FOUND",
  "EMAIL_ALREADY_EXISTS",
  "CATEGORY_NOT_FOUND",
  "ASSISTANT_NOT_FOUND",
  "ASSISTANT_INACTIVE",
  "LLM_PROVIDER_ERROR",
  "INTERNAL_ERROR",
] as const satisfies readonly ApiErrorCode[];

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}

function isApiErrorCode(value: unknown): value is ApiErrorCode {
  return typeof value === "string" && apiErrorCodes.some((code) => code === value);
}

function isRunStatus(value: unknown): value is RunStatus {
  return value === "pending" || value === "success" || value === "failed";
}

function isAssistantRun(value: unknown): value is AssistantRun {
  return (
    isRecord(value) &&
    typeof value.id === "string" &&
    typeof value.assistantId === "string" &&
    typeof value.userId === "string" &&
    typeof value.model === "string" &&
    typeof value.userPrompt === "string" &&
    isRunStatus(value.status)
  );
}

function isRunStreamDelta(value: unknown): value is AssistantRunStreamDelta {
  return isRecord(value) && typeof value.delta === "string";
}

function isRunStreamFailure(value: unknown): value is AssistantRunStreamFailure {
  return (
    isRecord(value) &&
    isAssistantRun(value.run) &&
    isRecord(value.error) &&
    isApiErrorCode(value.error.code) &&
    typeof value.error.message === "string"
  );
}

function malformedStreamEvent(event: string): ApiError {
  return new ApiError(0, "UNKNOWN", `Некорректное SSE-событие: ${event}`);
}

export type RunsQuery = {
  status?: RunStatus;
  page?: number;
  pageSize?: number;
};

export type AdminRunsQuery = RunsQuery & { assistantId?: string };

export const runs = {
  my: (query: RunsQuery = {}, signal?: AbortSignal) =>
    apiRequest<RunsList>("/runs/my", { query, signal }),
  admin: (query: AdminRunsQuery = {}, signal?: AbortSignal) =>
    apiRequest<RunsList>("/admin/runs", { query, signal }),
  setFeedback: (runId: string, input: RunFeedbackUpsertIn) =>
    apiRequest<AssistantRun>(`/runs/${runId}/feedback`, {
      method: "PUT",
      body: input,
    }),
};
