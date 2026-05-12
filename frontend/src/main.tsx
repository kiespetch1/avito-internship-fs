import { QueryClientProvider } from "@tanstack/react-query";
import { ThemeProvider } from "next-themes";
import { lazy, StrictMode, Suspense } from "react";
import { createRoot } from "react-dom/client";

import { App } from "./App";
import { Toaster } from "@/components/ui/sonner";
import { AuthProvider } from "@/lib/auth";
import { createQueryClient } from "@/lib/query";
import "./index.css";

const ReactQueryDevtools = import.meta.env.DEV
  ? lazy(() =>
      import("@tanstack/react-query-devtools").then((m) => ({ default: m.ReactQueryDevtools })),
    )
  : null;

const queryClient = createQueryClient();

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <ThemeProvider
      attribute="class"
      defaultTheme="system"
      enableSystem
      disableTransitionOnChange={false}
      storageKey="theme"
    >
      <QueryClientProvider client={queryClient}>
        <AuthProvider>
          <App />
          <Toaster richColors position="top-right" />
        </AuthProvider>
        {ReactQueryDevtools && (
          <Suspense fallback={null}>
            <ReactQueryDevtools initialIsOpen={false} />
          </Suspense>
        )}
      </QueryClientProvider>
    </ThemeProvider>
  </StrictMode>,
);
