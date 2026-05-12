import { useMutation, useQueryClient } from "@tanstack/react-query";
import { categories, type CategoryCreateIn } from "@/lib/api";
import { qk } from "@/lib/api/queryKeys.ts";

export function useCreateCategory() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: CategoryCreateIn) => categories.create(input),
    onSuccess: () => qc.invalidateQueries({ queryKey: qk.categories.all() }),
  });
}
