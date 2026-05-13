import { createContext } from "react";
import type { User } from "@/lib/api";

export type AuthState = {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  logout: () => void;
};

export const AuthContext = createContext<AuthState | null>(null);
