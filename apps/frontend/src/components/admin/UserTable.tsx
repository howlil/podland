import { User } from "@/hooks/useAdminUsers";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";

interface UserTableProps {
  users?: User[];
  isLoading?: boolean;
  onChangeRole?: (userId: string, role: string) => void;
  onBan?: (userId: string) => void;
  onUnban?: (userId: string) => void;
  isChangingRole?: boolean;
  isBanning?: boolean;
  isUnbanning?: boolean;
}

export function UserTable({ users, isLoading, onChangeRole, onBan, onUnban, isChangingRole, isBanning, isUnbanning }: UserTableProps) {
  if (isLoading) {
    return (
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden animate-pulse">
        <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
          <thead className="bg-gray-50 dark:bg-gray-700">
            <tr>
              {["User", "Email", "Role", "NIM", "Created", "Actions"].map((h, i) => (
                <th key={i} className="px-6 py-3">
                  <div className="h-4 w-20 bg-gray-200 dark:bg-gray-600 rounded" />
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {Array.from({ length: 5 }).map((_, i) => (
              <tr key={i} className="border-t border-gray-200 dark:border-gray-700">
                <td className="px-6 py-4">
                  <div className="h-4 w-32 bg-gray-200 dark:bg-gray-600 rounded" />
                </td>
                <td className="px-6 py-4">
                  <div className="h-4 w-40 bg-gray-200 dark:bg-gray-600 rounded" />
                </td>
                <td className="px-6 py-4">
                  <div className="h-6 w-20 bg-gray-200 dark:bg-gray-600 rounded-full" />
                </td>
                <td className="px-6 py-4">
                  <div className="h-4 w-16 bg-gray-200 dark:bg-gray-600 rounded" />
                </td>
                <td className="px-6 py-4">
                  <div className="h-4 w-24 bg-gray-200 dark:bg-gray-600 rounded" />
                </td>
                <td className="px-6 py-4">
                  <div className="h-8 w-20 bg-gray-200 dark:bg-gray-600 rounded" />
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    );
  }

  if (!users || users.length === 0) {
    return (
      <div className="text-center py-12 text-gray-500 dark:text-gray-400">
        No users found
      </div>
    );
  }

  return (
    <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
      <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
        <thead className="bg-gray-50 dark:bg-gray-700">
          <tr>
            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">User</th>
            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">Email</th>
            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">Role</th>
            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">NIM</th>
            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">Created</th>
            <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">Actions</th>
          </tr>
        </thead>
        <tbody className="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
          {users.map((user) => (
            <tr key={user.id} className="hover:bg-gray-50 dark:hover:bg-gray-700/50">
              <td className="px-6 py-4 whitespace-nowrap">
                <div className="text-sm font-medium text-gray-900 dark:text-white">{user.display_name}</div>
              </td>
              <td className="px-6 py-4 whitespace-nowrap">
                <div className="text-sm text-gray-600 dark:text-gray-300">{user.email}</div>
              </td>
              <td className="px-6 py-4 whitespace-nowrap">
                <Badge variant={user.role === "superadmin" ? "destructive" : user.role === "internal" ? "default" : "secondary"}>
                  {user.role}
                </Badge>
              </td>
              <td className="px-6 py-4 whitespace-nowrap">
                <div className="text-sm text-gray-600 dark:text-gray-300">{user.nim || "-"}</div>
              </td>
              <td className="px-6 py-4 whitespace-nowrap">
                <div className="text-sm text-gray-600 dark:text-gray-300">{new Date(user.created_at).toLocaleDateString()}</div>
              </td>
              <td className="px-6 py-4 whitespace-nowrap text-right">
                <div className="flex justify-end gap-2">
                  {user.role !== "superadmin" && (
                    <>
                      <select
                        value={user.role}
                        onChange={(e) => onChangeRole?.(user.id, e.target.value)}
                        disabled={isChangingRole}
                        className="text-sm border border-gray-300 dark:border-gray-600 rounded px-2 py-1 bg-white dark:bg-gray-700 text-gray-900 dark:text-white disabled:opacity-50"
                      >
                        <option value="external">External</option>
                        <option value="internal">Internal</option>
                        <option value="superadmin">Superadmin</option>
                      </select>
                      {user.role !== "banned" ? (
                        <Button
                          size="sm"
                          variant="destructive"
                          onClick={() => onBan?.(user.id)}
                          disabled={isBanning}
                        >
                          Ban
                        </Button>
                      ) : (
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={() => onUnban?.(user.id)}
                          disabled={isUnbanning}
                        >
                          Unban
                        </Button>
                      )}
                    </>
                  )}
                </div>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
