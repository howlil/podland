import { AlertTriangle } from "lucide-react";

interface AlertsListProps {
  alerts?: any[];
  isLoading?: boolean;
}

export function AlertsList({ alerts, isLoading }: AlertsListProps) {
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
        {alerts.map((alert, index) => (
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
