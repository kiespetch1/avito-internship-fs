import { useQuery } from "@tanstack/react-query";
import { type AssistantsQuery, assistants } from "@/lib/api";
import { qk } from "@/lib/api/queryKeys.ts";
import { useAuth } from "@/lib/auth";

export function useAssistantsList(query: AssistantsQuery) {
  const { user } = useAuth();

  return useQuery({
    queryKey: qk.assistants.list(user?.id, query),
    queryFn: ({ signal }) => assistants.list(query, signal),
    placeholderData: (prev) => prev,
  });
}
