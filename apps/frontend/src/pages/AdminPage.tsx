import { createFileRoute } from '@tanstack/react-router'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'

export const Route = createFileRoute("/admin")({
  component: AdminPage,
})

export default function AdminPage() {
  return (
    <div className="container mx-auto p-6">
      <h1 className="text-3xl font-bold mb-6">Admin Panel</h1>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        <Card>
          <CardHeader>
            <CardTitle>User Management</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-gray-600 mb-4">
              View and manage all users, change roles, ban/unban
            </p>
            <Button onClick={() => window.location.href = '/admin/users'}>
              Manage Users
            </Button>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>System Health</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-gray-600 mb-4">
              Monitor cluster CPU, memory, and storage usage
            </p>
            <Button variant="outline" onClick={() => window.location.href = '/admin/health'}>
              View Health
            </Button>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Audit Log</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-gray-600 mb-4">
              View all admin actions with timestamps
            </p>
            <Button variant="outline" onClick={() => window.location.href = '/admin/audit-log'}>
              View Logs
            </Button>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
