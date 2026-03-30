import { createFileRoute } from '@tanstack/react-router'
import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'

export const Route = createFileRoute('/admin/users')({
  component: AdminUsersPage,
})

interface User {
  id: string
  email: string
  display_name: string
  role: string
  nim?: string
  created_at: string
}

function AdminUsersPage() {
  const [roleFilter, setRoleFilter] = useState('')
  const queryClient = useQueryClient()

  const { data: users, isLoading } = useQuery({
    queryKey: ['admin-users', roleFilter],
    queryFn: async () => {
      const response = await fetch(`/api/admin/users${roleFilter ? `?role=${roleFilter}` : ''}`)
      if (!response.ok) throw new Error('Failed to fetch users')
      return response.json()
    },
  })

  const changeRoleMutation = useMutation({
    mutationFn: async ({ userID, role }: { userID: string; role: string }) => {
      const response = await fetch(`/api/admin/users/${userID}/role`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ role }),
      })
      if (!response.ok) throw new Error('Failed to update role')
      return response.json()
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-users'] })
    },
  })

  const banUserMutation = useMutation({
    mutationFn: async (userID: string) => {
      const response = await fetch(`/api/admin/users/${userID}/ban`, {
        method: 'POST',
      })
      if (!response.ok) throw new Error('Failed to ban user')
      return response.json()
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-users'] })
    },
  })

  const unbanUserMutation = useMutation({
    mutationFn: async (userID: string) => {
      const response = await fetch(`/api/admin/users/${userID}/unban`, {
        method: 'POST',
      })
      if (!response.ok) throw new Error('Failed to unban user')
      return response.json()
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-users'] })
    },
  })

  const getRoleBadgeColor = (role: string) => {
    switch (role) {
      case 'superadmin':
        return 'bg-red-100 text-red-800'
      case 'internal':
        return 'bg-blue-100 text-blue-800'
      case 'external':
        return 'bg-gray-100 text-gray-800'
      case 'banned':
        return 'bg-red-800 text-white'
      default:
        return 'bg-gray-100 text-gray-800'
    }
  }

  return (
    <div className="container mx-auto p-6">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-3xl font-bold">User Management</h1>

        <select
          value={roleFilter}
          onChange={(e) => setRoleFilter(e.target.value)}
          className="border rounded px-3 py-2 bg-white"
        >
          <option value="">All Roles</option>
          <option value="internal">Internal</option>
          <option value="external">External</option>
          <option value="superadmin">Superadmin</option>
        </select>
      </div>

      {isLoading ? (
        <div>Loading...</div>
      ) : (
        <Card>
          <CardContent className="p-0">
            <table className="w-full">
              <thead className="bg-gray-50 border-b">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Email</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Name</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">NIM</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Role</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Actions</th>
                </tr>
              </thead>
              <tbody>
                {users?.data?.map((user: User) => (
                  <tr key={user.id} className="border-b">
                    <td className="px-6 py-4">{user.email}</td>
                    <td className="px-6 py-4">{user.display_name}</td>
                    <td className="px-6 py-4">{user.nim || '-'}</td>
                    <td className="px-6 py-4">
                      <Badge className={getRoleBadgeColor(user.role)}>{user.role}</Badge>
                    </td>
                    <td className="px-6 py-4">
                      <div className="flex gap-2">
                        <select
                          value={user.role === 'banned' ? 'external' : user.role}
                          onChange={(e) => changeRoleMutation.mutate({ userID: user.id, role: e.target.value })}
                          className="border rounded px-2 py-1 text-sm"
                          disabled={changeRoleMutation.isPending}
                        >
                          <option value="internal">Internal</option>
                          <option value="external">External</option>
                          <option value="superadmin">Superadmin</option>
                        </select>
                        {user.role === 'banned' ? (
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={() => unbanUserMutation.mutate(user.id)}
                            disabled={unbanUserMutation.isPending}
                          >
                            Unban
                          </Button>
                        ) : (
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={() => banUserMutation.mutate(user.id)}
                            disabled={banUserMutation.isPending}
                          >
                            Ban
                          </Button>
                        )}
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </CardContent>
        </Card>
      )}
    </div>
  )
}
