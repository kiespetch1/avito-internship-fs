import { useNavigate } from "react-router-dom";
import { Compass } from "lucide-react";
import { Button } from "@/components/ui/button";

type Props = { fullScreen?: boolean };

export function NotFoundPage({ fullScreen = false }: Props) {
  const navigate = useNavigate();

  const content = (
    <div className="w-full max-w-md space-y-6 text-center">
      <div className="mx-auto flex size-20 items-center justify-center rounded-3xl bg-secondary">
        <Compass className="size-10 text-avito-purple" />
      </div>

      <div>
        <p className="bg-gradient-to-br from-primary to-avito-purple bg-clip-text text-7xl font-extrabold tracking-tight text-transparent">
          404
        </p>
        <h1 className="mt-4 text-3xl font-extrabold tracking-tight">
          Страница не найдена
        </h1>
        <p className="mt-2 text-sm text-muted-foreground">
          Возможно, она была удалена или вы перешли по неверной ссылке.
        </p>
      </div>

      <div className="flex flex-col gap-2 sm:flex-row sm:justify-center">
        <Button
          variant="black"
          size="lg"
          onClick={() => navigate("/assistants")}
        >
          На главную
        </Button>
        <Button variant="soft" size="lg" onClick={() => navigate(-1)}>
          Назад
        </Button>
      </div>
    </div>
  );

  if (fullScreen) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-background px-6">
        {content}
      </div>
    );
  }

  return <div className="flex justify-center py-16">{content}</div>;
}
