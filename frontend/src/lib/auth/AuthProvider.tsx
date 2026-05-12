import { useCallback, useMemo, useSyncExternalStore, type ReactNode } from "react";
import { auth as authApi, type Role } from "@/lib/api";
import { AuthContext, type AuthState } from "./context";
import { clearAuth, loadAuth, saveAuth, subscribeAuth } from "./storage";

type Props = { children: ReactNode };

export function AuthProvider({ children }: Props) {
  const snap = useSyncExternalStore(subscribeAuth, loadAuth, loadAuth);

  const login = useCallback(async (role: Role) => {
    const res = await authApi.dummyLogin(role);
    saveAuth({ token: res.token, user: res.user });
  }, []);

  const logout = useCallback(() => {
    clearAuth();
  }, []);

  const value = useMemo<AuthState>(
    () => ({
      user: snap?.user ?? null,
      token: snap?.token ?? null,
      isAuthenticated: snap !== null,
      login,
      logout,
    }),
    [snap, login, logout],
  );

  return <AuthContext value={value}>{children}</AuthContext>;
}
