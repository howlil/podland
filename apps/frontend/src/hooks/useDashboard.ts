import { useQuery } from "@tanstack/react-query";
import api from "@/lib/api";
import { useVMs } from "./useVMs";

interface Quota {
  cpu_limit: number;
  cpu_used: number;
  ram_limit: number;
  ram_used: number;
  storage_limit: number;
  storage_used: number;
  vm_count: number;
  vm_count_limit: number;
}

interface DashboardStats {
  totalVMs: number;
  runningVMs: number;
  totalCPU: number;
  totalRAM: number;
  totalStorage: number;
  domainsCount: number;
}

export function useDashboard() {
  const { vms, isLoading: vmsLoading } = useVMs();

  const { data: quota, isLoading: quotaLoading } = useQuery<Quota>({
    queryKey: ["quota"],
    queryFn: async () => {
      const { data } = await api.get("/users/me");
      return {
        cpu_limit: data.role === "internal" ? 4.0 : 0.5,
        cpu_used: 0,
        ram_limit: data.role === "internal" ? 8589934592 : 1073741824,
        ram_used: 0,
        storage_limit: data.role === "internal" ? 107374182400 : 10737418240,
        storage_used: 0,
        vm_count: vms.length,
        vm_count_limit: data.role === "internal" ? 5 : 2,
      };
    },
    enabled: !!vms,
  });

  const stats: DashboardStats | undefined = vms
    ? {
        totalVMs: vms.length,
        runningVMs: vms.filter((vm) => vm.status === "running").length,
        totalCPU: vms.reduce((sum, vm) => sum + vm.cpu, 0),
        totalRAM: vms.reduce((sum, vm) => sum + vm.ram, 0),
        totalStorage: vms.reduce((sum, vm) => sum + vm.storage, 0),
        domainsCount: vms.filter((vm) => vm.domain).length,
      }
    : undefined;

  return {
    stats,
    quota,
    vms,
    isLoading: vmsLoading || quotaLoading,
  };
}
