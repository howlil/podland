import { createRootRoute, Outlet } from "@tanstack/react-router";
import { QueryClientProvider } from "@tanstack/react-query";
import { queryClient } from "@/lib/queryClient";
import { initAuth } from "@/lib/auth";
import { useEffect, useState } from "react";
import { Toaster } from "sonner";
import { ErrorBoundary } from "@/lib/ErrorBoundary";
import { DashboardLayout } from "@/components/layout/DashboardLayout";

// Initialize auth on mount
function AuthInitializer() {
  useEffect(() => {
    initAuth();
  }, []);
  return null;
}

// Root component with providers
function RootComponent() {
  const [isInitialized, setIsInitialized] = useState(false);

  useEffect(() => {
    setIsInitialized(true);
  }, []);

  if (!isInitialized) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-900">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary" role="status" aria-label="Loading">
          <span className="sr-only">Loading...</span>
        </div>
      </div>
    );
  }

  return (
    <ErrorBoundary>
      <QueryClientProvider client={queryClient}>
        <AuthInitializer />
        <DashboardLayout>
          <Outlet />
        </DashboardLayout>
        <Toaster
          position="top-right"
          toastOptions={{
            duration: 4000,
            classNames: {
              toast: "bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 shadow-lg",
              title: "text-gray-900 dark:text-white",
              description: "text-gray-600 dark:text-gray-400",
              success: "border-green-500",
              error: "border-red-500",
              warning: "border-yellow-500",
              info: "border-blue-500",
            },
          }}
        />
      </QueryClientProvider>
    </ErrorBoundary>
  );
}

export const Route = createRootRoute({
  component: RootComponent,
});
