import { CheckIcon, MonitorIcon, MoonIcon, SunIcon } from "lucide-react";
import { useTheme } from "next-themes";
import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";

type ThemeOption = "light" | "dark" | "system";

const OPTIONS: { value: ThemeOption; label: string; icon: typeof SunIcon }[] = [
  { value: "light", label: "Светлая", icon: SunIcon },
  { value: "dark", label: "Тёмная", icon: MoonIcon },
  { value: "system", label: "Системная", icon: MonitorIcon },
];

export function ThemeToggle() {
  const { theme, setTheme, resolvedTheme } = useTheme();
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
  }, []);

  const active = (theme ?? "system") as ThemeOption;
  const showDarkIcon = mounted && resolvedTheme === "dark";

  return (
    <DropdownMenu>
      <DropdownMenuTrigger
        render={
          <Button
            variant="outline"
            size="icon-sm"
            aria-label="Переключить тему"
            className="relative"
          >
            <SunIcon
              className={`size-4 transition-all duration-200 ${
                showDarkIcon
                  ? "-rotate-90 scale-0 opacity-0"
                  : "rotate-0 scale-100 opacity-100"
              }`}
            />
            <MoonIcon
              className={`absolute size-4 transition-all duration-200 ${
                showDarkIcon
                  ? "rotate-0 scale-100 opacity-100"
                  : "rotate-90 scale-0 opacity-0"
              }`}
            />
          </Button>
        }
      />
      <DropdownMenuContent align="end" sideOffset={8} className="min-w-40">
        {OPTIONS.map(({ value, label, icon: Icon }) => {
          const selected = active === value;
          return (
            <DropdownMenuItem
              key={value}
              onClick={() => setTheme(value)}
              className="justify-between"
            >
              <span className="flex items-center gap-2">
                <Icon className="size-4 text-muted-foreground" />
                {label}
              </span>
              {selected && <CheckIcon className="size-4 text-primary" />}
            </DropdownMenuItem>
          );
        })}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
