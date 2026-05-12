import { useState } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { Loader2 } from "lucide-react";
import { toast } from "sonner";
import { ThemeToggle } from "@/components/ThemeToggle";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import { useAuth } from "@/lib/auth";
import { ApiError, type Role } from "@/lib/api";

type LocationState = { from?: string };

function isLocationState(value: unknown): value is LocationState {
  if (typeof value !== "object" || value === null) return false;
  if (!("from" in value)) return true;
  const { from } = value;
  return from === undefined || typeof from === "string";
}

export function LoginPage() {
  const { login } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const from = (isLocationState(location.state) && location.state.from) || "/assistants";
  const [role, setRole] = useState<Role>("user");
  const [pending, setPending] = useState(false);

  const handleLogin = async () => {
    setPending(true);
    try {
      await login(role);
      navigate(from, { replace: true });
    } catch (err) {
      const message = err instanceof ApiError ? err.message : "Не удалось войти";
      toast.error(message);
    } finally {
      setPending(false);
    }
  };

  return (
    <div className="relative flex min-h-screen items-center justify-center bg-background px-6">
      <div className="absolute right-6 top-6">
        <ThemeToggle />
      </div>
      <div className="w-full max-w-md space-y-6">
        <div>
          <p className="text-sm font-semibold text-primary">AI-каталог</p>
          <h1 className="mt-1 text-4xl font-extrabold tracking-tight">Войти</h1>
          <p className="mt-2 text-sm text-muted-foreground">
            Тестовый вход через <code>/dummyLogin</code>. Выберите роль.
          </p>
        </div>

        <RadioGroup
          value={role}
          onValueChange={(value) => setRole(value as Role)}
          className="gap-2"
        >
          <Label
            htmlFor="role-user"
            className="flex cursor-pointer items-start gap-3 rounded-2xl bg-secondary p-4 has-data-checked:ring-2 has-data-checked:ring-avito-black"
          >
            <div className="flex-1 space-y-1">
              <div className="text-[0.9375rem] font-bold">Пользователь</div>
              <p className="text-sm text-muted-foreground">
                Просмотр каталога и запуск ассистентов
              </p>
            </div>
            <RadioGroupItem value="user" id="role-user" />
          </Label>
          <Label
            htmlFor="role-admin"
            className="flex cursor-pointer items-start gap-3 rounded-2xl bg-secondary p-4 has-data-checked:ring-2 has-data-checked:ring-avito-black"
          >
            <div className="flex-1 space-y-1">
              <div className="text-[0.9375rem] font-bold">Администратор</div>
              <p className="text-sm text-muted-foreground">
                Управление категориями и ассистентами, история всех запусков
              </p>
            </div>
            <RadioGroupItem value="admin" id="role-admin" />
          </Label>
        </RadioGroup>

        <Button
          variant="black"
          size="lg"
          className="w-full"
          onClick={handleLogin}
          disabled={pending}
        >
          {pending ? (
            <>
              <Loader2 className="animate-spin" />
              Входим...
            </>
          ) : (
            <>Войти как {role === "admin" ? "Admin" : "User"}</>
          )}
        </Button>
      </div>
    </div>
  );
}
