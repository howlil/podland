import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import api from "@/lib/api";
import { DashboardLayout } from "@/components/layout/DashboardLayout";

interface VM {
  id: string;
  name: string;
  os: string;
  tier: string;
  cpu: number;
  ram: number;
  storage: number;
  status: "pending" | "running" | "stopped" | "error";
  domain: string;
  created_at: string;
}

export default function VMDetailRoute() {
  // Get VM ID from URL path
  const pathSegments = window.location.pathname.split("/");
  const id = pathSegments[pathSegments.length - 1];
  const queryClient = useQueryClient();

  const { data: vm, isLoading, error } = useQuery<VM>({
    queryKey: ["vm", id],
    queryFn: async () => {
      const { data } = await api.get(`/vms/${id}`);
      return data;
    },
    refetchInterval: 5000, // Poll every 5 seconds
    enabled: !!id,
  });

  const startMutation = useMutation({
    mutationFn: () => api.post(`/vms/${id}/start`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["vm", id] });
      queryClient.invalidateQueries({ queryKey: ["vms"] });
    },
  });

  const stopMutation = useMutation({
    mutationFn: () => api.post(`/vms/${id}/stop`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["vm", id] });
      queryClient.invalidateQueries({ queryKey: ["vms"] });
    },
  });

  const restartMutation = useMutation({
    mutationFn: () => api.post(`/vms/${id}/restart`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["vm", id] });
      queryClient.invalidateQueries({ queryKey: ["vms"] });
    },
  });

  const deleteMutation = useMutation({
    mutationFn: () => api.delete(`/vms/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["vm", id] });
      queryClient.invalidateQueries({ queryKey: ["vms"] });
      // Redirect to VM list after delete
      setTimeout(() => {
        window.location.href = "/dashboard/vms";
      }, 500);
    },
  });

  const handleStart = () => {
    if (confirm("Start this VM?")) {
      startMutation.mutate();
    }
  };

  const handleStop = () => {
    if (confirm("Stop this VM?")) {
      stopMutation.mutate();
    }
  };

  const handleRestart = () => {
    if (confirm("Restart this VM?")) {
      restartMutation.mutate();
    }
  };

  const handleDelete = () => {
    if (confirm("Are you sure you want to delete this VM? This action cannot be undone.")) {
      deleteMutation.mutate();
    }
  };

  const handleDownloadSSHKey = () => {
    // In a real implementation, the SSH key would be fetched from the API
    // For now, show a message that it's only shown once
    alert("SSH private key is only shown once during VM creation. If you didn't save it, you'll need to delete and recreate the VM.");
  };

  const getStatusBadgeClass = (status: string) => {
    switch (status) {
      case "running":
        return "bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400";
      case "stopped":
        return "bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300";
      case "pending":
        return "bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400";
      case "error":
        return "bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400";
      default:
        return "bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300";
    }
  };

  const formatBytes = (bytes: number) => {
    if (bytes === 0) return "0 B";
    const k = 1024;
    const sizes = ["B", "KB", "MB", "GB", "TB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
  };

  if (isLoading) {
    return (
      <DashboardLayout>
        <div className="max-w-4xl mx-auto">
          <div className="text-center py-12">
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto"></div>
            <p className="mt-4 text-gray-600 dark:text-gray-400">Loading VM...</p>
          </div>
        </div>
      </DashboardLayout>
    );
  }

  if (error || !vm) {
    return (
      <DashboardLayout>
        <div className="max-w-4xl mx-auto">
          <div className="text-center py-12">
            <p className="text-red-600 dark:text-red-400 text-lg font-medium">
              VM not found
            </p>
            <a
              href="/dashboard/-vms"
              className="mt-4 inline-block text-primary hover:text-primary/80"
            >
              ← Back to VMs
            </a>
          </div>
        </div>
      </DashboardLayout>
    );
  }

  return (
    <DashboardLayout>
      <div className="max-w-4xl mx-auto">
        {/* Header */}
        <div className="flex justify-between items-start mb-6">
          <div>
            <div className="flex items-center gap-3">
              <a
                href="/dashboard/-vms"
                className="text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200"
              >
                ← Back
              </a>
              <h1 className="text-2xl font-bold text-gray-900 dark:text-white">
                {vm.name}
              </h1>
            </div>
            <p className="text-gray-600 dark:text-gray-400 mt-1 ml-12">
              {vm.os === "ubuntu-2204" ? "Ubuntu 22.04" : "Debian 12"} · {vm.tier}
            </p>
          </div>
          <span
            className={`px-3 py-1 inline-flex text-sm font-semibold rounded-full ${getStatusBadgeClass(
              vm.status
            )}`}
          >
            {vm.status.toUpperCase()}
          </span>
        </div>

        {/* Main Content */}
        <div className="grid gap-6">
          {/* Resource Usage Card */}
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
            <h2 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">
              Resource Allocation
            </h2>
            <div className="grid grid-cols-3 gap-4">
              <div className="bg-gray-50 dark:bg-gray-700 rounded-lg p-4">
                <p className="text-sm text-gray-500 dark:text-gray-400">CPU</p>
                <p className="text-2xl font-bold text-gray-900 dark:text-white">
                  {vm.cpu} {vm.cpu === 1 ? "Core" : "Cores"}
                </p>
              </div>
              <div className="bg-gray-50 dark:bg-gray-700 rounded-lg p-4">
                <p className="text-sm text-gray-500 dark:text-gray-400">RAM</p>
                <p className="text-2xl font-bold text-gray-900 dark:text-white">
                  {formatBytes(vm.ram)}
                </p>
              </div>
              <div className="bg-gray-50 dark:bg-gray-700 rounded-lg p-4">
                <p className="text-sm text-gray-500 dark:text-gray-400">Storage</p>
                <p className="text-2xl font-bold text-gray-900 dark:text-white">
                  {formatBytes(vm.storage)}
                </p>
              </div>
            </div>
          </div>

          {/* Connection Info Card */}
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
            <h2 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">
              Connection Information
            </h2>
            <div className="space-y-4">
              <div>
                <p className="text-sm text-gray-500 dark:text-gray-400">Domain</p>
                <p className="text-lg font-mono text-gray-900 dark:text-white">
                  {vm.domain || `${vm.name}.podland.app`}
                </p>
              </div>
              <div>
                <p className="text-sm text-gray-500 dark:text-gray-400">SSH Access</p>
                <div className="flex items-center gap-2 mt-1">
                  <code className="bg-gray-100 dark:bg-gray-700 px-3 py-2 rounded text-sm font-mono text-gray-900 dark:text-white">
                    ssh -i ~/.ssh/id_ed25519 user@{vm.domain || `${vm.name}.podland.app`}
                  </code>
                  <button
                    onClick={handleDownloadSSHKey}
                    className="px-3 py-2 bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 rounded-lg text-sm font-medium text-gray-700 dark:text-gray-300 transition-colors"
                    title="Download SSH Key"
                  >
                    📥 Download Key
                  </button>
                </div>
                <p className="text-xs text-yellow-600 dark:text-yellow-400 mt-2">
                  ⚠️ The SSH private key was shown only once during VM creation. If you didn't save it, you'll need to recreate the VM.
                </p>
              </div>
            </div>
          </div>

          {/* Metadata Card */}
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
            <h2 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">
              Metadata
            </h2>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <p className="text-sm text-gray-500 dark:text-gray-400">VM ID</p>
                <p className="text-sm font-mono text-gray-900 dark:text-white">
                  {vm.id}
                </p>
              </div>
              <div>
                <p className="text-sm text-gray-500 dark:text-gray-400">Created</p>
                <p className="text-sm text-gray-900 dark:text-white">
                  {new Date(vm.created_at).toLocaleString()}
                </p>
              </div>
              <div>
                <p className="text-sm text-gray-500 dark:text-gray-400">Tier</p>
                <p className="text-sm text-gray-900 dark:text-white capitalize">
                  {vm.tier}
                </p>
              </div>
              <div>
                <p className="text-sm text-gray-500 dark:text-gray-400">Operating System</p>
                <p className="text-sm text-gray-900 dark:text-white">
                  {vm.os === "ubuntu-2204" ? "Ubuntu 22.04 LTS" : "Debian 12"}
                </p>
              </div>
            </div>
          </div>

          {/* Actions Card */}
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
            <h2 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">
              Actions
            </h2>
            <div className="flex flex-wrap gap-3">
              {vm.status === "stopped" && (
                <button
                  onClick={handleStart}
                  disabled={startMutation.isPending}
                  className="px-4 py-2 bg-green-600 hover:bg-green-700 disabled:bg-green-800 text-white rounded-lg font-medium transition-colors"
                >
                  {startMutation.isPending ? "Starting..." : "▶ Start VM"}
                </button>
              )}
              {vm.status === "running" && (
                <>
                  <button
                    onClick={handleStop}
                    disabled={stopMutation.isPending}
                    className="px-4 py-2 bg-yellow-600 hover:bg-yellow-700 disabled:bg-yellow-800 text-white rounded-lg font-medium transition-colors"
                  >
                    {stopMutation.isPending ? "Stopping..." : "⏹ Stop VM"}
                  </button>
                  <button
                    onClick={handleRestart}
                    disabled={restartMutation.isPending}
                    className="px-4 py-2 bg-blue-600 hover:bg-blue-700 disabled:bg-blue-800 text-white rounded-lg font-medium transition-colors"
                  >
                    {restartMutation.isPending ? "Restarting..." : "🔄 Restart VM"}
                  </button>
                </>
              )}
              <button
                onClick={handleDelete}
                disabled={deleteMutation.isPending}
                className="px-4 py-2 bg-red-600 hover:bg-red-700 disabled:bg-red-800 text-white rounded-lg font-medium transition-colors"
              >
                {deleteMutation.isPending ? "Deleting..." : "🗑 Delete VM"}
              </button>
            </div>
          </div>
        </div>
      </div>
    </DashboardLayout>
  );
}
