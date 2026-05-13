import { useState } from "react";
import { AssistantPicker } from "@/components/AssistantPicker";
import { PaginationControl } from "@/components/PaginationControl";
import { QueryStateBoundary } from "@/components/QueryStateBoundary";
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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { type AssistantRun, type RunStatus } from "@/lib/api";
import { useAdminRuns } from "@/lib/api/queries/useAdminRuns";
import { formatDateTime, runStatusLabel } from "@/lib/format";
import {
  enumParam,
  intParam,
  stringParam,
  useSearchParamsState,
} from "@/lib/hooks/useSearchParamsState";

const RUNS_PAGE_SIZE = 20;
const ALL_STATUSES = "__all__";

const RUN_STATUSES = ["pending", "success", "failed"] as const satisfies readonly RunStatus[];

function isRunStatus(value: string): value is RunStatus {
  return (RUN_STATUSES as readonly string[]).includes(value);
}

const filtersSchema = {
  status: enumParam(RUN_STATUSES),
  assistantId: stringParam(""),
  page: intParam(1),
};

export function AdminRunsPage() {
  const [filters, setFilters] = useSearchParamsState(filtersSchema);
  const [openRun, setOpenRun] = useState<AssistantRun | null>(null);

  const runsQuery = useAdminRuns({
    status: filters.status ?? undefined,
    assistantId: filters.assistantId === "" ? undefined : filters.assistantId,
    page: filters.page,
    pageSize: RUNS_PAGE_SIZE,
  });

  return (
    <section className="flex flex-col gap-8">
      <header>
        <p className="text-sm font-semibold text-primary">Администрирование</p>
        <h1 className="mt-1 text-4xl font-extrabold tracking-tight">Все запуски</h1>
        <p className="mt-2 text-sm text-muted-foreground">
          История запусков ассистентов по всем пользователям.
        </p>
      </header>

      <div className="flex flex-wrap items-end gap-3">
        <div className="flex flex-col gap-2">
          <Label className="text-xs font-semibold text-muted-foreground">Ассистент</Label>
          <AssistantPicker
            className="w-72"
            value={filters.assistantId === "" ? null : filters.assistantId}
            onChange={(v) => setFilters({ assistantId: v ?? "", page: 1 })}
          />
        </div>

        <div className="flex flex-col gap-2">
          <Label className="text-xs font-semibold text-muted-foreground">Статус</Label>
          <Select
            value={filters.status ?? ALL_STATUSES}
            onValueChange={(v) =>
              setFilters({
                status: v !== null && isRunStatus(v) ? v : null,
                page: 1,
              })
            }
          >
            <SelectTrigger className="w-44">
              <SelectValue
                placeholder="Статус"
                items={{
                  [ALL_STATUSES]: "Все",
                  ...runStatusLabel,
                }}
              />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value={ALL_STATUSES}>Все</SelectItem>
              <SelectItem value="pending">В процессе</SelectItem>
              <SelectItem value="success">Успех</SelectItem>
              <SelectItem value="failed">Ошибка</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </div>

      <QueryStateBoundary
        query={runsQuery}
        isEmpty={(data) => data.runs.length === 0}
        emptyFallback={
          <div className="rounded-2xl border border-dashed border-border p-12 text-center text-sm text-muted-foreground">
            Запусков по выбранным фильтрам нет.
          </div>
        }
      >
        {(data) => (
          <div className="flex flex-col gap-6">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Ассистент</TableHead>
                  <TableHead>Запрос</TableHead>
                  <TableHead>Статус</TableHead>
                  <TableHead className="w-[160px]">Дата</TableHead>
                  <TableHead className="w-[100px]" />
                </TableRow>
              </TableHeader>
              <TableBody>
                {data.runs.map((run) => (
                  <TableRow key={run.id}>
                    <TableCell className="font-semibold">
                      {run.assistantName ?? "—"}
                    </TableCell>
                    <TableCell className="max-w-[280px] truncate text-muted-foreground">
                      {run.userPrompt}
                    </TableCell>
                    <TableCell>
                      <RunStatusBadge status={run.status} />
                    </TableCell>
                    <TableCell>{formatDateTime(run.createdAt)}</TableCell>
                    <TableCell>
                      <Button size="sm" variant="soft" onClick={() => setOpenRun(run)}>
                        Открыть
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>

            <PaginationControl
              pagination={data.pagination}
              onChange={(page) => setFilters({ page })}
            />
          </div>
        )}
      </QueryStateBoundary>

      <RunDetailDialog run={openRun} onClose={() => setOpenRun(null)} />
    </section>
  );
}

function RunDetailDialog({ run, onClose }: { run: AssistantRun | null; onClose: () => void }) {
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
