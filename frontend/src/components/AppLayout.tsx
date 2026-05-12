import { Link, NavLink, Outlet, useNavigate } from "react-router-dom";
import { Button } from "@/components/ui/button";
import { useAuth } from "@/lib/auth";

export function AppLayout() {
  const { user, logout } = useAuth();
  const navigate = useNavigate();

  const handleLogout = () => {
    logout();
    navigate("/login", { replace: true });
  };

  return (
    <div className="min-h-screen bg-background text-foreground">
      <header className="border-b border-border bg-card">
        <div className="mx-auto flex h-14 max-w-6xl items-center justify-between gap-6 px-6">
          <Link to="/assistants" className="flex items-baseline gap-2">
            <span className="text-sm font-semibold text-primary">Авито</span>
            <span className="text-sm text-muted-foreground">AI Assistants</span>
          </Link>

          <nav className="flex items-center gap-1 text-sm">
            <NavItem to="/assistants" label="Каталог" end />
            <NavItem to="/runs/my" label="Мои запуски" />
            {user?.role === "admin" && <NavItem to="/admin/runs" label="Все запуски" />}
          </nav>

          <div className="flex items-center gap-3">
            {user && (
              <span className="text-sm text-muted-foreground">
                {user.email} · {user.role}
              </span>
            )}
            <Button variant="outline" size="sm" onClick={handleLogout}>
              Выйти
            </Button>
          </div>
        </div>
      </header>

      <main className="mx-auto max-w-6xl px-6 py-8">
        <Outlet />
      </main>
    </div>
  );
}

function NavItem({ to, label, end }: { to: string; label: string; end?: boolean }) {
  return (
    <NavLink
      to={to}
      end={end}
      className={({ isActive }) =>
        [
          "rounded-md px-3 py-1.5 transition-colors",
          isActive
            ? "bg-accent text-accent-foreground"
            : "text-muted-foreground hover:text-foreground",
        ].join(" ")
      }
    >
      {label}
    </NavLink>
  );
}
