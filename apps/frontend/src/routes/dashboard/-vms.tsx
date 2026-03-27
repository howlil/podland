import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import api from "@/lib/api";
import { DashboardLayout } from "@/components/layout/DashboardLayout";
import { CreateVMWizard } from "@/components/vm/CreateVMWizard";

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

type SortField = "name" | "created_at" | "status";
type SortOrder = "asc" | "desc";

export default function VMsRoute() {
  const [statusFilter, setStatusFilter] = useState<string>("all");
  const [sortField, setSortField] = useState<SortField>("created_at");
  const [sortOrder, setSortOrder] = useState<SortOrder>("desc");
  const [isWizardOpen, setIsWizardOpen] = useState(false);

  const { data: vms = [], isLoading, refetch } = useQuery<VM[]>({
    queryKey: ["vms"],
    queryFn: async () => {
      const { data } = await api.get("/vms");
      return data;
    },
    refetchInterval: 5000, // Poll every 5 seconds
  });

  const queryClient = useQueryClient();

  const startMutation = useMutation({
    mutationFn: (vmId: string) => api.post(`/vms/${vmId}/start`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["vms"] });
    },
  });

  const stopMutation = useMutation({
    mutationFn: (vmId: string) => api.post(`/vms/${vmId}/stop`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["vms"] });
    },
  });

  const restartMutation = useMutation({
    mutationFn: (vmId: string) => api.post(`/vms/${vmId}/restart`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["vms"] });
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (vmId: string) => api.delete(`/vms/${vmId}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["vms"] });
    },
  });

  const handleStart = (vmId: string) => {
    if (confirm("Start this VM?")) {
      startMutation.mutate(vmId);
    }
  };

  const handleStop = (vmId: string) => {
    if (confirm("Stop this VM?")) {
      stopMutation.mutate(vmId);
    }
  };

  const handleRestart = (vmId: string) => {
    if (confirm("Restart this VM?")) {
      restartMutation.mutate(vmId);
    }
  };

  const handleDelete = (vmId: string) => {
    if (confirm("Are you sure you want to delete this VM? This action cannot be undone.")) {
      deleteMutation.mutate(vmId);
    }
  };

  const handleSort = (field: SortField) => {
    if (sortField === field) {
      setSortOrder(sortOrder === "asc" ? "desc" : "asc");
    } else {
      setSortField(field);
      setSortOrder("asc");
    }
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

  // Filter and sort VMs
  let filteredVms = vms.filter((vm) => {
    if (statusFilter === "all") return true;
    return vm.status === statusFilter;
  });

  filteredVms.sort((a, b) => {
    let aVal: string | number = a[sortField];
    let bVal: string | number = b[sortField];

    if (sortField === "created_at") {
      aVal = new Date(a.created_at).getTime();
      bVal = new Date(b.created_at).getTime();
    }

    if (aVal < bVal) return sortOrder === "asc" ? -1 : 1;
    if (aVal > bVal) return sortOrder === "asc" ? 1 : -1;
    return 0;
  });

  return (
    <DashboardLayout>
      <div className="max-w-7xl mx-auto">
        <div className="flex justify-between items-center mb-6">
          <div>
            <h1 className="text-2xl font-bold text-gray-900 dark:text-white">
              My VMs
            </h1>
            <p className="text-gray-600 dark:text-gray-400 mt-1">
              Manage your virtual machines
            </p>
          </div>
          <button
            onClick={() => setIsWizardOpen(true)}
            className="px-4 py-2 bg-primary text-white rounded-lg hover:bg-primary/90 transition-colors font-medium"
          >
            + Create VM
          </button>
        </div>

        {/* Filters */}
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-4 mb-6">
          <div className="flex flex-wrap gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Status
              </label>
              <select
                value={statusFilter}
                onChange={(e) => setStatusFilter(e.target.value)}
                className="px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-primary focus:border-transparent"
              >
                <option value="all">All</option>
                <option value="running">Running</option>
                <option value="stopped">Stopped</option>
                <option value="pending">Pending</option>
                <option value="error">Error</option>
              </select>
            </div>
          </div>
        </div>

        {/* VM Table */}
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
          <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
            <thead className="bg-gray-50 dark:bg-gray-700">
              <tr>
                <th
                  onClick={() => handleSort("name")}
                  className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider cursor-pointer hover:text-gray-700 dark:hover:text-gray-100"
                >
                  Name {sortField === "name" && (sortOrder === "asc" ? "↑" : "↓")}
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                  OS
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                  Tier
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                  Resources
                </th>
                <th
                  onClick={() => handleSort("status")}
                  className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider cursor-pointer hover:text-gray-700 dark:hover:text-gray-100"
                >
                  Status {sortField === "status" && (sortOrder === "asc" ? "↑" : "↓")}
                </th>
                <th
                  onClick={() => handleSort("created_at")}
                  className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider cursor-pointer hover:text-gray-700 dark:hover:text-gray-100"
                >
                  Created {sortField === "created_at" && (sortOrder === "asc" ? "↑" : "↓")}
                </th>
                <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody className="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
              {isLoading ? (
                <tr>
                  <td colSpan={7} className="px-6 py-8 text-center text-gray-500 dark:text-gray-400">
                    Loading VMs...
                  </td>
                </tr>
              ) : filteredVms.length === 0 ? (
                <tr>
                  <td colSpan={7} className="px-6 py-8 text-center text-gray-500 dark:text-gray-400">
                    {statusFilter === "all"
                      ? "No VMs yet. Create your first VM to get started!"
                      : `No ${statusFilter} VMs.`}
                  </td>
                </tr>
              ) : (
                filteredVms.map((vm) => (
                  <tr key={vm.id} className="hover:bg-gray-50 dark:hover:bg-gray-700/50">
                    <td className="px-6 py-4 whitespace-nowrap">
                      <a
                        href={`/dashboard/-vms/${vm.id}`}
                        className="text-primary hover:text-primary/80 font-medium"
                      >
                        {vm.name}
                      </a>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-gray-600 dark:text-gray-300">
                      {vm.os === "ubuntu-2204" ? "Ubuntu 22.04" : "Debian 12"}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-gray-600 dark:text-gray-300">
                      {vm.tier}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-gray-600 dark:text-gray-300">
                      {vm.cpu} CPU · {formatBytes(vm.ram)} RAM · {formatBytes(vm.storage)} Disk
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span
                        className={`px-2 py-1 inline-flex text-xs leading-5 font-semibold rounded-full ${getStatusBadgeClass(
                          vm.status
                        )}`}
                      >
                        {vm.status}
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-gray-600 dark:text-gray-300">
                      {new Date(vm.created_at).toLocaleDateString()}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                      {vm.status === "stopped" && (
                        <button
                          onClick={() => handleStart(vm.id)}
                          className="text-green-600 hover:text-green-900 dark:text-green-400 dark:hover:text-green-300 mr-3"
                        >
                          Start
                        </button>
                      )}
                      {vm.status === "running" && (
                        <>
                          <button
                            onClick={() => handleStop(vm.id)}
                            className="text-yellow-600 hover:text-yellow-900 dark:text-yellow-400 dark:hover:text-yellow-300 mr-3"
                          >
                            Stop
                          </button>
                          <button
                            onClick={() => handleRestart(vm.id)}
                            className="text-blue-600 hover:text-blue-900 dark:text-blue-400 dark:hover:text-blue-300 mr-3"
                          >
                            Restart
                          </button>
                        </>
                      )}
                      <button
                        onClick={() => handleDelete(vm.id)}
                        className="text-red-600 hover:text-red-900 dark:text-red-400 dark:hover:text-red-300"
                      >
                        Delete
                      </button>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>

      {/* Create VM Wizard Modal */}
      {isWizardOpen && (
        <CreateVMWizard
          onClose={() => setIsWizardOpen(false)}
          onSuccess={() => {
            setIsWizardOpen(false);
            refetch();
          }}
        />
      )}
    </DashboardLayout>
  );
}
