import { createFileRoute } from '@tanstack/react-router'
import { useQuery } from '@tanstack/react-query'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Progress } from '@/components/ui/progress'
import { REFRESH_INTERVALS } from '@/lib/constants'

export const Route = createFileRoute('/admin-health')({
  component: AdminHealthPage,
})

export default function AdminHealthPage() {
  const { data: health, isLoading } = useQuery({
    queryKey: ['admin-health'],
    queryFn: async () => {
      const response = await fetch('/api/admin/health')
      if (!response.ok) throw new Error('Failed to fetch health')
      return response.json()
    },
    refetchInterval: REFRESH_INTERVALS.HEALTH,
  })

  if (isLoading) return <div>Loading...</div>

  return (
    <div className="container mx-auto p-6">
      <h1 className="text-3xl font-bold mb-6">System Health</h1>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        <Card>
          <CardHeader>
            <CardTitle>Cluster CPU</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-4xl font-bold">{health?.data?.cluster_cpu || 0}%</div>
            <Progress value={health?.data?.cluster_cpu || 0} className="mt-2" />
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Cluster Memory</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-4xl font-bold">{health?.data?.cluster_memory || 0}%</div>
            <Progress value={health?.data?.cluster_memory || 0} className="mt-2" />
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Cluster Storage</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-4xl font-bold">{health?.data?.cluster_storage || 0}%</div>
            <Progress value={health?.data?.cluster_storage || 0} className="mt-2" />
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Total Users</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-4xl font-bold">{health?.data?.total_users || 0}</div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Total VMs</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-4xl font-bold">{health?.data?.total_vms || 0}</div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Active VMs</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-4xl font-bold">{health?.data?.active_vms || 0}</div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
