import type { MouseEvent } from "react";
import { Star } from "lucide-react";
import { Button } from "@/components/ui/button";
import { showErrorToast } from "@/lib/api/errorMessage";
import { useSetAssistantFavorite } from "@/lib/api/queries/useSetAssistantFavorite";
import { cn } from "@/lib/utils";

type Props = {
  assistantId: string;
  isFavorite: boolean;
  className?: string;
  compact?: boolean;
};

export function AssistantFavoriteButton({
  assistantId,
  isFavorite,
  className,
  compact = false,
}: Props) {
  const mutation = useSetAssistantFavorite();
  const label = isFavorite ? "Убрать из избранного" : "Добавить в избранное";

  const handleClick = (event: MouseEvent<HTMLButtonElement>) => {
    event.preventDefault();
    event.stopPropagation();
    mutation.mutate(
      { assistantId, isFavorite: !isFavorite },
      {
        onError: (err) => {
          showErrorToast(err, { fallback: "Не удалось обновить избранное" });
        },
      },
    );
  };

  return (
    <Button
      type="button"
      variant="soft"
      size={compact ? "icon-sm" : "sm"}
      className={cn(
        compact ? "shadow-sm" : "",
        isFavorite && "bg-yellow-50 text-yellow-600 hover:bg-yellow-100",
        className,
      )}
      aria-label={label}
      title={label}
      disabled={mutation.isPending}
      onClick={handleClick}
    >
      <Star
        className={cn(
          "transition-colors",
          isFavorite
            ? "fill-yellow-400 text-yellow-500"
            : "text-foreground"
        )}
      />      {!compact && <span>{isFavorite ? "В избранном" : "В избранное"}</span>}
    </Button>
  );
}
