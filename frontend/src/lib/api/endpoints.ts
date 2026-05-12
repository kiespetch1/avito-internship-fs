import { apiRequest, type Schemas } from "./client";

export type Role = Schemas["Role"];
export type User = Schemas["User"];
export type Token = Schemas["Token"];
export type Category = Schemas["Category"];
export type CategoryCreateIn = Schemas["CategoryCreateIn"];
export type Assistant = Schemas["Assistant"];
export type AssistantCreateIn = Schemas["AssistantCreateIn"];
export type AssistantUpdateIn = Schemas["AssistantUpdateIn"];
export type AssistantRun = Schemas["AssistantRun"];
export type AssistantRunCreateIn = Schemas["AssistantRunCreateIn"];
export type RunStatus = Schemas["RunStatus"];
export type Pagination = Schemas["Pagination"];

export type AssistantsList = { assistants: Assistant[]; pagination: Pagination };
export type RunsList = { runs: AssistantRun[]; pagination: Pagination };

export const auth = {
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
  run: (id: string, input: AssistantRunCreateIn) =>
    apiRequest<AssistantRun>(`/assistants/${id}/run`, { method: "POST", body: input }),
};

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
};
