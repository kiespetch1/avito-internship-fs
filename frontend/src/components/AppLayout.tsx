import { useState, type ReactNode } from "react";
import { Link, NavLink, Outlet, useNavigate } from "react-router-dom";
import { ThemeToggle } from "@/components/ThemeToggle";
import { Button } from "@/components/ui/button";
import { useAuth } from "@/lib/auth";

type Props = { children?: ReactNode };

export function AppLayout({ children }: Props) {
  const { user, logout } = useAuth();
  const navigate = useNavigate();
  const [isMenuOpen, setIsMenuOpen] = useState(false);

  const handleLogout = () => {
    logout();
    navigate("/login", { replace: true });
  };

  const closeMenu = () => {
    setIsMenuOpen(false);
  };

  return (
    <div className="min-h-screen bg-background text-foreground">
      <header className="sticky top-0 z-40 border-b border-border bg-card/80 backdrop-blur-md supports-[backdrop-filter]:bg-card/70">
        <div className="mx-auto flex h-14 max-w-6xl items-center justify-between gap-4 px-4 sm:px-6">
          <Link
            to="/assistants"
            className="flex min-w-0 items-baseline gap-2"
            onClick={closeMenu}
          >
            <span className="truncate text-sm text-muted-foreground">
              AI ассистенты
            </span>
          </Link>

          <nav className="hidden items-center gap-1 text-sm md:flex">
            <NavItem to="/assistants" label="Каталог" end />
            <NavItem to="/runs/my" label="Мои запуски" />
            {user?.role === "admin" && (
              <NavItem to="/admin/runs" label="Все запуски" />
            )}
          </nav>

          <div className="hidden items-center gap-3 md:flex">
            <ThemeToggle />
            <Button variant="outline" size="sm" onClick={handleLogout}>
              Выйти
            </Button>
          </div>

          <div className="flex items-center gap-2 md:hidden">
            <ThemeToggle />

            <Button
              variant="outline"
              size="sm"
              type="button"
              aria-label={isMenuOpen ? "Закрыть меню" : "Открыть меню"}
              aria-expanded={isMenuOpen}
              onClick={() => setIsMenuOpen((value) => !value)}
              className="px-2"
            >
              <span className="flex flex-col gap-1">
                <span className="block h-0.5 w-4 rounded-full bg-current" />
                <span className="block h-0.5 w-4 rounded-full bg-current" />
                <span className="block h-0.5 w-4 rounded-full bg-current" />
              </span>
            </Button>
          </div>
        </div>

        {isMenuOpen && (
          <div className="border-t border-border md:hidden">
            <div className="mx-auto flex max-w-6xl flex-col gap-3 px-4 py-4 sm:px-6">
              <nav className="flex flex-col gap-1 text-sm">
                <MobileNavItem
                  to="/assistants"
                  label="Каталог"
                  end
                  onClick={closeMenu}
                />
                <MobileNavItem
                  to="/runs/my"
                  label="Мои запуски"
                  onClick={closeMenu}
                />
                {user?.role === "admin" && (
                  <MobileNavItem
                    to="/admin/runs"
                    label="Все запуски"
                    onClick={closeMenu}
                  />
                )}
              </nav>

              <div className="flex flex-col gap-3 border-t border-border pt-3">
                {user && (
                  <span className="break-all text-sm text-muted-foreground">
                    Вы вошли как: {user.email}
                  </span>
                )}

                <Button
                  variant="outline"
                  size="sm"
                  onClick={handleLogout}
                  className="w-full justify-center"
                >
                  Выйти
                </Button>
              </div>
            </div>
          </div>
        )}
      </header>

      <main className="mx-auto max-w-6xl px-4 py-6 sm:px-6 sm:py-8">
        {children ?? <Outlet />}
      </main>
    </div>
  );
}

function NavItem({
                   to,
                   label,
                   end,
                 }: {
  to: string;
  label: string;
  end?: boolean;
}) {
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

function MobileNavItem({
                         to,
                         label,
                         end,
                         onClick,
                       }: {
  to: string;
  label: string;
  end?: boolean;
  onClick?: () => void;
}) {
  return (
    <NavLink
      to={to}
      end={end}
      onClick={onClick}
      className={({ isActive }) =>
        [
          "rounded-md px-3 py-2 transition-colors",
          isActive
            ? "bg-accent text-accent-foreground"
            : "text-muted-foreground hover:bg-accent/50 hover:text-foreground",
        ].join(" ")
      }
    >
      {label}
    </NavLink>
  );
}
