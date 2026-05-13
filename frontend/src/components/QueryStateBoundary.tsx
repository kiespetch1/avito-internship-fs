import { type ReactNode } from "react";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import {getErrorMessage} from "@/lib/api/errorMessage.ts";

type Props<T> = {
  query: {
    data: T | undefined;
    isPending: boolean;
    isError: boolean;
    error: unknown;
    refetch: () => void;
  };
  isEmpty?: (data: T) => boolean;
  loadingFallback?: ReactNode;
  emptyFallback?: ReactNode;
  children: (data: T) => ReactNode;
};

export function QueryStateBoundary<T>({
  query,
  isEmpty,
  loadingFallback,
  emptyFallback,
  children,
}: Props<T>) {
  if (query.isPending) {
    return <>{loadingFallback ?? <DefaultLoading />}</>;
  }

  if (query.isError) {
    return <ErrorState error={query.error} onRetry={query.refetch} />;
  }

  if (query.data === undefined) return null;

  if (isEmpty && isEmpty(query.data)) {
    return <>{emptyFallback ?? <DefaultEmpty />}</>;
  }

  return <>{children(query.data)}</>;
}

function DefaultLoading() {
  return (
    <div className="flex flex-col gap-3">
      <Skeleton className="h-8 w-1/3" />
      <Skeleton className="h-24 w-full" />
      <Skeleton className="h-24 w-full" />
    </div>
  );
}

function DefaultEmpty() {
  return (
    <div className="rounded-md border border-dashed border-border p-8 text-center text-sm text-muted-foreground">
      Здесь пока ничего нет.
    </div>
  );
}

function ErrorState({ error, onRetry }: { error: unknown; onRetry: () => void }) {
  const message = getErrorMessage(error);

  return (
    <Alert variant="destructive" className="flex items-start justify-between gap-4">
      <div>
        <AlertTitle>Ошибка</AlertTitle>
        <AlertDescription>{message}</AlertDescription>
      </div>
      <Button size="sm" variant="outline" onClick={onRetry}>
        Повторить
      </Button>
    </Alert>
  );
}
