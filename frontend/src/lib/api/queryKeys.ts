import type { AdminRunsQuery, AssistantsQuery, RunsQuery } from "./endpoints";

export const qk = {
  categories: {
    all: () => ["categories"] as const,
  },
  assistants: {
    root: () => ["assistants"] as const,
    list: (q: AssistantsQuery) => ["assistants", "list", q] as const,
    byId: (id: string) => ["assistants", "byId", id] as const,
  },
  runs: {
    root: () => ["runs"] as const,
    my: (q: RunsQuery) => ["runs", "my", q] as const,
    admin: (q: AdminRunsQuery) => ["runs", "admin", q] as const,
  },
} as const;
