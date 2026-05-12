import { lazy, Suspense } from "react";
import { BrowserRouter, Navigate, Route, Routes } from "react-router-dom";
import { AppLayout } from "@/components/AppLayout";
import { ProtectedRoute } from "@/components/ProtectedRoute";
import { useAuth } from "@/lib/auth";
import { AdminRunsPage } from "@/pages/AdminRunsPage";
import { AssistantDetailPage } from "@/pages/AssistantDetailPage";
import { CatalogPage } from "@/pages/CatalogPage";
import { LoginPage } from "@/pages/LoginPage";
import { MyRunsPage } from "@/pages/MyRunsPage";
import { NotFoundPage } from "@/pages/NotFoundPage";
import { EditAssistantPage } from "@/pages/admin/EditAssistantPage";
import { NewAssistantPage } from "@/pages/admin/NewAssistantPage";
import { NewCategoryPage } from "@/pages/admin/NewCategoryPage";

function NotFoundRoute() {
  const { isAuthenticated } = useAuth();
  if (isAuthenticated) {
    return (
      <AppLayout>
        <NotFoundPage />
      </AppLayout>
    );
  }
  return <NotFoundPage fullScreen />;
}

const UiKitPage = import.meta.env.DEV
  ? lazy(() => import("@/pages/dev/UiKitPage.tsx").then((m) => ({ default: m.UiKitPage })))
  : null;

export function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route path="/404" element={<NotFoundPage fullScreen />} />
        {UiKitPage && (
          <Route
            path="/dev/ui"
            element={
              <Suspense fallback={null}>
                <UiKitPage />
              </Suspense>
            }
          />
        )}

        <Route
          element={
            <ProtectedRoute>
              <AppLayout />
            </ProtectedRoute>
          }
        >
          <Route index element={<Navigate to="/assistants" replace />} />
          <Route path="assistants" element={<CatalogPage />} />
          <Route path="assistants/:id" element={<AssistantDetailPage />} />
          <Route path="runs/my" element={<MyRunsPage />} />
          <Route
            path="admin/runs"
            element={
              <ProtectedRoute role="admin">
                <AdminRunsPage />
              </ProtectedRoute>
            }
          />
          <Route
            path="admin/categories/new"
            element={
              <ProtectedRoute role="admin">
                <NewCategoryPage />
              </ProtectedRoute>
            }
          />
          <Route
            path="admin/assistants/new"
            element={
              <ProtectedRoute role="admin">
                <NewAssistantPage />
              </ProtectedRoute>
            }
          />
          <Route
            path="admin/assistants/:id/edit"
            element={
              <ProtectedRoute role="admin">
                <EditAssistantPage />
              </ProtectedRoute>
            }
          />
        </Route>

        <Route path="*" element={<NotFoundRoute />} />
      </Routes>
    </BrowserRouter>
  );
}
