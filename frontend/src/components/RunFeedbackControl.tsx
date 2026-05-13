import { ThumbsDown, ThumbsUp, type LucideIcon } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { type AssistantRun, type RunFeedbackRating } from "@/lib/api";
import { showErrorToast } from "@/lib/api/errorMessage";
import { useSetRunFeedback } from "@/lib/api/queries/useSetRunFeedback";

type Props = {
  run: AssistantRun;
  currentUserId: string | null | undefined;
  onRunUpdated: (run: AssistantRun) => void;
};

const FEEDBACK_OPTIONS = [
  { rating: 1, Icon: ThumbsUp },
  { rating: -1, Icon: ThumbsDown },
] as const satisfies readonly {
  rating: RunFeedbackRating;
  Icon: LucideIcon;
}[];

export function RunFeedbackControl({ run, currentUserId, onRunUpdated }: Props) {
  const feedbackMutation = useSetRunFeedback();
  const canRate =
    currentUserId !== null &&
    currentUserId !== undefined &&
    run.userId === currentUserId &&
    run.status === "success";

  if (!canRate) return null;

  const handleFeedbackClick = (rating: RunFeedbackRating) => {
    feedbackMutation.mutate(
      { runId: run.id, rating },
      {
        onSuccess: onRunUpdated,
        onError: (error) => {
          showErrorToast(error, { fallback: "Не удалось сохранить оценку" });
        },
      },
    );
  };

  return (
    <div className="flex flex-wrap items-center gap-2 pt-1">
      <Label className="mr-1 text-[0.9375rem] font-bold">
        Оцените ответ ассистента:
      </Label>
      {FEEDBACK_OPTIONS.map(({ rating, Icon }) => {
        const selected = run.feedbackRating === rating;

        return (
          <Button
            key={rating}
            type="button"
            size="sm"
            variant={selected ? "black" : "soft"}
            aria-pressed={selected}
            disabled={feedbackMutation.isPending}
            onClick={() => handleFeedbackClick(rating)}
          >
            <Icon />
          </Button>
        );
      })}
    </div>
  );
}
