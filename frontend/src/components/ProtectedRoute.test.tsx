import { render, screen } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";
import type { Role, User } from "@/lib/api";
import type { AuthState } from "@/lib/auth/context";
import { ProtectedRoute } from "./ProtectedRoute";

const toastError = vi.fn();
vi.mock("sonner", () => ({
  toast: {
    error: (msg: string) => toastError(msg),
  },
}));

const useAuthMock = vi.fn<() => AuthState>();
vi.mock("@/lib/auth", () => ({
  useAuth: () => useAuthMock(),
}));

function setAuth(user: User | null) {
  useAuthMock.mockReturnValue({
    user,
    token: user ? "token" : null,
    isAuthenticated: user !== null,
    login: vi.fn(),
    logout: vi.fn(),
  });
}

function makeUser(role: Role): User {
  return { id: "u1", email: `${role}@example.com`, role, createdAt: null };
}

function renderAt(initialPath: string, role?: Role) {
  return render(
    <MemoryRouter initialEntries={[initialPath]}>
      <Routes>
        <Route path="/login" element={<div>login-page</div>} />
        <Route path="/assistants" element={<div>catalog-page</div>} />
        <Route
          path="/admin/runs"
          element={
            <ProtectedRoute role={role}>
              <div>admin-runs-page</div>
            </ProtectedRoute>
          }
        />
        <Route
          path="/runs/my"
          element={
            <ProtectedRoute>
              <div>my-runs-page</div>
            </ProtectedRoute>
          }
        />
      </Routes>
    </MemoryRouter>,
  );
}

describe("ProtectedRoute", () => {
  afterEach(() => {
    toastError.mockClear();
    useAuthMock.mockReset();
  });

  it("redirects unauthenticated users to /login", () => {
    setAuth(null);
    renderAt("/runs/my");
    expect(screen.getByText("login-page")).toBeInTheDocument();
  });

  it("renders children when authenticated and no role is required", () => {
    setAuth(makeUser("user"));
    renderAt("/runs/my");
    expect(screen.getByText("my-runs-page")).toBeInTheDocument();
  });

  it("redirects to /assistants and shows a toast when role does not match", () => {
    setAuth(makeUser("user"));
    renderAt("/admin/runs", "admin");
    expect(screen.getByText("catalog-page")).toBeInTheDocument();
    expect(toastError).toHaveBeenCalledWith("Недостаточно прав для этой страницы");
  });

  it("renders admin children when role matches", () => {
    setAuth(makeUser("admin"));
    renderAt("/admin/runs", "admin");
    expect(screen.getByText("admin-runs-page")).toBeInTheDocument();
    expect(toastError).not.toHaveBeenCalled();
  });
});
