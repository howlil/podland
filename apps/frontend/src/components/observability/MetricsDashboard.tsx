interface MetricsDashboardProps {
  metrics?: any;
  isLoading?: boolean;
  timeRange?: string;
}

export function MetricsDashboard({ metrics, isLoading, timeRange }: MetricsDashboardProps) {
  if (isLoading) {
    return (
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {Array.from({ length: 4 }).map((_, i) => (
          <div key={i} className="bg-white dark:bg-gray-800 rounded-lg p-4 animate-pulse">
            <div className="h-4 bg-gray-200 dark:bg-gray-700 rounded w-1/2 mb-2" />
            <div className="h-8 bg-gray-200 dark:bg-gray-700 rounded w-3/4 mb-2" />
            <div className="h-3 bg-gray-200 dark:bg-gray-700 rounded w-full" />
          </div>
        ))}
      </div>
    );
  }

  if (!metrics) {
    return (
      <div className="text-center py-12 text-gray-500 dark:text-gray-400">
        No metrics data available
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Time Range Selector */}
      <div className="flex items-center gap-2">
        {["1h", "6h", "24h", "7d", "30d"].map((range) => (
          <button
            key={range}
            className={`px-3 py-1 text-sm rounded-md transition-colors ${
              range === timeRange
                ? "bg-blue-600 text-white"
                : "bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700"
            }`}
          >
            {range}
          </button>
        ))}
      </div>

      {/* Metrics Summary Placeholder */}
      <div className="bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-lg p-4">
        <p className="text-sm text-blue-800 dark:text-blue-300">
          Metrics dashboard will display CPU, memory, network usage over time.
          Integration with Prometheus/Grafana coming soon.
        </p>
      </div>
    </div>
  );
}
