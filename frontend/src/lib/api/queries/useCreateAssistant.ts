import { useMutation, useQueryClient } from "@tanstack/react-query";
import { type AssistantCreateIn, assistants } from "@/lib/api";
import { qk } from "@/lib/api/queryKeys.ts";

export function useCreateAssistant() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: AssistantCreateIn) => assistants.create(input),
    onSuccess: () => qc.invalidateQueries({ queryKey: qk.assistants.root() }),
  });
}
