import { RouterProvider, createRouter, createRootRoute, createRoute, Outlet } from "@tanstack/react-router";
import { QueryClientProvider } from "@tanstack/react-query";
import { queryClient } from "@/lib/queryClient";
import { useAuth, initAuth } from "@/lib/auth";
import { useEffect, useState } from "react";
import { DashboardLayout } from "@/components/layout/DashboardLayout";
import { QuotaUsageCard } from "@/components/dashboard/QuotaUsageCard";
import { VMCountCard } from "@/components/dashboard/VMCountCard";
import { ActivityLog } from "@/components/dashboard/ActivityLog";
import api from "@/lib/api";
import "../styles.css";

// Root route
const rootRoute = createRootRoute({
  component: RootComponent,
});

function RootComponent() {
  const { refreshUser, isAuthenticated, isLoading } = useAuth();

  useEffect(() => {
    initAuth();
  }, []);

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-900">
        <div className="animate-spin rounded-full h-16 w-16 border-b-2 border-primary"></div>
      </div>
    );
  }

  return <Outlet />;
}

// Index route (landing/login)
const indexRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/",
  component: IndexComponent,
});

function IndexComponent() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-blue-50 to-indigo-100 dark:from-gray-900 dark:to-gray-800">
      <div className="text-center p-8">
        <h1 className="text-5xl font-bold text-gray-900 dark:text-white mb-4">
          Podland
        </h1>
        <p className="text-xl text-gray-600 dark:text-gray-400 mb-8">
          Student PaaS Platform — Deploy apps with zero DevOps
        </p>
        <a
          href="/api/auth/login"
          className="inline-flex items-center gap-3 px-8 py-4 bg-gray-900 dark:bg-gray-100 text-white dark:text-gray-900 rounded-xl font-semibold hover:bg-gray-800 dark:hover:bg-gray-200 transition-all shadow-lg hover:shadow-xl"
        >
          <svg className="w-6 h-6" viewBox="0 0 24 24" fill="currentColor">
            <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/>
          </svg>
          Sign in with GitHub
        </a>
        <p className="mt-6 text-sm text-gray-500 dark:text-gray-400">
          Requires @student.unand.ac.id email
        </p>
      </div>
    </div>
  );
}

// Auth callback route
const authCallbackRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/auth/callback",
  component: AuthCallback,
});

function AuthCallback() {
  const { refreshUser } = useAuth();
  
  useEffect(() => {
    refreshUser().then(() => {
      window.location.href = "/dashboard";
    });
  }, [refreshUser]);

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-900">
      <div className="text-center">
        <div className="animate-spin rounded-full h-16 w-16 border-b-2 border-primary mx-auto mb-4"></div>
        <p className="text-lg text-gray-600 dark:text-gray-400">
          Completing sign in...
        </p>
      </div>
    </div>
  );
}

// Auth welcome route
const authWelcomeRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/auth/welcome",
  component: AuthWelcome,
});

function AuthWelcome() {
  const { setUser } = useAuth();
  const [nim, setNim] = useState("");
  const [extractedNim, setExtractedNim] = useState("");
  const [role, setRole] = useState("");
  const [displayName, setDisplayName] = useState("");
  const [termsAccepted, setTermsAccepted] = useState(false);
  const [isEditing, setIsEditing] = useState(false);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const userId = params.get("userId");
    if (userId) {
      api
        .get(`/users/${userId}`)
        .then((res) => {
          setExtractedNim(res.data.nim);
          setNim(res.data.nim);
          setRole(res.data.role);
          setDisplayName(res.data.displayName);
          setIsLoading(false);
        })
        .catch(() => {
          window.location.href = "/";
        });
    } else {
      window.location.href = "/";
    }
  }, []);

  const handleConfirm = async () => {
    try {
      await api.post("/users/confirm-nim", { nim });
      const { data } = await api.get("/users/me");
      setUser(data);
      window.location.href = "/dashboard";
    } catch (error) {
      console.error("Failed to confirm NIM:", error);
    }
  };

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-900">
        <div className="animate-spin rounded-full h-16 w-16 border-b-2 border-primary"></div>
      </div>
    );
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-900">
      <div className="max-w-md w-full p-6 bg-white dark:bg-gray-800 rounded-lg shadow-lg">
        <h1 className="text-2xl font-bold text-gray-900 dark:text-white mb-4">
          Welcome to Podland!
        </h1>

        <div className="space-y-4">
          <div className="p-4 bg-gray-100 dark:bg-gray-700 rounded-lg">
            <p className="text-lg font-medium text-gray-900 dark:text-white">
              {displayName}
            </p>
          </div>

          {!isEditing ? (
            <div className="p-4 bg-gray-100 dark:bg-gray-700 rounded-lg">
              <p className="text-sm font-medium text-gray-600 dark:text-gray-400 mb-1">
                Your NIM:
              </p>
              <p className="text-2xl font-bold text-gray-900 dark:text-white">
                {extractedNim}
              </p>
              <p className="text-sm text-gray-600 dark:text-gray-400 mt-2">
                Role:{" "}
                <span className="font-semibold">
                  {role === "internal" ? "Internal (SI UNAND)" : "External"}
                </span>
              </p>
            </div>
          ) : (
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Enter your NIM
              </label>
              <input
                type="text"
                value={nim}
                onChange={(e) => setNim(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-primary focus:border-transparent"
                placeholder="221152001"
                maxLength={9}
              />
            </div>
          )}

          <div className="flex gap-2">
            <button
              onClick={() => setIsEditing(!isEditing)}
              className="text-sm text-primary hover:underline font-medium"
            >
              {isEditing ? "Cancel" : "Edit NIM"}
            </button>
          </div>

          <div className="flex items-start gap-2">
            <input
              type="checkbox"
              id="terms"
              checked={termsAccepted}
              onChange={(e) => setTermsAccepted(e.target.checked)}
              className="mt-1 rounded"
            />
            <label htmlFor="terms" className="text-sm text-gray-600 dark:text-gray-400">
              I agree to the{" "}
              <a href="/terms" target="_blank" rel="noopener noreferrer" className="text-primary hover:underline">
                Terms of Service
              </a>{" "}
              and confirm I am a current Unand student.
            </label>
          </div>

          <button
            onClick={handleConfirm}
            disabled={!termsAccepted}
            className="w-full py-3 px-4 bg-primary text-white rounded-lg hover:bg-primary-dark disabled:opacity-50 disabled:cursor-not-allowed transition-colors font-semibold"
          >
            Activate Account
          </button>
        </div>
      </div>
    </div>
  );
}

// Auth rejected route
const authRejectedRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/auth/rejected",
  component: AuthRejected,
});

function AuthRejected() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-900">
      <div className="max-w-md w-full p-6 bg-white dark:bg-gray-800 rounded-lg shadow-lg text-center">
        <div className="w-16 h-16 mx-auto mb-4 bg-red-100 dark:bg-red-900/20 rounded-full flex items-center justify-center">
          <svg className="w-8 h-8 text-red-600 dark:text-red-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
          </svg>
        </div>
        <h1 className="text-2xl font-bold text-gray-900 dark:text-white mb-2">Access Denied</h1>
        <p className="text-gray-600 dark:text-gray-400 mb-4">
          Your GitHub email must end with{" "}
          <code className="px-2 py-1 bg-gray-100 dark:bg-gray-700 rounded text-sm">@student.unand.ac.id</code>
        </p>
        <div className="text-left p-4 bg-gray-50 dark:bg-gray-700/50 rounded-lg mb-6">
          <h2 className="font-semibold text-gray-900 dark:text-white mb-2">How to fix this:</h2>
          <ol className="list-decimal list-inside space-y-2 text-sm text-gray-600 dark:text-gray-400">
            <li>Go to your <a href="https://github.com/settings/emails" target="_blank" rel="noopener noreferrer" className="text-primary hover:underline">GitHub Email Settings</a></li>
            <li>Add your <code className="px-1 bg-gray-100 dark:bg-gray-600 rounded">@student.unand.ac.id</code> email</li>
            <li>Verify the email address</li>
            <li>Set it as your primary email</li>
            <li>Click "Retry" below</li>
          </ol>
        </div>
        <a href="/api/auth/login" className="inline-block px-6 py-3 bg-primary text-white rounded-lg hover:bg-primary-dark transition-colors font-semibold">Retry with Updated Email</a>
      </div>
    </div>
  );
}

// Dashboard route
const dashboardRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/dashboard",
  component: DashboardWrapper,
});

function DashboardWrapper() {
  return (
    <DashboardLayout>
      <DashboardHome />
    </DashboardLayout>
  );
}

function DashboardHome() {
  const { user } = useAuth();
  const { data: activities } = useAuthQuery();

  const quota = user?.role === "internal" ? { cpu: 1, ram: 2048 } : { cpu: 0.5, ram: 1024 };

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900 dark:text-white">Dashboard</h1>
        <p className="text-gray-600 dark:text-gray-400 mt-1">Welcome back, {user?.displayName}</p>
      </div>
      <QuotaUsageCard usedCpu={0} maxCpu={quota.cpu} usedRam={0} maxRam={quota.ram} />
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <VMCountCard count={0} />
        <ActivityLog activities={activities} className="lg:col-span-2" />
      </div>
      <div className="p-6 bg-gradient-to-r from-blue-50 to-indigo-50 dark:from-gray-800 dark:to-gray-700 rounded-lg border border-blue-100 dark:border-gray-600">
        <h2 className="text-lg font-semibold text-gray-900 dark:text-white mb-2">🚀 Coming Soon: VM Management</h2>
        <p className="text-gray-600 dark:text-gray-400">Create and manage your application VMs with automatic resource allocation.</p>
      </div>
    </div>
  );
}

function useAuthQuery() {
  const [data, setData] = useState(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    api.get("/activity").then((res) => {
      setData(res.data);
      setIsLoading(false);
    }).catch(() => {
      setIsLoading(false);
    });
  }, []);

  return { data, isLoading };
}

// Dashboard profile route
const dashboardProfileRoute = createRoute({
  getParentRoute: () => dashboardRoute,
  path: "/profile",
  component: DashboardProfile,
});

function DashboardProfile() {
  const { user } = useAuth();

  if (!user) return null;

  return (
    <div className="max-w-2xl">
      <h1 className="text-2xl font-bold text-gray-900 dark:text-white mb-6">Profile</h1>
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
        <div className="p-6 border-b border-gray-200 dark:border-gray-700">
          <div className="flex items-center gap-4">
            <img src={user.avatarUrl || `https://avatars.githubusercontent.com/${user.githubId}`} alt={user.displayName} className="w-20 h-20 rounded-full" />
            <div>
              <h2 className="text-xl font-semibold text-gray-900 dark:text-white">{user.displayName}</h2>
              <p className="text-sm text-gray-500 dark:text-gray-400">{user.email}</p>
            </div>
          </div>
        </div>
        <div className="p-6 space-y-4">
          <ProfileField label="Display Name" value={user.displayName} />
          <ProfileField label="Email" value={user.email} readOnly />
          <ProfileField label="NIM" value={user.nim} readOnly />
          <ProfileField label="Role" value={user.role === "internal" ? "Internal (SI UNAND)" : user.role === "superadmin" ? "Super Admin" : "External"} readOnly />
          <ProfileField label="Member Since" value={new Date(user.createdAt).toLocaleDateString("en-US", { year: "numeric", month: "long", day: "numeric" })} readOnly />
        </div>
      </div>
    </div>
  );
}

function ProfileField({ label, value, readOnly = true }: { label: string; value: string; readOnly?: boolean }) {
  return (
    <div>
      <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{label}</label>
      <input type="text" value={value || ""} readOnly={readOnly} className={`w-full px-3 py-2 border rounded-md ${readOnly ? "bg-gray-50 dark:bg-gray-700 text-gray-500 dark:text-gray-400" : "bg-white dark:bg-gray-800 text-gray-900 dark:text-white"}`} />
    </div>
  );
}

const routeTree = rootRoute.addChildren([
  indexRoute,
  authCallbackRoute,
  authWelcomeRoute,
  authRejectedRoute,
  dashboardRoute.addChildren([dashboardProfileRoute]),
]);

const router = createRouter({ routeTree });

declare module "@tanstack/react-router" {
  interface Register {
    router: typeof router;
  }
}

export function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>
  );
}
