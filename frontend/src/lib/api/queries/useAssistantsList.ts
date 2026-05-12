import { useQuery } from "@tanstack/react-query";
import { type AssistantsQuery, assistants } from "@/lib/api";
import { qk } from "@/lib/api/queryKeys.ts";

export function useAssistantsList(query: AssistantsQuery) {
  return useQuery({
    queryKey: qk.assistants.list(query),
    queryFn: ({ signal }) => assistants.list(query, signal),
    placeholderData: (prev) => prev,
  });
}
