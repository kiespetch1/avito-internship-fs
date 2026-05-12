import { useState } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
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
  const [pending, setPending] = useState<Role | null>(null);

  const handleLogin = async (role: Role) => {
    setPending(role);
    try {
      await login(role);
      navigate(from, { replace: true });
    } catch (err) {
      const message = err instanceof ApiError ? err.message : "Не удалось войти";
      toast.error(message);
    } finally {
      setPending(null);
    }
  };

  return (
    <div className="flex min-h-screen items-center justify-center bg-background px-6">
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle>Вход в каталог</CardTitle>
          <CardDescription>
            Тестовый вход через <code>/dummyLogin</code>. Выберите роль.
          </CardDescription>
        </CardHeader>
        <CardContent className="flex flex-col gap-3">
          <Button
            onClick={() => handleLogin("user")}
            disabled={pending !== null}
          >
            {pending === "user" ? "Входим..." : "Войти как пользователь"}
          </Button>
          <Button
            variant="outline"
            onClick={() => handleLogin("admin")}
            disabled={pending !== null}
          >
            {pending === "admin" ? "Входим..." : "Войти как администратор"}
          </Button>
        </CardContent>
      </Card>
    </div>
  );
}
