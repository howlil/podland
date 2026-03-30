import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import api from "@/lib/api";
import { toast } from "sonner";

export interface VM {
  id: string;
  name: string;
  os: string;
  tier: string;
  cpu: number;
  ram: number;
  storage: number;
  status: "pending" | "running" | "stopped" | "error";
  domain?: string;
  is_pinned?: boolean;
  created_at: string;
}

export function useVMs() {
  const queryClient = useQueryClient();

  const { data: vms = [], isLoading, error, refetch } = useQuery<VM[]>({
    queryKey: ["vms"],
    queryFn: async () => {
      const { data } = await api.get("/vms");
      return data;
    },
    refetchInterval: 5000,
    retry: 2,
  });

  const startMutation = useMutation({
    mutationFn: (vmId: string) => api.post(`/vms/${vmId}/start`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["vms"] });
      toast.success("VM started successfully");
    },
    onError: (error: any) => {
      toast.error(`Failed to start VM: ${error.response?.data?.message || "Unknown error"}`);
    },
  });

  const stopMutation = useMutation({
    mutationFn: (vmId: string) => api.post(`/vms/${vmId}/stop`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["vms"] });
      toast.success("VM stopped successfully");
    },
    onError: (error: any) => {
      toast.error(`Failed to stop VM: ${error.response?.data?.message || "Unknown error"}`);
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (vmId: string) => api.delete(`/vms/${vmId}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["vms"] });
      toast.success("VM deleted successfully");
    },
    onError: (error: any) => {
      toast.error(`Failed to delete VM: ${error.response?.data?.message || "Unknown error"}`);
    },
  });

  return {
    vms,
    isLoading,
    error,
    refetch,
    startVM: startMutation.mutate,
    stopVM: stopMutation.mutate,
    deleteVM: deleteMutation.mutate,
    isStarting: startMutation.isPending,
    isStopping: stopMutation.isPending,
    isDeleting: deleteMutation.isPending,
  };
}

export function useVM(id: string) {
  const queryClient = useQueryClient();

  const { data: vm, isLoading, error } = useQuery<VM>({
    queryKey: ["vm", id],
    queryFn: async () => {
      const { data } = await api.get(`/vms/${id}`);
      return data;
    },
    refetchInterval: 5000,
    enabled: !!id,
  });

  const startMutation = useMutation({
    mutationFn: () => api.post(`/vms/${id}/start`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["vm", id] });
      queryClient.invalidateQueries({ queryKey: ["vms"] });
      toast.success("VM started successfully");
    },
    onError: (error: any) => {
      toast.error(`Failed to start VM: ${error.response?.data?.message || "Unknown error"}`);
    },
  });

  const stopMutation = useMutation({
    mutationFn: () => api.post(`/vms/${id}/stop`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["vm", id] });
      queryClient.invalidateQueries({ queryKey: ["vms"] });
      toast.success("VM stopped successfully");
    },
    onError: (error: any) => {
      toast.error(`Failed to stop VM: ${error.response?.data?.message || "Unknown error"}`);
    },
  });

  const deleteMutation = useMutation({
    mutationFn: () => api.delete(`/vms/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["vm", id] });
      queryClient.invalidateQueries({ queryKey: ["vms"] });
      toast.success("VM deleted successfully");
    },
    onError: (error: any) => {
      toast.error(`Failed to delete VM: ${error.response?.data?.message || "Unknown error"}`);
    },
  });

  const restartMutation = useMutation({
    mutationFn: () => api.post(`/vms/${id}/restart`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["vm", id] });
      queryClient.invalidateQueries({ queryKey: ["vms"] });
      toast.success("VM restarted successfully");
    },
    onError: (error: any) => {
      toast.error(`Failed to restart VM: ${error.response?.data?.message || "Unknown error"}`);
    },
  });

  const pinMutation = useMutation({
    mutationFn: () => api.post(`/vms/${id}/pin`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["vm", id] });
      toast.success("VM pinned successfully");
    },
    onError: (error: any) => {
      toast.error(`Failed to pin VM: ${error.response?.data?.message || "Unknown error"}`);
    },
  });

  const unpinMutation = useMutation({
    mutationFn: () => api.delete(`/vms/${id}/pin`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["vm", id] });
      toast.success("VM unpinned successfully");
    },
    onError: (error: any) => {
      toast.error(`Failed to unpin VM: ${error.response?.data?.message || "Unknown error"}`);
    },
  });

  return {
    vm,
    isLoading,
    error,
    startVM: () => startMutation.mutate(),
    stopVM: () => stopMutation.mutate(),
    restartVM: () => restartMutation.mutate(),
    deleteVM: () => deleteMutation.mutate(),
    pinVM: () => pinMutation.mutate(),
    unpinVM: () => unpinMutation.mutate(),
    isStarting: startMutation.isPending,
    isStopping: stopMutation.isPending,
    isRestarting: restartMutation.isPending,
    isDeleting: deleteMutation.isPending,
    isPinning: pinMutation.isPending,
    isUnpinning: unpinMutation.isPending,
  };
}
