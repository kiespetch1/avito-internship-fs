import { useQuery } from "@tanstack/react-query";
import { type AdminRunsQuery, runs } from "@/lib/api";
import { qk } from "@/lib/api/queryKeys.ts";

export function useAdminRuns(query: AdminRunsQuery) {
  return useQuery({
    queryKey: qk.runs.admin(query),
    queryFn: ({ signal }) => runs.admin(query, signal),
    placeholderData: (prev) => prev,
  });
}
