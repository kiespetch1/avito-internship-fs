import {
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useSyncExternalStore,
  type ReactNode,
} from "react";
import { useQueryClient } from "@tanstack/react-query";
import { AuthContext, type AuthState } from "./context";
import { clearAuth, loadAuth, subscribeAuth } from "./storage";

type Props = { children: ReactNode };

export function AuthProvider({ children }: Props) {
  const queryClient = useQueryClient();
  const snap = useSyncExternalStore(subscribeAuth, loadAuth, loadAuth);
  const authScope = snap?.user.id ?? null;
  const prevAuthScopeRef = useRef<string | null>(authScope);

  const logout = useCallback(() => {
    clearAuth();
  }, []);

  const value = useMemo<AuthState>(
    () => ({
      user: snap?.user ?? null,
      token: snap?.token ?? null,
      isAuthenticated: snap !== null,
      logout,
    }),
    [snap, logout],
  );

  useEffect(() => {
    if (prevAuthScopeRef.current !== authScope) {
      queryClient.clear();
      prevAuthScopeRef.current = authScope;
    }
  }, [authScope, queryClient]);

  return <AuthContext value={value}>{children}</AuthContext>;
}
