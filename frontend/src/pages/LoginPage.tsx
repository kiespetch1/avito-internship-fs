import { useState } from "react";
import { useForm } from "@tanstack/react-form";
import { KeyRound, Loader2, LogIn, TestTube2, UserPlus } from "lucide-react";
import { useLocation, useNavigate } from "react-router-dom";
import { toast } from "sonner";
import { FormField } from "@/components/FormField";
import { ThemeToggle } from "@/components/ThemeToggle";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import { ApiError, type Role } from "@/lib/api";
import {
  useDummyLogin,
  usePasswordLogin,
  useRegister,
} from "@/lib/api/queries/useAuthMutations";
import { fieldErrorMessage } from "@/lib/forms/fieldError";
import {
  authCredentialsSchema,
  type AuthCredentialsFormValues,
} from "@/lib/forms/schemas/authSchema";

type LocationState = { from?: string };
type AuthMode = "password" | "dummy";
type PasswordMode = "login" | "register";

const defaults: AuthCredentialsFormValues = { email: "", password: "" };

function isLocationState(value: unknown): value is LocationState {
  if (typeof value !== "object" || value === null) return false;
  if (!("from" in value)) return true;
  const { from } = value;
  return from === undefined || typeof from === "string";
}

function isRole(value: string): value is Role {
  return value === "admin" || value === "user";
}

function authErrorMessage(err: unknown, fallback: string): string {
  return err instanceof ApiError ? err.message : fallback;
}

export function LoginPage() {
  const navigate = useNavigate();
  const location = useLocation();
  const from =
    (isLocationState(location.state) && location.state.from) || "/assistants";
  const [authMode, setAuthMode] = useState<AuthMode>("password");
  const [passwordMode, setPasswordMode] = useState<PasswordMode>("login");

  const finishAuth = () => navigate(from, { replace: true });

  const title =
    authMode === "dummy"
      ? "Тестовый вход"
      : passwordMode === "register"
        ? "Регистрация"
        : "Войти";

  return (
    <div className="relative flex min-h-screen items-center justify-center bg-background px-6">
      <div className="absolute right-6 top-6">
        <ThemeToggle />
      </div>
      <div className="w-full max-w-md space-y-6">
        <div>
          <p className="text-sm font-semibold text-primary">AI-каталог</p>
          <h1 className="mt-1 text-4xl font-extrabold tracking-tight">{title}</h1>
        </div>

        {authMode === "password" ? (
          <div className="space-y-5">
            {passwordMode === "login" ? (
              <PasswordLoginForm
                onDone={finishAuth}
                onRegister={() => setPasswordMode("register")}
              />
            ) : (
              <RegisterForm
                onDone={finishAuth}
                onLogin={() => setPasswordMode("login")}
              />
            )}
            <div className="border-t pt-4">
              <Button
                type="button"
                variant="soft"
                size="lg"
                className="w-full"
                onClick={() => setAuthMode("dummy")}
              >
                <TestTube2 />
                Войти через dummyLogin
              </Button>
            </div>
          </div>
        ) : (
          <DummyLoginForm
            onDone={finishAuth}
            onPasswordLogin={() => setAuthMode("password")}
          />
        )}
      </div>
    </div>
  );
}

function PasswordLoginForm({
  onDone,
  onRegister,
}: {
  onDone: () => void;
  onRegister: () => void;
}) {
  const loginMutation = usePasswordLogin();
  const form = useForm({
    defaultValues: defaults,
    validators: { onChange: authCredentialsSchema },
    onSubmit: async ({ value }) => {
      const parsed = authCredentialsSchema.parse(value);
      try {
        await loginMutation.mutateAsync(parsed);
        onDone();
      } catch (err) {
        toast.error(authErrorMessage(err, "Не удалось войти"));
      }
    },
  });

  return (
    <form
      className="space-y-5"
      onSubmit={(e) => {
        e.preventDefault();
        void form.handleSubmit();
      }}
    >
      <div className="space-y-4">
        <form.Field name="email">
          {(field) => (
            <FormField label="Почта" error={fieldErrorMessage(field.state.meta.errors)}>
              {(p) => (
                <Input
                  {...p}
                  type="email"
                  autoComplete="username"
                  value={field.state.value}
                  onBlur={field.handleBlur}
                  onChange={(e) => field.handleChange(e.target.value)}
                  placeholder="user@example.com"
                />
              )}
            </FormField>
          )}
        </form.Field>

        <form.Field name="password">
          {(field) => (
            <FormField
              label="Пароль"
              error={fieldErrorMessage(field.state.meta.errors)}
            >
              {(p) => (
                <Input
                  {...p}
                  type="password"
                  autoComplete="current-password"
                  value={field.state.value}
                  onBlur={field.handleBlur}
                  onChange={(e) => field.handleChange(e.target.value)}
                  placeholder="Минимум 8 символов"
                />
              )}
            </FormField>
          )}
        </form.Field>
      </div>
      <form.Subscribe
        selector={(s) => ({ canSubmit: s.canSubmit, isSubmitting: s.isSubmitting })}
      >
        {({ canSubmit, isSubmitting }) => {
          const pending = isSubmitting || loginMutation.isPending;
          return (
            <div className="space-y-3">
              <Button
                type="submit"
                variant="black"
                size="lg"
                className="w-full"
                disabled={!canSubmit || pending}
              >
                {pending ? <Loader2 className="animate-spin" /> : <LogIn />}
                {pending ? "Входим..." : "Войти"}
              </Button>
              <Button
                type="button"
                variant="ghost"
                size="lg"
                className="w-full"
                onClick={onRegister}
                disabled={pending}
              >
                <UserPlus />
                Зарегистрироваться
              </Button>
            </div>
          );
        }}
      </form.Subscribe>
    </form>
  );
}

function RegisterForm({
  onDone,
  onLogin,
}: {
  onDone: () => void;
  onLogin: () => void;
}) {
  const registerMutation = useRegister();
  const form = useForm({
    defaultValues: defaults,
    validators: { onChange: authCredentialsSchema },
    onSubmit: async ({ value }) => {
      const parsed = authCredentialsSchema.parse(value);
      try {
        await registerMutation.mutateAsync(parsed);
        onDone();
      } catch (err) {
        toast.error(authErrorMessage(err, "Не удалось зарегистрироваться"));
      }
    },
  });

  return (
    <form
      className="space-y-5"
      onSubmit={(e) => {
        e.preventDefault();
        void form.handleSubmit();
      }}
    >
      <div className="space-y-4">
        <form.Field name="email">
          {(field) => (
            <FormField label="Почта" error={fieldErrorMessage(field.state.meta.errors)}>
              {(p) => (
                <Input
                  {...p}
                  type="email"
                  autoComplete="username"
                  value={field.state.value}
                  onBlur={field.handleBlur}
                  onChange={(e) => field.handleChange(e.target.value)}
                  placeholder="user@example.com"
                />
              )}
            </FormField>
          )}
        </form.Field>

        <form.Field name="password">
          {(field) => (
            <FormField
              label="Пароль"
              error={fieldErrorMessage(field.state.meta.errors)}
            >
              {(p) => (
                <Input
                  {...p}
                  type="password"
                  autoComplete="new-password"
                  value={field.state.value}
                  onBlur={field.handleBlur}
                  onChange={(e) => field.handleChange(e.target.value)}
                  placeholder="Минимум 8 символов"
                />
              )}
            </FormField>
          )}
        </form.Field>
      </div>
      <form.Subscribe
        selector={(s) => ({ canSubmit: s.canSubmit, isSubmitting: s.isSubmitting })}
      >
        {({ canSubmit, isSubmitting }) => {
          const pending = isSubmitting || registerMutation.isPending;
          return (
            <div className="space-y-3">
              <Button
                type="submit"
                variant="black"
                size="lg"
                className="w-full"
                disabled={!canSubmit || pending}
              >
                {pending ? <Loader2 className="animate-spin" /> : <UserPlus />}
                {pending ? "Создаём..." : "Создать аккаунт"}
              </Button>
              <Button
                type="button"
                variant="ghost"
                size="lg"
                className="w-full"
                onClick={onLogin}
                disabled={pending}
              >
                <KeyRound />
                Уже есть аккаунт
              </Button>
            </div>
          );
        }}
      </form.Subscribe>
    </form>
  );
}

function DummyLoginForm({
  onDone,
  onPasswordLogin,
}: {
  onDone: () => void;
  onPasswordLogin: () => void;
}) {
  const [role, setRole] = useState<Role>("user");
  const dummyMutation = useDummyLogin();

  const handleLogin = async () => {
    try {
      await dummyMutation.mutateAsync(role);
      onDone();
    } catch (err) {
      toast.error(authErrorMessage(err, "Не удалось войти"));
    }
  };

  return (
    <div className="space-y-5">
      <RadioGroup
        value={role}
        onValueChange={(value) => {
          if (isRole(value)) setRole(value);
        }}
        className="gap-2"
      >
        <Label
          htmlFor="role-user"
          className="flex cursor-pointer items-start gap-3 rounded-lg bg-secondary p-4 has-data-checked:ring-2 has-data-checked:ring-avito-black"
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
          className="flex cursor-pointer items-start gap-3 rounded-lg bg-secondary p-4 has-data-checked:ring-2 has-data-checked:ring-avito-black"
        >
          <div className="flex-1 space-y-1">
            <div className="text-[0.9375rem] font-bold">Администратор</div>
            <p className="text-sm text-muted-foreground">
              Управление категориями и ассистентами
            </p>
          </div>
          <RadioGroupItem value="admin" id="role-admin" />
        </Label>
      </RadioGroup>

      <div className="space-y-3">
        <Button
          variant="black"
          size="lg"
          className="w-full"
          onClick={handleLogin}
          disabled={dummyMutation.isPending}
        >
          {dummyMutation.isPending ? (
            <Loader2 className="animate-spin" />
          ) : (
            <TestTube2 />
          )}
          {dummyMutation.isPending
            ? "Входим..."
            : `Войти как ${role === "admin" ? "Admin" : "User"}`}
        </Button>
        <Button
          type="button"
          variant="soft"
          size="lg"
          className="w-full"
          onClick={onPasswordLogin}
          disabled={dummyMutation.isPending}
        >
          <KeyRound />
          Войти по email и паролю
        </Button>
      </div>
    </div>
  );
}
