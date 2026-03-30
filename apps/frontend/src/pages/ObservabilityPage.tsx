import { useState } from "react";
import { useObservability } from "@/hooks/useObservability";
import { DashboardLayout } from "@/components/layout/DashboardLayout";
import { TabNav } from "@/components/observability/TabNav";
import { ObservabilityHeader } from "@/components/observability/ObservabilityHeader";
import { TabContent } from "@/components/observability/TabContent";
import { EmptyState } from "@/components/ui/EmptyState";

export default function ObservabilityPage() {
  const [vmId] = useState<string>("");
  const {
    activeTab,
    setActiveTab,
    metrics,
    timeRange,
    alerts,
    isLoading,
  } = useObservability(vmId);

  if (!vmId) {
    return (
      <DashboardLayout>
        <EmptyState
          title="No VM Selected"
          description="Please select a VM from the dashboard to view observability data."
          actionLabel="Go to VM Dashboard"
          actionHref="/dashboard/vms"
        />
      </DashboardLayout>
    );
  }

  return (
    <DashboardLayout>
      <div className="max-w-7xl mx-auto">
        <ObservabilityHeader vmId={vmId} />
        <TabNav
          activeTab={activeTab}
          onTabChange={setActiveTab}
          alertsCount={alerts?.length || 0}
        />
        <TabContent
          activeTab={activeTab}
          metrics={metrics}
          alerts={alerts}
          isLoading={isLoading}
          timeRange={timeRange}
          vmId={vmId}
        />
      </div>
    </DashboardLayout>
  );
}
