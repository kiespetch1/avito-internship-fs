import { useQuery } from "@tanstack/react-query";
import { categories } from "@/lib/api";
import { qk } from "@/lib/api/queryKeys.ts";

export type CategoryOption = { value: string; label: string };

export function useCategoryOptions() {
  return useQuery({
    queryKey: qk.categories.all(),
    queryFn: ({ signal }) => categories.list(signal),
    select: (data): CategoryOption[] =>
      data.categories.map((c) => ({ value: c.id, label: c.name })),
  });
}
