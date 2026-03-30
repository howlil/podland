import { useState } from "react";
import { useAdminUsers } from "@/hooks/useAdminUsers";
import { DashboardLayout } from "@/components/layout/DashboardLayout";
import { UserTable } from "@/components/admin/UserTable";

export default function AdminUsersPage() {
  const [roleFilter, setRoleFilter] = useState<string>("all");

  const {
    users,
    isLoading,
    changeRole,
    banUser,
    unbanUser,
    isChangingRole,
    isBanning,
    isUnbanning,
  } = useAdminUsers(roleFilter === "all" ? undefined : roleFilter);

  const handleChangeRole = (userId: string, role: string) => {
    changeRole({ userId, role });
  };

  return (
    <DashboardLayout>
      <div className="max-w-7xl mx-auto">
        {/* Header */}
        <div className="flex justify-between items-center mb-6">
          <div>
            <h1 className="text-2xl font-bold text-gray-900 dark:text-white">User Management</h1>
            <p className="text-gray-600 dark:text-gray-400 mt-1">Manage users, roles, and permissions</p>
          </div>
        </div>

        {/* Filters */}
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-4 mb-6">
          <div className="flex flex-wrap gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Filter by Role
              </label>
              <select
                value={roleFilter}
                onChange={(e) => setRoleFilter(e.target.value)}
                className="px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-primary focus:border-transparent"
              >
                <option value="all">All Roles</option>
                <option value="external">External</option>
                <option value="internal">Internal</option>
                <option value="superadmin">Superadmin</option>
              </select>
            </div>
          </div>
        </div>

        {/* User Table */}
        <UserTable
          users={users}
          isLoading={isLoading}
          onChangeRole={handleChangeRole}
          onBan={banUser}
          onUnban={unbanUser}
          isChangingRole={isChangingRole}
          isBanning={isBanning}
          isUnbanning={isUnbanning}
        />
      </div>
    </DashboardLayout>
  );
}
