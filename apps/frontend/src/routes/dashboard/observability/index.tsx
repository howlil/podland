import { createFileRoute } from "@tanstack/react-router";
import { useState } from "react";
import { MetricsSummary } from "@/components/observability/MetricsSummary";
import { LogViewer } from "@/components/observability/LogViewer";
import { Activity, FileText, AlertTriangle, ExternalLink } from "lucide-react";

export const Route = createFileRoute("/dashboard/observability/")({
  component: ObservabilityPage,
});

function ObservabilityPage() {
  // Note: vmId should be passed via query param: ?vm=xxx
  const params = new URLSearchParams(window.location.search);
  const vmId = params.get("vm") || undefined;
  const [activeTab, setActiveTab] = useState<"metrics" | "logs" | "alerts">("metrics");
  const [timeRange, setTimeRange] = useState("24h");

  if (!vmId) {
    return (
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
              href="/dashboard/-vms"
              className="mt-4 inline-block text-blue-600 hover:text-blue-700 dark:text-blue-400"
            >
              Go to VM Dashboard →
            </a>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900 p-6">
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
        <div className="mb-6">
          <div className="border-b border-gray-200 dark:border-gray-700">
            <nav className="-mb-px flex gap-4">
              <TabButton
                icon={Activity}
                label="Metrics"
                active={activeTab === "metrics"}
                onClick={() => setActiveTab("metrics")}
              />
              <TabButton
                icon={FileText}
                label="Logs"
                active={activeTab === "logs"}
                onClick={() => setActiveTab("logs")}
              />
              <TabButton
                icon={AlertTriangle}
                label="Alerts"
                active={activeTab === "alerts"}
                onClick={() => setActiveTab("alerts")}
                badge={null}
              />
            </nav>
          </div>
        </div>

        {/* Tab Content */}
        {activeTab === "metrics" && (
          <div className="space-y-6">
            {/* Time Range Selector */}
            <div className="flex items-center gap-2">
              {["1h", "6h", "24h", "7d", "30d"].map((range) => (
                <button
                  key={range}
                  onClick={() => setTimeRange(range)}
                  className={`px-3 py-1 text-sm rounded-md transition-colors ${
                    timeRange === range
                      ? "bg-blue-600 text-white"
                      : "bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700"
                  }`}
                >
                  {range}
                </button>
              ))}
            </div>

            {/* Metrics Summary */}
            <MetricsSummary vmId={vmId} timeRange={timeRange} />

            {/* Note about detailed metrics */}
            <div className="bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-lg p-4">
              <p className="text-sm text-blue-800 dark:text-blue-300">
                For detailed metrics analysis, use the Grafana dashboard. Click "Open in Grafana" above.
              </p>
            </div>
          </div>
        )}

        {activeTab === "logs" && (
          <LogViewer vmId={vmId} limit={1000} showLiveTail={true} height="600px" />
        )}

        {activeTab === "alerts" && (
          <AlertHistory />
        )}
      </div>
    </div>
  );
}

interface TabButtonProps {
  icon: React.ElementType;
  label: string;
  active: boolean;
  onClick: () => void;
  badge?: number | null;
}

function TabButton({ icon: Icon, label, active, onClick, badge }: TabButtonProps) {
  return (
    <button
      onClick={onClick}
      className={`flex items-center gap-2 py-2 px-1 border-b-2 text-sm font-medium transition-colors ${
        active
          ? "border-blue-500 text-blue-600 dark:text-blue-400"
          : "border-transparent text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-300 hover:border-gray-300"
      }`}
    >
      <Icon className="h-4 w-4" />
      {label}
      {badge !== null && badge !== undefined && badge > 0 && (
        <span className="ml-1 px-2 py-0.5 text-xs bg-red-500 text-white rounded-full">
          {badge}
        </span>
      )}
    </button>
  );
}

function AlertHistory() {
  // TODO: Implement alert history fetching
  return (
    <div className="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-6">
      <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">
        Alert History
      </h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Alert history will be displayed here. This feature is coming soon.
      </p>
    </div>
  );
}
