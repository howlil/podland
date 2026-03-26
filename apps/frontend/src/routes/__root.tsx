import { createRootRoute, Outlet } from "@tanstack/react-router";

export const Route = createRootRoute({
  component: RootComponent,
});

function RootComponent() {
  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
      <div className="container mx-auto px-4 py-8">
        <h1 className="text-3xl font-bold text-gray-900 dark:text-white">
          Podland
        </h1>
        <p className="mt-2 text-gray-600 dark:text-gray-400">
          Student PaaS Platform - Development Mode
        </p>
        <div className="mt-8 p-4 bg-white dark:bg-gray-800 rounded-lg shadow">
          <h2 className="text-xl font-semibold text-gray-900 dark:text-white">
            Getting Started
          </h2>
          <ul className="mt-4 space-y-2 text-gray-600 dark:text-gray-400">
            <li>✅ Database: PostgreSQL running on port 5432</li>
            <li>✅ Backend: API running on http://localhost:8080</li>
            <li>✅ Frontend: Running on http://localhost:3000</li>
          </ul>
        </div>
      </div>
      <Outlet />
    </div>
  );
}
