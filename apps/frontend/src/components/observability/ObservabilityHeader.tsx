import { ExternalLink } from "lucide-react";

interface ObservabilityHeaderProps {
  vmId: string;
  grafanaUrl?: string;
}

export function ObservabilityHeader({ vmId, grafanaUrl }: ObservabilityHeaderProps) {
  const defaultGrafanaUrl = `http://grafana.monitoring.svc:3000/d/vm-metrics?var-vm_id=${vmId}`;
  const url = grafanaUrl || defaultGrafanaUrl;

  return (
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
        href={url}
        target="_blank"
        rel="noopener noreferrer"
        className="inline-flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg text-sm font-medium transition-colors"
      >
        <ExternalLink className="h-4 w-4" />
        Open in Grafana
      </a>
    </div>
  );
}
