import { useQuery } from "@tanstack/react-query";
import { useState } from "react";
import { PaginationControl } from "@/components/PaginationControl";
import { QueryStateBoundary } from "@/components/QueryStateBoundary";
import { RunStatusBadge } from "@/components/RunStatusBadge";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
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
import { type AssistantRun, type RunStatus, runs as runsApi } from "@/lib/api";
import { qk } from "@/lib/api/queryKeys";
import { formatDateTime } from "@/lib/format";
import { enumParam, intParam, useSearchParamsState } from "@/lib/hooks/useSearchParamsState";

const RUNS_PAGE_SIZE = 20;
const ALL_STATUSES = "__all__";

const RUN_STATUSES = ["pending", "success", "failed"] as const satisfies readonly RunStatus[];

function isRunStatus(value: string): value is RunStatus {
  return (RUN_STATUSES as readonly string[]).includes(value);
}

const filtersSchema = {
  status: enumParam(RUN_STATUSES),
  page: intParam(1),
};

export function MyRunsPage() {
  const [filters, setFilters] = useSearchParamsState(filtersSchema);
  const [openRun, setOpenRun] = useState<AssistantRun | null>(null);

  const listQuery = {
    status: filters.status ?? undefined,
    page: filters.page,
    pageSize: RUNS_PAGE_SIZE,
  };

  const runsQuery = useQuery({
    queryKey: qk.runs.my(listQuery),
    queryFn: ({ signal }) => runsApi.my(listQuery, signal),
    placeholderData: (prev) => prev,
  });

  return (
    <section className="flex flex-col gap-6">
      <header className="flex flex-col gap-2">
        <h1 className="text-2xl font-semibold tracking-tight">Мои запуски</h1>
        <p className="text-sm text-muted-foreground">История ваших обращений к ассистентам.</p>
      </header>

      <div className="flex flex-wrap items-end gap-3 rounded-md border border-border p-4">
        <div className="flex w-[220px] flex-col gap-1.5">
          <Label>Статус</Label>
          <Select
            value={filters.status ?? ALL_STATUSES}
            onValueChange={(v) =>
              setFilters({
                status: v !== null && isRunStatus(v) ? v : null,
                page: 1,
              })
            }
          >
            <SelectTrigger>
              <SelectValue placeholder="Все" />
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
          <div className="rounded-md border border-dashed border-border p-8 text-center text-sm text-muted-foreground">
            У вас ещё нет запусков.
          </div>
        }
      >
        {(data) => (
          <div className="flex flex-col gap-4">
            <div className="rounded-md border border-border">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Ассистент</TableHead>
                    <TableHead>Статус</TableHead>
                    <TableHead>Запрос</TableHead>
                    <TableHead>Ответ</TableHead>
                    <TableHead className="w-[140px]">Дата</TableHead>
                    <TableHead className="w-[100px]" />
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {data.runs.map((run) => (
                    <TableRow key={run.id}>
                      <TableCell>{run.assistantName ?? "—"}</TableCell>
                      <TableCell>
                        <RunStatusBadge status={run.status} />
                      </TableCell>
                      <TableCell className="max-w-[240px] truncate">{run.userPrompt}</TableCell>
                      <TableCell className="max-w-[240px] truncate text-muted-foreground">
                        {run.status === "failed" ? (run.error ?? "—") : (run.output ?? "—")}
                      </TableCell>
                      <TableCell>{formatDateTime(run.createdAt)}</TableCell>
                      <TableCell>
                        <Button size="sm" variant="outline" onClick={() => setOpenRun(run)}>
                          Открыть
                        </Button>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>

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
      <DialogContent className="max-w-2xl">
        {run && (
          <>
            <DialogHeader>
              <DialogTitle className="flex items-center gap-2">
                {run.assistantName ?? "Запуск"}
                <RunStatusBadge status={run.status} />
              </DialogTitle>
              <DialogDescription>{formatDateTime(run.createdAt)}</DialogDescription>
            </DialogHeader>

            <div className="flex flex-col gap-3 text-sm">
              <div>
                <Label className="mb-1 block">Запрос пользователя</Label>
                <pre className="whitespace-pre-wrap rounded-md bg-muted p-3">{run.userPrompt}</pre>
              </div>
              <div>
                <Label className="mb-1 block">
                  {run.status === "failed" ? "Ошибка" : "Ответ"}
                </Label>
                <pre className="whitespace-pre-wrap rounded-md bg-muted p-3">
                  {run.status === "failed" ? (run.error ?? "—") : (run.output ?? "—")}
                </pre>
              </div>
            </div>
          </>
        )}
      </DialogContent>
    </Dialog>
  );
}
