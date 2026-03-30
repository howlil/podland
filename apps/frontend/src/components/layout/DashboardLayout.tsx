import { useAuth } from "@/lib/auth";
import { useState } from "react";
import { NotificationBell } from "@/components/notifications/NotificationBell";
import { LayoutDashboard, Server, User, ChevronDown, LogOut } from "lucide-react";

interface DashboardLayoutProps {
  children: React.ReactNode;
}

export function DashboardLayout({ children }: DashboardLayoutProps) {
  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
      {/* Desktop Sidebar */}
      <aside className="hidden md:fixed md:inset-y-0 md:flex md:w-64 md:flex-col">
        <div className="flex flex-col flex-grow pt-5 overflow-y-auto bg-white dark:bg-gray-800 border-r border-gray-200 dark:border-gray-700">
          <div className="flex items-center flex-shrink-0 px-4 mb-5">
            <h1 className="text-xl font-bold text-primary">Podland</h1>
          </div>

          <nav className="flex-1 px-2 space-y-1" role="navigation" aria-label="Main navigation">
            <a
              href="/dashboard"
              className="group flex items-center px-2 py-2 text-sm font-medium rounded-md text-gray-900 dark:text-white bg-gray-100 dark:bg-gray-700"
              aria-current="page"
            >
              <LayoutDashboard className="mr-3 h-5 w-5" aria-hidden="true" />
              Dashboard
            </a>
            <a
              href="/dashboard/-vms"
              className="group flex items-center px-2 py-2 text-sm font-medium rounded-md text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700"
            >
              <Server className="mr-3 h-5 w-5" aria-hidden="true" />
              VMs
            </a>
            <a
              href="/dashboard/profile"
              className="group flex items-center px-2 py-2 text-sm font-medium rounded-md text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700"
            >
              <User className="mr-3 h-5 w-5" aria-hidden="true" />
              Profile
            </a>
          </nav>

          {/* User Avatar Dropdown */}
          <div className="flex items-center p-4 border-t border-gray-200 dark:border-gray-700">
            <UserDropdown />
          </div>
        </div>
      </aside>

      {/* Mobile Bottom Tab Bar */}
      <nav className="md:hidden fixed bottom-0 inset-x-0 bg-white dark:bg-gray-800 border-t border-gray-200 dark:border-gray-700 z-50" role="navigation" aria-label="Mobile navigation">
        <div className="grid grid-cols-3 h-16">
          <a
            href="/dashboard"
            className="flex flex-col items-center justify-center text-gray-900 dark:text-white"
            aria-label="Go to Dashboard"
          >
            <LayoutDashboard className="h-6 w-6" aria-hidden="true" />
            <span className="text-xs mt-1">Dashboard</span>
          </a>
          <a
            href="/dashboard/-vms"
            className="flex flex-col items-center justify-center text-gray-700 dark:text-gray-300"
            aria-label="Go to VMs"
          >
            <Server className="h-6 w-6" aria-hidden="true" />
            <span className="text-xs mt-1">VMs</span>
          </a>
          <a
            href="/dashboard/profile"
            className="flex flex-col items-center justify-center text-gray-700 dark:text-gray-300"
            aria-label="Go to Profile"
          >
            <User className="h-6 w-6" aria-hidden="true" />
            <span className="text-xs mt-1">Profile</span>
          </a>
        </div>
      </nav>

      {/* Main Content */}
      <main className="md:pl-64 pb-16 md:pb-0">
        {/* Top Header Bar */}
        <header className="bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700 px-4 py-3 sm:px-6">
          <div className="flex items-center justify-between">
            <div className="md:hidden">
              <h1 className="text-lg font-bold text-primary">Podland</h1>
            </div>
            <div className="flex-1"></div>
            <div className="flex items-center gap-4">
              <NotificationBell />
              <UserDropdown />
            </div>
          </div>
        </header>

        <div className="px-4 py-6 sm:px-6">
          {children}
        </div>
      </main>
    </div>
  );
}

function UserDropdown() {
  const { user, logout } = useAuth();
  const [isOpen, setIsOpen] = useState(false);

  if (!user) return null;

  return (
    <div className="relative">
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="flex items-center gap-2 focus:outline-none focus:ring-2 focus:ring-primary focus:ring-offset-2 dark:focus:ring-offset-gray-800 rounded-md"
        aria-expanded={isOpen}
        aria-haspopup="menu"
        aria-label={`User menu for ${user.displayName}`}
      >
        <img
          src={user.avatarUrl || `https://avatars.githubusercontent.com/${user.githubId}`}
          alt={user.displayName}
          className="w-8 h-8 rounded-full"
        />
        <span className="text-sm font-medium text-gray-700 dark:text-gray-300 hidden lg:block">
          {user.displayName}
        </span>
        <ChevronDown className="w-4 h-4 text-gray-500" aria-hidden="true" />
      </button>

      {isOpen && (
        <div
          role="menu"
          aria-orientation="vertical"
          aria-labelledby="user-menu"
          className="absolute bottom-full left-0 mb-2 w-48 bg-white dark:bg-gray-800 rounded-md shadow-lg py-1 border border-gray-200 dark:border-gray-700 z-50"
        >
          <div className="px-4 py-2 border-b border-gray-200 dark:border-gray-700" id="user-menu">
            <p className="text-sm font-medium text-gray-900 dark:text-white">
              {user.displayName}
            </p>
            <p className="text-xs text-gray-500 dark:text-gray-400 truncate">
              {user.email}
            </p>
          </div>
          <button
            onClick={() => {
              logout();
              setIsOpen(false);
            }}
            role="menuitem"
            className="w-full text-left px-4 py-2 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 flex items-center gap-2"
          >
            <LogOut className="w-4 h-4" aria-hidden="true" />
            Sign out
          </button>
        </div>
      )}
    </div>
  );
}
