import { useMutation } from "@tanstack/react-query";
import { auth, type AuthCredentialsIn, type Role, type Token } from "@/lib/api";
import { saveAuth } from "@/lib/auth/storage";

function saveSession(session: Token): void {
  saveAuth({ token: session.token, user: session.user });
}

export function usePasswordLogin() {
  return useMutation({
    mutationFn: (input: AuthCredentialsIn) => auth.login(input),
    onSuccess: saveSession,
  });
}

export function useRegister() {
  return useMutation({
    mutationFn: (input: AuthCredentialsIn) => auth.register(input),
    onSuccess: saveSession,
  });
}

export function useDummyLogin() {
  return useMutation({
    mutationFn: (role: Role) => auth.dummyLogin(role),
    onSuccess: saveSession,
  });
}
