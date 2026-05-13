import { useQuery } from "@tanstack/react-query";
import { useState } from "react";
import { PaginationControl } from "@/components/PaginationControl";
import { QueryStateBoundary } from "@/components/QueryStateBoundary";
import { RunDetailDialog } from "@/components/RunDetailDialog";
import { RunStatusBadge } from "@/components/RunStatusBadge";
import { Button } from "@/components/ui/button";
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
import { useAuth } from "@/lib/auth";
import { formatDateTime, runStatusLabel } from "@/lib/format";
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
  const { user } = useAuth();

  const listQuery = {
    status: filters.status ?? undefined,
    page: filters.page,
    pageSize: RUNS_PAGE_SIZE,
  };

  const runsQuery = useQuery({
    queryKey: qk.runs.my(user?.id, listQuery),
    queryFn: ({ signal }) => runsApi.my(listQuery, signal),
    placeholderData: (prev) => prev,
  });

  return (
    <section className="flex flex-col gap-8">
      <header>
        <p className="text-sm font-semibold text-primary">История</p>
        <h1 className="mt-1 text-4xl font-extrabold tracking-tight">
          Мои запуски
        </h1>
        <p className="mt-2 text-sm text-muted-foreground">
          История ваших обращений к ассистентам.
        </p>
      </header>

      <div className="flex flex-wrap items-center gap-3">
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

      <QueryStateBoundary
        query={runsQuery}
        isEmpty={(data) => data.runs.length === 0}
        emptyFallback={
          <div className="rounded-2xl border border-dashed border-border p-12 text-center text-sm text-muted-foreground">
            У вас ещё нет запусков.
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

      <RunDetailDialog
        run={openRun}
        currentUserId={user?.id}
        onRunUpdated={setOpenRun}
        onClose={() => setOpenRun(null)}
      />
    </section>
  );
}
