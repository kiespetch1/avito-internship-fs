import { useQuery } from "@tanstack/react-query";
import { assistants } from "@/lib/api";
import { qk } from "@/lib/api/queryKeys.ts";

export function useAssistant(id: string) {
  return useQuery({
    queryKey: qk.assistants.byId(id),
    queryFn: ({ signal }) => assistants.get(id, signal),
    enabled: id !== "",
  });
}
