import { useMutation, useQueryClient } from "@tanstack/react-query";
import {
  type AssistantRun,
  type AssistantRunCreateIn,
  type AssistantRunStreamFailure,
  assistants,
} from "@/lib/api";
import { qk } from "@/lib/api/queryKeys.ts";

type RunAssistantProgress = {
  onRun?: (run: AssistantRun) => void;
  onDelta?: (delta: string) => void;
  onDone?: (run: AssistantRun) => void;
  onFailed?: (failure: AssistantRunStreamFailure) => void;
};

export function useRunAssistant(
  assistantId: string,
  progress: RunAssistantProgress = {},
) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: AssistantRunCreateIn) =>
      assistants.runStream(assistantId, input, progress),
    onSettled: () => qc.invalidateQueries({ queryKey: qk.runs.root() }),
  });
}
