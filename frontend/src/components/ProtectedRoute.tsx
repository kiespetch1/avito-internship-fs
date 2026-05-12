import { useEffect, type ReactNode } from "react";
import { Navigate, useLocation } from "react-router-dom";
import { toast } from "sonner";
import { useAuth } from "@/lib/auth";
import type { Role } from "@/lib/api";

type Props = {
  children: ReactNode;
  role?: Role;
};

export function ProtectedRoute({ children, role }: Props) {
  const { isAuthenticated, user } = useAuth();
  const location = useLocation();

  const roleMismatch = role !== undefined && isAuthenticated && user?.role !== role;

  useEffect(() => {
    if (roleMismatch) {
      toast.error("Недостаточно прав для этой страницы");
    }
  }, [roleMismatch]);

  if (!isAuthenticated) {
    return <Navigate to="/login" replace state={{ from: location.pathname }} />;
  }

  if (roleMismatch) {
    return <Navigate to="/assistants" replace />;
  }

  return <>{children}</>;
}
