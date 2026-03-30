import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import api from "@/lib/api";

type ActiveTab = "metrics" | "logs" | "alerts";
type TimeRange = "1h" | "6h" | "24h" | "7d" | "30d";

interface ObservabilityData {
  metrics?: any;
  logs?: any;
  alerts?: any;
}

export function useObservability(vmId: string) {
  const [activeTab, setActiveTab] = useState<ActiveTab>("metrics");
  const [timeRange, setTimeRange] = useState<TimeRange>("24h");

  const { data: metrics, isLoading: metricsLoading } = useQuery({
    queryKey: ["metrics", vmId, timeRange],
    queryFn: async () => {
      const { data } = await api.get(`/vms/${vmId}/metrics?range=${timeRange}`);
      return data;
    },
    refetchInterval: 30000,
    enabled: activeTab === "metrics" && !!vmId,
  });

  const { data: logs, isLoading: logsLoading } = useQuery({
    queryKey: ["logs", vmId],
    queryFn: async () => {
      const { data } = await api.get(`/vms/${vmId}/logs?limit=1000`);
      return data;
    },
    refetchInterval: activeTab === "logs" ? 5000 : false,
    enabled: activeTab === "logs" && !!vmId,
  });

  const { data: alerts, isLoading: alertsLoading } = useQuery({
    queryKey: ["alerts", vmId],
    queryFn: async () => {
      const { data } = await api.get(`/vms/${vmId}/alerts`);
      return data;
    },
    enabled: activeTab === "alerts" && !!vmId,
  });

  return {
    activeTab,
    setActiveTab,
    timeRange,
    setTimeRange,
    metrics,
    logs,
    alerts,
    isLoading: metricsLoading || logsLoading || alertsLoading,
  };
}
