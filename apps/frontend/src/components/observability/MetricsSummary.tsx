import { useQuery } from "@tanstack/react-query";
import { api } from "../../lib/api";
import { Activity, Server, Network, HardDrive } from "lucide-react";

interface MetricSeries {
  current: number;
  average: number;
  max: number;
  min: number;
  points?: { timestamp: number; value: number }[];
}

interface VMMetrics {
  cpu?: MetricSeries;
  memory?: MetricSeries;
  network_rx?: MetricSeries;
  network_tx?: MetricSeries;
}

interface MetricsSummaryProps {
  vmId: string;
  timeRange?: string;
}

export function MetricsSummary({ vmId, timeRange = "24h" }: MetricsSummaryProps) {
  const { data, isLoading, error } = useQuery<{ data: VMMetrics }>({
    queryKey: ["metrics", vmId, timeRange],
    queryFn: () => api.get(`/vms/${vmId}/metrics?range=${timeRange}`),
    refetchInterval: 30000, // Refresh every 30 seconds
    retry: 2,
  });

  if (isLoading) {
    return (
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {[1, 2, 3, 4].map((i) => (
          <div key={i} className="bg-white dark:bg-gray-800 rounded-lg p-4 animate-pulse">
            <div className="h-4 bg-gray-200 dark:bg-gray-700 rounded w-1/2 mb-2"></div>
            <div className="h-8 bg-gray-200 dark:bg-gray-700 rounded w-3/4 mb-2"></div>
            <div className="h-3 bg-gray-200 dark:bg-gray-700 rounded w-full"></div>
          </div>
        ))}
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-4">
        <p className="text-red-600 dark:text-red-400 text-sm">
          Failed to load metrics. Please try again later.
        </p>
      </div>
    );
  }

  const metrics = data?.data;

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
      <MetricCard
        icon={Activity}
        title="CPU Usage"
        value={formatCPU(metrics?.cpu?.current)}
        average={formatCPU(metrics?.cpu?.average)}
        color="blue"
      />
      <MetricCard
        icon={Server}
        title="Memory Usage"
        value={formatBytes(metrics?.memory?.current)}
        average={formatBytes(metrics?.memory?.average)}
        color="green"
      />
      <MetricCard
        icon={Network}
        title="Network RX"
        value={formatBytesPerSec(metrics?.network_rx?.current)}
        average={formatBytesPerSec(metrics?.network_rx?.average)}
        color="purple"
      />
      <MetricCard
        icon={HardDrive}
        title="Network TX"
        value={formatBytesPerSec(metrics?.network_tx?.current)}
        average={formatBytesPerSec(metrics?.network_tx?.average)}
        color="orange"
      />
    </div>
  );
}

interface MetricCardProps {
  icon: React.ElementType;
  title: string;
  value: string;
  average: string;
  color: "blue" | "green" | "purple" | "orange";
}

function MetricCard({ icon: Icon, title, value, average, color }: MetricCardProps) {
  const colorClasses = {
    blue: "text-blue-600 dark:text-blue-400 bg-blue-50 dark:bg-blue-900/20",
    green: "text-green-600 dark:text-green-400 bg-green-50 dark:bg-green-900/20",
    purple: "text-purple-600 dark:text-purple-400 bg-purple-50 dark:bg-purple-900/20",
    orange: "text-orange-600 dark:text-orange-400 bg-orange-50 dark:bg-orange-900/20",
  };

  return (
    <div className="bg-white dark:bg-gray-800 rounded-lg p-4 border border-gray-200 dark:border-gray-700">
      <div className="flex items-center justify-between mb-2">
        <span className="text-sm font-medium text-gray-600 dark:text-gray-400">{title}</span>
        <Icon className={`h-5 w-5 ${colorClasses[color].split(" ")[0]}`} />
      </div>
      <div className="text-2xl font-bold text-gray-900 dark:text-white mb-1">{value}</div>
      <div className="text-xs text-gray-500 dark:text-gray-400">Avg: {average}</div>
    </div>
  );
}

function formatCPU(value?: number): string {
  if (value === undefined || value === null) return "N/A";
  return `${(value * 100).toFixed(1)}%`;
}

function formatBytes(value?: number): string {
  if (value === undefined || value === null) return "N/A";
  if (value === 0) return "0 B";
  const k = 1024;
  const sizes = ["B", "KB", "MB", "GB", "TB"];
  const i = Math.floor(Math.log(value) / Math.log(k));
  return `${(value / Math.pow(k, i)).toFixed(1)} ${sizes[i]}`;
}

function formatBytesPerSec(value?: number): string {
  if (value === undefined || value === null) return "N/A";
  return `${formatBytes(value)}/s`;
}
