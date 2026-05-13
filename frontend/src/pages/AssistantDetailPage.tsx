import { useQuery } from "@tanstack/react-query";
import { useEffect, useRef, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { AlertCircle, Loader2, Pencil } from "lucide-react";
import { toast } from "sonner";
import { AssistantFavoriteButton } from "@/components/AssistantFavoriteButton";
import { QueryStateBoundary } from "@/components/QueryStateBoundary";
import { RunFeedbackControl } from "@/components/RunFeedbackControl";
import { RunStatusBadge } from "@/components/RunStatusBadge";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Tag } from "@/components/ui/tag";
import { Textarea } from "@/components/ui/textarea";
import { type AssistantRun, assistants as assistantsApi } from "@/lib/api";
import { getRunErrorMessage, showErrorToast } from "@/lib/api/errorMessage";
import { useRunAssistant } from "@/lib/api/queries/useRunAssistant";
import { qk } from "@/lib/api/queryKeys";
import { useAuth } from "@/lib/auth";

export function AssistantDetailPage() {
  const { id = "" } = useParams<{ id: string }>();
  const { user } = useAuth();
  const isAdmin = user?.role === "admin";

  const [prompt, setPrompt] = useState("");
  const [streaming, setStreaming] = useState(true);
  const [lastRun, setLastRun] = useState<AssistantRun | null>(null);
  const resultRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    if (lastRun) {
      resultRef.current?.scrollIntoView({ behavior: "smooth", block: "start" });
    }
  }, [lastRun?.id]);

  const assistantQuery = useQuery({
    queryKey: qk.assistants.byId(user?.id, id),
    queryFn: ({ signal }) => assistantsApi.get(id, signal),
    enabled: id !== "",
  });

  const runMutation = useRunAssistant(id, {
    onRun: (run) => {
      setLastRun({ ...run, output: "" });
    },
    onDelta: (delta) => {
      setLastRun((current) => {
        if (!current) return current;

        return {
          ...current,
          output: `${current.output ?? ""}${delta}`,
          status: "pending",
        };
      });
    },
    onDone: (run) => {
      setLastRun(run);
    },
    onFailed: (failure) => {
      setLastRun(failure.run);
    },
  });

  const handleRun = () => {
    const trimmed = prompt.trim();
    if (trimmed === "") {
      toast.error("Введите запрос");
      return;
    }
    setLastRun(null);
    runMutation.mutate(
      { input: { userPrompt: trimmed }, streaming },
      {
        onSuccess: (run) => {
          setLastRun(run);
          if (run.status === "failed") {
            toast.error(getRunErrorMessage(run.error));
          }
        },
        onError: (err) => {
          showErrorToast(err);
        },
      },
    );
  };

  return (
    <section className="flex flex-col gap-8">
      <QueryStateBoundary query={assistantQuery}>
        {(assistant) => (
          <>
            <Breadcrumb>
              <BreadcrumbList>
                <BreadcrumbItem>
                  <BreadcrumbLink href="/assistants">Каталог</BreadcrumbLink>
                </BreadcrumbItem>
                <BreadcrumbSeparator />
                {assistant.categoryName && (
                  <>
                    <BreadcrumbItem>
                      <BreadcrumbLink
                        href={`/assistants?categoryId=${assistant.categoryId ?? ""}`}
                      >
                        {assistant.categoryName}
                      </BreadcrumbLink>
                    </BreadcrumbItem>
                    <BreadcrumbSeparator />
                  </>
                )}
                <BreadcrumbItem>
                  <BreadcrumbPage>{assistant.name}</BreadcrumbPage>
                </BreadcrumbItem>
              </BreadcrumbList>
            </Breadcrumb>

            <div className="grid gap-8 lg:grid-cols-[1fr_360px]">
              <div className="space-y-6">
                <header className="space-y-3">
                  <div className="flex flex-wrap items-center gap-2">
                    {assistant.categoryName && <Tag>{assistant.categoryName}</Tag>}
                    {assistant.tags.map((tag) => (
                      <Tag key={tag} variant="secondary">
                        {tag}
                      </Tag>
                    ))}
                    {assistant.isActive ? (
                      <Tag variant="success">Активен</Tag>
                    ) : (
                      <Tag variant="secondary">Неактивен</Tag>
                    )}
                  </div>
                  <div className="flex flex-wrap items-center justify-between gap-3">
                    <h1 className="text-4xl font-extrabold tracking-tight">
                      {assistant.name}
                    </h1>
                    <div className="flex flex-wrap items-center gap-2">
                      <AssistantFavoriteButton
                        assistantId={assistant.id}
                        isFavorite={assistant.isFavorite}
                      />
                      {isAdmin && (
                        <Button
                          variant="soft"
                          size="sm"
                          render={
                            <Link to={`/admin/assistants/${assistant.id}/edit`}>
                              <Pencil />
                              Редактировать
                            </Link>
                          }
                        />
                      )}
                    </div>
                  </div>
                  <p className="text-sm text-muted-foreground">
                    {assistant.description}
                  </p>
                  <div className="text-xs font-semibold text-muted-foreground">
                    Модель: {assistant.model}
                  </div>
                </header>

                {isAdmin && assistant.systemPrompt && (
                  <div className="space-y-2 rounded-2xl bg-secondary p-5">
                    <h2 className="text-[0.9375rem] font-bold">Системный промпт</h2>
                    <pre className="whitespace-pre-wrap font-mono text-xs text-muted-foreground">
                      {assistant.systemPrompt}
                    </pre>
                  </div>
                )}

                {!lastRun && (
                  <div className="rounded-2xl border border-dashed border-border p-12 text-center text-sm text-muted-foreground">
                    Здесь появится ответ ассистента после запуска.
                  </div>
                )}

                {lastRun && (
                  <div ref={resultRef} className="space-y-3 scroll-mt-6">
                    <div className="flex items-center justify-between">
                      <h2 className="text-xl font-extrabold">Результат</h2>
                      <div className="flex items-center gap-2">
                        {lastRun.status === "pending" && (
                          <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />
                        )}
                        <RunStatusBadge status={lastRun.status} />
                      </div>
                    </div>
                    {lastRun.status === "failed" ? (
                      <Alert variant="destructive">
                        <AlertCircle className="h-4 w-4" />
                        <AlertTitle>Ошибка</AlertTitle>
                        <AlertDescription>
                          {getRunErrorMessage(lastRun.error)}
                        </AlertDescription>
                      </Alert>
                    ) : (
                      <div className="space-y-2">
                        <div className="rounded-2xl bg-secondary p-4 text-[0.9375rem] leading-relaxed whitespace-pre-wrap">
                          {lastRun.output ||
                            (lastRun.status === "pending"
                              ? "Генерация началась..."
                              : "—")}
                        </div>
                        <RunFeedbackControl
                          run={lastRun}
                          currentUserId={user?.id}
                          onRunUpdated={setLastRun}
                        />
                      </div>
                    )}
                  </div>
                )}
              </div>

              <div className="space-y-6">
                <div className="space-y-4 rounded-2xl bg-secondary p-5">
                  <div>
                    <h2 className="text-xl font-extrabold">Запуск</h2>
                    <p className="mt-1 text-sm text-muted-foreground">
                      Опишите задачу — ассистент сгенерирует ответ.
                    </p>
                  </div>

                  <div className="space-y-2">
                    <Label htmlFor="user-prompt" className="text-[0.9375rem] font-bold">
                      Ваш запрос
                    </Label>
                    <Textarea
                      id="user-prompt"
                      rows={5}
                      placeholder={assistant.exampleUserPrompt ?? ""}
                      value={prompt}
                      onChange={(e) => setPrompt(e.target.value)}
                      disabled={!assistant.isActive || runMutation.isPending}
                    />
                    {assistant.exampleUserPrompt && (
                      <p className="text-xs text-muted-foreground">
                        Пример: {assistant.exampleUserPrompt}
                      </p>
                    )}
                  </div>

                  <div className="flex items-center gap-3">
                    <Switch
                      id="assistant-run-streaming"
                      checked={streaming}
                      onCheckedChange={setStreaming}
                      disabled={!assistant.isActive || runMutation.isPending}
                    />
                    <Label htmlFor="assistant-run-streaming">
                      Потоковый ответ
                    </Label>
                  </div>

                  <div className="flex gap-3">
                    <Button
                      variant="black"
                      size="lg"
                      onClick={handleRun}
                      disabled={!assistant.isActive || runMutation.isPending}
                    >
                      {runMutation.isPending ? (
                        <>
                          <Loader2 className="animate-spin" />
                          Запускаем...
                        </>
                      ) : (
                        "Запустить"
                      )}
                    </Button>
                    <Button
                      variant="soft"
                      size="lg"
                      onClick={() => setPrompt("")}
                      disabled={prompt === "" || runMutation.isPending}
                    >
                      Очистить
                    </Button>
                  </div>

                  {!assistant.isActive && (
                    <Alert>
                      <AlertCircle className="h-4 w-4" />
                      <AlertTitle>Ассистент выключен</AlertTitle>
                      <AlertDescription>Запуск недоступен.</AlertDescription>
                    </Alert>
                  )}
                </div>
              </div>
            </div>
          </>
        )}
      </QueryStateBoundary>
    </section>
  );
}
