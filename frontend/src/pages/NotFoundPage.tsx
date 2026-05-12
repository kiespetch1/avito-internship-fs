import { useNavigate } from "react-router-dom";
import { Button } from "@/components/ui/button";

export function NotFoundPage() {
  const navigate = useNavigate();
  return (
    <section className="flex flex-col items-start gap-4">
      <h1 className="text-2xl font-semibold tracking-tight">Страница не найдена</h1>
      <p className="text-sm text-muted-foreground">Такой страницы нет в каталоге.</p>
      <Button onClick={() => navigate("/assistants")}>На главную</Button>
    </section>
  );
}
