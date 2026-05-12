import { useMutation, useQueryClient } from "@tanstack/react-query";
import { type AssistantUpdateIn, assistants } from "@/lib/api";
import { qk } from "@/lib/api/queryKeys.ts";

export function useUpdateAssistant(id: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: AssistantUpdateIn) => assistants.update(id, input),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: qk.assistants.root() });
      qc.invalidateQueries({ queryKey: qk.assistants.byId(id) });
    },
  });
}
