import { createContext } from "react";
import type { Role, User } from "@/lib/api";

export type AuthState = {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  login: (role: Role) => Promise<void>;
  logout: () => void;
};

export const AuthContext = createContext<AuthState | null>(null);
