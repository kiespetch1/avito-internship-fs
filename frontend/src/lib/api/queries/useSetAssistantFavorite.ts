import { useMutation, useQueryClient } from "@tanstack/react-query";
import { assistants } from "@/lib/api";
import { qk } from "@/lib/api/queryKeys";

type SetAssistantFavoriteInput = {
  assistantId: string;
  isFavorite: boolean;
};

export function useSetAssistantFavorite() {
  const qc = useQueryClient();

  return useMutation({
    mutationFn: ({ assistantId, isFavorite }: SetAssistantFavoriteInput) =>
      isFavorite
        ? assistants.addFavorite(assistantId)
        : assistants.removeFavorite(assistantId),
    onSuccess: () => qc.invalidateQueries({ queryKey: qk.assistants.root() }),
  });
}
