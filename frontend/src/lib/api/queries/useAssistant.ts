import { useQuery } from "@tanstack/react-query";
import { assistants } from "@/lib/api";
import { qk } from "@/lib/api/queryKeys.ts";
import { useAuth } from "@/lib/auth";

export function useAssistant(id: string) {
  const { user } = useAuth();

  return useQuery({
    queryKey: qk.assistants.byId(user?.id, id),
    queryFn: ({ signal }) => assistants.get(id, signal),
    enabled: id !== "",
  });
}
