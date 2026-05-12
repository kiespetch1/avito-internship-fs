import { useQuery } from "@tanstack/react-query";
import { categories } from "@/lib/api";
import { qk } from "@/lib/api/queryKeys.ts";

export function useCategories() {
  return useQuery({
    queryKey: qk.categories.all(),
    queryFn: ({ signal }) => categories.list(signal),
  });
}
