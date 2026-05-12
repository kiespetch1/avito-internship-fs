import { useMutation, useQueryClient } from "@tanstack/react-query";
import { type AssistantRunCreateIn, assistants } from "@/lib/api";
import { qk } from "@/lib/api/queryKeys.ts";

export function useRunAssistant(assistantId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: AssistantRunCreateIn) => assistants.run(assistantId, input),
    onSuccess: () => qc.invalidateQueries({ queryKey: qk.runs.root() }),
  });
}
