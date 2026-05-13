import { useQuery } from "@tanstack/react-query";
import { useEffect, useMemo, useState } from "react";
import { Link } from "react-router-dom";
import { Plus, Search, Star } from "lucide-react";
import { AssistantFavoriteButton } from "@/components/AssistantFavoriteButton";
import { PaginationControl } from "@/components/PaginationControl";
import { QueryStateBoundary } from "@/components/QueryStateBoundary";
import { Button } from "@/components/ui/button";
import { Card, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Skeleton } from "@/components/ui/skeleton";
import { Switch } from "@/components/ui/switch";
import { Tag } from "@/components/ui/tag";
import { assistants as assistantsApi } from "@/lib/api";
import { useCategories } from "@/lib/api/queries/useCategories";
import { qk } from "@/lib/api/queryKeys";
import { useAuth } from "@/lib/auth";
import { useDebouncedValue } from "@/lib/hooks/useDebouncedValue";
import {
  boolParam,
  intParam,
  stringParam,
  useSearchParamsState,
} from "@/lib/hooks/useSearchParamsState";

const CATALOG_PAGE_SIZE = 12;
const ALL_CATEGORIES = "__all__";

const filtersSchema = {
  q: stringParam(""),
  categoryId: stringParam(""),
  includeInactive: boolParam(false),
  favoriteOnly: boolParam(false),
  page: intParam(1),
};

export function CatalogPage() {
  const { user } = useAuth();
  const isAdmin = user?.role === "admin";

  const [filters, setFilters] = useSearchParamsState(filtersSchema);
  const [searchInput, setSearchInput] = useState(filters.q);
  const debouncedSearch = useDebouncedValue(searchInput, 300);

  useEffect(() => {
    if (debouncedSearch !== filters.q) {
      setFilters({ q: debouncedSearch, page: 1 });
    }
  }, [debouncedSearch, filters.q, setFilters]);

  const categoriesQuery = useCategories();

  const categoryLabels = useMemo<Record<string, string>>(() => {
    const map: Record<string, string> = { [ALL_CATEGORIES]: "Все категории" };
    for (const c of categoriesQuery.data?.categories ?? []) {
      map[c.id] = c.name;
    }
    return map;
  }, [categoriesQuery.data]);

  const listQuery = {
    q: filters.q || undefined,
    categoryId: filters.categoryId || undefined,
    includeInactive: isAdmin && filters.includeInactive ? true : undefined,
    favoriteOnly: filters.favoriteOnly ? true : undefined,
    page: filters.page,
    pageSize: CATALOG_PAGE_SIZE,
  };

  const assistantsQuery = useQuery({
    queryKey: qk.assistants.list(user?.id, listQuery),
    queryFn: ({ signal }) => assistantsApi.list(listQuery, signal),
    placeholderData: (prev) => prev,
  });

  return (
    <section className="flex flex-col gap-8">
      <header className="flex flex-wrap items-end justify-between gap-4">
        <div>
          <p className="text-sm font-semibold text-primary">AI-каталог</p>
          <h1 className="mt-1 text-4xl font-extrabold tracking-tight">Ассистенты</h1>
          <p className="mt-2 text-sm text-muted-foreground">
            Найдите подходящего AI-ассистента и запустите его с вашим контекстом.
          </p>
        </div>
        {isAdmin && (
          <div className="flex gap-2">
            <Button
              variant="soft"
              size="sm"
              render={
                <Link to="/admin/categories/new">
                  <Plus />
                  Категория
                </Link>
              }
            />
            <Button
              variant="black"
              size="sm"
              render={
                <Link to="/admin/assistants/new">
                  <Plus />
                  Ассистент
                </Link>
              }
            />
          </div>
        )}
      </header>

      <div className="flex flex-wrap gap-3">
        <div className="relative w-64">
          <Search className="absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            className="pl-9"
            placeholder="Поиск по названию или описанию"
            value={searchInput}
            onChange={(e) => setSearchInput(e.target.value)}
          />
        </div>

        <Select
          value={filters.categoryId === "" ? ALL_CATEGORIES : filters.categoryId}
          onValueChange={(v) =>
            setFilters({
              categoryId: v === null || v === ALL_CATEGORIES ? "" : v,
              page: 1,
            })
          }
        >
          <SelectTrigger className="w-44">
            <SelectValue placeholder="Категория" items={categoryLabels} />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value={ALL_CATEGORIES}>Все категории</SelectItem>
            {(categoriesQuery.data?.categories ?? []).map((c) => (
              <SelectItem key={c.id} value={c.id}>
                {c.name}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>

        {isAdmin && (
          <div className="flex items-center gap-2">
            <Switch
              id="include-inactive"
              checked={filters.includeInactive}
              onCheckedChange={(v) => setFilters({ includeInactive: v, page: 1 })}
            />
            <Label htmlFor="include-inactive">Показать неактивных</Label>
          </div>
        )}

        <div className="flex items-center gap-2">
          <Switch
            id="favorite-only"
            checked={filters.favoriteOnly}
            onCheckedChange={(v) => setFilters({ favoriteOnly: v, page: 1 })}
          />
          <Label htmlFor="favorite-only" className="inline-flex items-center gap-1.5">
            <Star className="size-4" />
            Только избранные
          </Label>
        </div>
      </div>

      <QueryStateBoundary
        query={assistantsQuery}
        loadingFallback={
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {[1, 2, 3, 4, 5, 6].map((i) => (
              <Card
                key={i}
                className="overflow-hidden rounded-2xl border bg-secondary p-0 shadow-none"
              >
                <Skeleton className="aspect-[4/3] w-full rounded-none" />
                <div className="space-y-2 p-4">
                  <Skeleton className="h-4 w-20" />
                  <Skeleton className="h-5 w-32" />
                  <Skeleton className="h-3 w-full" />
                  <Skeleton className="h-3 w-2/3" />
                </div>
              </Card>
            ))}
          </div>
        }
        isEmpty={(data) => data.assistants.length === 0}
        emptyFallback={
          <div className="rounded-2xl border border-dashed border-border p-12 text-center text-sm text-muted-foreground">
            {filters.favoriteOnly
              ? "В избранном пока нет ассистентов."
              : "По вашему запросу ничего не найдено."}
          </div>
        }
      >
        {(data) => (
          <div className="flex flex-col gap-6">
            <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
              {data.assistants.map((a) => (
                <Link
                  key={a.id}
                  to={`/assistants/${a.id}`}
                  className="block transition-transform hover:-translate-y-0.5"
                >
                  <Card
                    className={`relative h-full overflow-hidden rounded-2xl border bg-secondary p-0 shadow-none ${
                      a.isActive ? "" : "opacity-60"
                    }`}
                  >
                    <AssistantFavoriteButton
                      assistantId={a.id}
                      isFavorite={a.isFavorite}
                      compact
                      className="absolute top-3 right-3 z-10 bg-card/90 backdrop-blur hover:bg-card"
                    />
                    <div className="space-y-2 p-4">
                      <div className="flex flex-wrap items-center gap-1.5">
                        <Tag>{a.categoryName ?? "Без категории"}</Tag>
                        {a.isActive ? (
                          <Tag variant="success">Активен</Tag>
                        ) : (
                          <Tag variant="secondary">Неактивен</Tag>
                        )}
                      </div>
                      <CardTitle className="text-base font-bold">{a.name}</CardTitle>
                      <p className="line-clamp-2 text-xs text-muted-foreground">
                        {a.description}
                      </p>
                      <div className="pt-1 text-xs font-semibold text-muted-foreground">
                        {a.model}
                      </div>
                    </div>
                  </Card>
                </Link>
              ))}
            </div>

            <PaginationControl
              pagination={data.pagination}
              onChange={(page) => setFilters({ page })}
            />
          </div>
        )}
      </QueryStateBoundary>
    </section>
  );
}
