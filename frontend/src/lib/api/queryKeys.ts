import type { AdminRunsQuery, AssistantsQuery, RunsQuery } from "./endpoints";

type ViewerScope = string | null | undefined;

function viewerKey(viewerId: ViewerScope): string {
  return viewerId ?? "anonymous";
}

export const qk = {
  categories: {
    all: () => ["categories"] as const,
  },
  assistants: {
    root: () => ["assistants"] as const,
    list: (viewerId: ViewerScope, q: AssistantsQuery) =>
      ["assistants", viewerKey(viewerId), "list", q] as const,
    byId: (viewerId: ViewerScope, id: string) =>
      ["assistants", viewerKey(viewerId), "byId", id] as const,
  },
  runs: {
    root: () => ["runs"] as const,
    my: (viewerId: ViewerScope, q: RunsQuery) =>
      ["runs", viewerKey(viewerId), "my", q] as const,
    admin: (q: AdminRunsQuery) => ["runs", "admin", q] as const,
  },
} as const;
