import { RunFeedbackControl } from "@/components/RunFeedbackControl";
import { RunStatusBadge } from "@/components/RunStatusBadge";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";
import { type AssistantRun } from "@/lib/api";
import { formatDateTime } from "@/lib/format";

type Props = {
  run: AssistantRun | null;
  currentUserId: string | null | undefined;
  onRunUpdated: (run: AssistantRun) => void;
  onClose: () => void;
};

export function RunDetailDialog({ run, currentUserId, onRunUpdated, onClose }: Props) {
  return (
    <Dialog open={run !== null} onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="flex max-h-[calc(100dvh-2rem)] grid-rows-none flex-col overflow-hidden sm:max-w-2xl">
        {run && (
          <>
            <DialogHeader className="shrink-0 pr-8">
              <DialogTitle className="flex items-center gap-2">
                {run.assistantName ?? "Запуск"}
                <RunStatusBadge status={run.status} />
              </DialogTitle>
              <DialogDescription>{formatDateTime(run.createdAt)}</DialogDescription>
            </DialogHeader>

            <div className="min-h-0 space-y-4 overflow-y-auto pr-1">
              <div className="space-y-2">
                <Label className="text-[0.9375rem] font-bold">Запрос пользователя</Label>
                <div className="rounded-2xl bg-secondary p-4 text-[0.9375rem] leading-relaxed whitespace-pre-wrap">
                  {run.userPrompt}
                </div>
              </div>
              <div className="space-y-2">
                <Label className="text-[0.9375rem] font-bold">
                  {run.status === "failed" ? "Ошибка" : "Ответ"}
                </Label>
                <div className="rounded-2xl bg-secondary p-4 text-[0.9375rem] leading-relaxed whitespace-pre-wrap">
                  {run.status === "failed" ? (run.error ?? "—") : (run.output ?? "—")}
                </div>
                <RunFeedbackControl
                  run={run}
                  currentUserId={currentUserId}
                  onRunUpdated={onRunUpdated}
                />
              </div>
            </div>

            <DialogFooter className="shrink-0">
              <Button variant="black" size="lg" onClick={onClose}>
                Закрыть
              </Button>
            </DialogFooter>
          </>
        )}
      </DialogContent>
    </Dialog>
  );
}
