import { MetricsDashboard } from "./MetricsDashboard";
import { LogViewer } from "./LogViewer";
import { AlertsList } from "./AlertsList";

interface TabContentProps {
  activeTab: string;
  metrics?: any;
  alerts?: any;
  isLoading: boolean;
  timeRange: string;
  vmId: string;
}

export function TabContent({
  activeTab,
  metrics,
  alerts,
  isLoading,
  timeRange,
  vmId,
}: TabContentProps) {
  if (activeTab === "metrics") {
    return (
      <MetricsDashboard
        metrics={metrics}
        isLoading={isLoading}
        timeRange={timeRange}
      />
    );
  }

  if (activeTab === "logs") {
    return (
      <LogViewer vmId={vmId} limit={1000} showLiveTail height="600px" />
    );
  }

  if (activeTab === "alerts") {
    return (
      <AlertsList alerts={alerts} isLoading={isLoading} />
    );
  }

  return null;
}
