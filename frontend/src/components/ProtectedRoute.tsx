import { type ReactNode } from "react";
import { Navigate, useLocation } from "react-router-dom";
import { useAuth } from "@/lib/auth";
import type { Role } from "@/lib/api";

type Props = {
  children: ReactNode;
  role?: Role;
};

export function ProtectedRoute({ children, role }: Props) {
  const { isAuthenticated, user } = useAuth();
  const location = useLocation();

  if (!isAuthenticated) {
    return <Navigate to="/login" replace state={{ from: location.pathname }} />;
  }

  if (role && user?.role !== role) {
    return <Navigate to="/assistants" replace />;
  }

  return <>{children}</>;
}
