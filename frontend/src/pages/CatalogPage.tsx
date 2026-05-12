import { useQuery } from "@tanstack/react-query";
import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { PaginationControl } from "@/components/PaginationControl";
import { QueryStateBoundary } from "@/components/QueryStateBoundary";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";
import { assistants as assistantsApi, categories as categoriesApi } from "@/lib/api";
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

  const categoriesQuery = useQuery({
    queryKey: qk.categories.all(),
    queryFn: ({ signal }) => categoriesApi.list(signal),
  });

  const listQuery = {
    q: filters.q || undefined,
    categoryId: filters.categoryId || undefined,
    includeInactive: isAdmin && filters.includeInactive ? true : undefined,
    page: filters.page,
    pageSize: CATALOG_PAGE_SIZE,
  };

  const assistantsQuery = useQuery({
    queryKey: qk.assistants.list(listQuery),
    queryFn: ({ signal }) => assistantsApi.list(listQuery, signal),
    placeholderData: (prev) => prev,
  });

  return (
    <section className="flex flex-col gap-6">
      <header className="flex flex-col gap-2">
        <h1 className="text-2xl font-semibold tracking-tight">Каталог ассистентов</h1>
        <p className="text-sm text-muted-foreground">
          Найдите подходящего AI-ассистента и запустите его с вашим контекстом.
        </p>
      </header>

      <div className="grid gap-3 rounded-md border border-border p-4 sm:grid-cols-[1fr_220px_auto]">
        <div className="flex flex-col gap-1.5">
          <Label htmlFor="catalog-search">Поиск</Label>
          <Input
            id="catalog-search"
            placeholder="По названию или описанию"
            value={searchInput}
            onChange={(e) => setSearchInput(e.target.value)}
          />
        </div>

        <div className="flex flex-col gap-1.5">
          <Label>Категория</Label>
          <Select
            value={filters.categoryId === "" ? ALL_CATEGORIES : filters.categoryId}
            onValueChange={(v) =>
              setFilters({
                categoryId: v === null || v === ALL_CATEGORIES ? "" : v,
                page: 1,
              })
            }
          >
            <SelectTrigger>
              <SelectValue placeholder="Все категории" />
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
        </div>

        {isAdmin && (
          <div className="flex items-end gap-2">
            <Switch
              id="include-inactive"
              checked={filters.includeInactive}
              onCheckedChange={(v) => setFilters({ includeInactive: v, page: 1 })}
            />
            <Label htmlFor="include-inactive" className="cursor-pointer">
              Неактивные
            </Label>
          </div>
        )}
      </div>

      <QueryStateBoundary
        query={assistantsQuery}
        isEmpty={(data) => data.assistants.length === 0}
        emptyFallback={
          <div className="rounded-md border border-dashed border-border p-8 text-center text-sm text-muted-foreground">
            По вашему запросу ничего не найдено.
          </div>
        }
      >
        {(data) => (
          <div className="flex flex-col gap-4">
            <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
              {data.assistants.map((a) => (
                <Link key={a.id} to={`/assistants/${a.id}`} className="block">
                  <Card className="h-full transition-colors hover:border-primary">
                    <CardHeader>
                      <div className="flex items-center justify-between gap-2">
                        <CardTitle className="text-base">{a.name}</CardTitle>
                        {!a.isActive && <Badge variant="destructive">Выключен</Badge>}
                      </div>
                      <CardDescription className="line-clamp-2">{a.description}</CardDescription>
                    </CardHeader>
                    <CardContent className="flex items-center justify-between text-xs text-muted-foreground">
                      <span>{a.categoryName ?? "—"}</span>
                      <span>{a.model}</span>
                    </CardContent>
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
