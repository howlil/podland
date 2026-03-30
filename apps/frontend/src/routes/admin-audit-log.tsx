import { createFileRoute } from '@tanstack/react-router'
import { useQuery } from '@tanstack/react-query'
import { Card, CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'

export const Route = createFileRoute('/admin-audit-log')({
  component: AuditLogPage,
})

interface AuditLog {
  id: string
  user_id: string
  action: string
  ip_address: string
  user_agent: string
  created_at: string
}

function AuditLogPage() {
  const { data: logs, isLoading } = useQuery({
    queryKey: ['admin-audit-log'],
    queryFn: async () => {
      const response = await fetch('/api/admin/audit-log?limit=100')
      if (!response.ok) throw new Error('Failed to fetch audit logs')
      return response.json()
    },
    refetchInterval: 60000, // Refresh every minute
  })

  if (isLoading) return <div>Loading...</div>

  return (
    <div className="container mx-auto p-6">
      <h1 className="text-3xl font-bold mb-6">Audit Log</h1>

      <Card>
        <CardContent className="p-0">
          <table className="w-full">
            <thead className="bg-gray-50 border-b">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Timestamp</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">User ID</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Action</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">IP Address</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">User Agent</th>
              </tr>
            </thead>
            <tbody>
              {logs?.data?.map((log: AuditLog) => (
                <tr key={log.id} className="border-b">
                  <td className="px-6 py-4">
                    <Badge variant="outline">{new Date(log.created_at).toLocaleString()}</Badge>
                  </td>
                  <td className="px-6 py-4 font-mono text-sm">{log.user_id}</td>
                  <td className="px-6 py-4 font-mono text-sm">{log.action}</td>
                  <td className="px-6 py-4 font-mono text-sm">{log.ip_address}</td>
                  <td className="px-6 py-4 text-sm text-gray-500 truncate max-w-xs">{log.user_agent}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </CardContent>
      </Card>
    </div>
  )
}
