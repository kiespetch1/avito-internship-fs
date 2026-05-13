import { useMutation, useQueryClient } from "@tanstack/react-query";
import { type RunFeedbackRating, runs } from "@/lib/api";
import { qk } from "@/lib/api/queryKeys";

type SetRunFeedbackInput = {
  runId: string;
  rating: RunFeedbackRating;
};

export function useSetRunFeedback() {
  const qc = useQueryClient();

  return useMutation({
    mutationFn: ({ runId, rating }: SetRunFeedbackInput) =>
      runs.setFeedback(runId, { rating }),
    onSuccess: () => qc.invalidateQueries({ queryKey: qk.runs.root() }),
  });
}
