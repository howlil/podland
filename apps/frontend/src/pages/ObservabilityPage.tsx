import { useState } from "react";
import { useObservability } from "@/hooks/useObservability";
import { DashboardLayout } from "@/components/layout/DashboardLayout";
import { TabNav } from "@/components/observability/TabNav";
import { MetricsDashboard } from "@/components/observability/MetricsDashboard";
import { LogViewer } from "@/components/observability/LogViewer";
import { Play, Square, Activity, FileText, AlertTriangle, ExternalLink } from "lucide-react";

export default function ObservabilityPage() {
  const [vmId, setVmId] = useState<string>("");

  // For now, get vmId from query param or user input
  // In production, this would come from route params
  const {
    activeTab,
    setActiveTab,
    timeRange,
    setTimeRange,
    metrics,
    logs,
    alerts,
    isLoading,
  } = useObservability(vmId);

  if (!vmId) {
    return (
      <DashboardLayout>
        <div className="min-h-screen bg-gray-50 dark:bg-gray-900 p-6">
          <div className="max-w-7xl mx-auto">
            <div className="text-center py-12">
              <h1 className="text-2xl font-bold text-gray-900 dark:text-white mb-2">
                No VM Selected
              </h1>
              <p className="text-gray-600 dark:text-gray-400">
                Please select a VM from the dashboard to view observability data.
              </p>
              <a
                href="/dashboard/vms"
                className="mt-4 inline-block text-blue-600 hover:text-blue-700 dark:text-blue-400"
              >
                Go to VM Dashboard →
              </a>
            </div>
          </div>
        </div>
      </DashboardLayout>
    );
  }

  return (
    <DashboardLayout>
      <div className="max-w-7xl mx-auto">
        {/* Header */}
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-6">
          <div>
            <h1 className="text-2xl font-bold text-gray-900 dark:text-white">
              Observability
            </h1>
            <p className="text-sm text-gray-600 dark:text-gray-400 mt-1">
              VM ID: {vmId}
            </p>
          </div>
          <a
            href={`http://grafana.monitoring.svc:3000/d/vm-metrics?var-vm_id=${vmId}`}
            target="_blank"
            rel="noopener noreferrer"
            className="inline-flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg text-sm font-medium transition-colors"
          >
            <ExternalLink className="h-4 w-4" />
            Open in Grafana
          </a>
        </div>

        {/* Tabs */}
        <TabNav
          activeTab={activeTab}
          onTabChange={setActiveTab}
          alertsCount={alerts?.length || 0}
        />

        {/* Tab Content */}
        {activeTab === "metrics" && (
          <MetricsDashboard
            metrics={metrics}
            isLoading={isLoading}
            timeRange={timeRange}
          />
        )}

        {activeTab === "logs" && (
          <LogViewer vmId={vmId} limit={1000} showLiveTail height="600px" />
        )}

        {activeTab === "alerts" && (
          <AlertsList alerts={alerts} isLoading={isLoading} />
        )}
      </div>
    </DashboardLayout>
  );
}

interface AlertsListProps {
  alerts?: any;
  isLoading?: boolean;
}

function AlertsList({ alerts, isLoading }: AlertsListProps) {
  if (isLoading) {
    return (
      <div className="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-6 animate-pulse">
        <div className="h-6 w-32 bg-gray-200 dark:bg-gray-700 rounded mb-4" />
        <div className="space-y-3">
          {Array.from({ length: 3 }).map((_, i) => (
            <div key={i} className="h-16 bg-gray-200 dark:bg-gray-700 rounded" />
          ))}
        </div>
      </div>
    );
  }

  if (!alerts || alerts.length === 0) {
    return (
      <div className="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-6 text-center">
        <AlertTriangle className="h-16 w-16 mx-auto mb-4 text-gray-400 dark:text-gray-600" />
        <p className="text-lg font-medium text-gray-900 dark:text-white mb-2">
          No alerts
        </p>
        <p className="text-gray-600 dark:text-gray-400">
          Your VM is running smoothly. No alerts to display.
        </p>
      </div>
    );
  }

  return (
    <div className="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-6">
      <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">
        Alert History
      </h3>
      <div className="space-y-3">
        {alerts.map((alert: any, index: number) => (
          <div
            key={index}
            className="p-4 border border-gray-200 dark:border-gray-700 rounded-lg"
          >
            <p className="font-medium text-gray-900 dark:text-white">
              {alert.name}
            </p>
            <p className="text-sm text-gray-600 dark:text-gray-400">
              {alert.message}
            </p>
          </div>
        ))}
      </div>
    </div>
  );
}
