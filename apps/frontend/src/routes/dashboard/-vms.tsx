import { useState, useEffect } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import api from "@/lib/api";
import { DashboardLayout } from "@/components/layout/DashboardLayout";
import { CreateVMWizard } from "@/components/vm/CreateVMWizard";
import { Skeleton } from "@/components/ui/skeleton";
import { formatBytes } from "@/lib/utils";
import { POLLING_INTERVALS, VM_STATUS, UI } from "@/lib/constants";

interface VM {
  id: string;
  name: string;
  os: string;
  tier: string;
  cpu: number;
  ram: number;
  storage: number;
  status: typeof VM_STATUS[keyof typeof VM_STATUS];
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
  const [currentPage, setCurrentPage] = useState(1);
  const [isTabVisible, setIsTabVisible] = useState(true);

  // Track tab visibility for polling optimization
  useEffect(() => {
    const handleVisibilityChange = () => {
      setIsTabVisible(document.visibilityState === "visible");
    };

    document.addEventListener("visibilitychange", handleVisibilityChange);
    return () => document.removeEventListener("visibilitychange", handleVisibilityChange);
  }, []);

  const { data: vms = [], isLoading, error, refetch } = useQuery<VM[]>({
    queryKey: ["vms"],
    queryFn: async () => {
      const { data } = await api.get("/vms");
      return data;
    },
    // Only poll when tab is visible
    refetchInterval: () => {
      return isTabVisible ? POLLING_INTERVALS.VM_STATUS : false;
    },
    retry: 2,
  });

  const queryClient = useQueryClient();

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

  const restartMutation = useMutation({
    mutationFn: (vmId: string) => api.post(`/vms/${vmId}/restart`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["vms"] });
      toast.success("VM restarted successfully");
    },
    onError: (error: any) => {
      toast.error(`Failed to restart VM: ${error.response?.data?.message || "Unknown error"}`);
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

  const handleStart = (vmId: string) => {
    toast.loading("Starting VM...", { id: `vm-${vmId}` });
    startMutation.mutate(vmId);
  };

  const handleStop = (vmId: string) => {
    toast.loading("Stopping VM...", { id: `vm-${vmId}` });
    stopMutation.mutate(vmId);
  };

  const handleRestart = (vmId: string) => {
    toast.loading("Restarting VM...", { id: `vm-${vmId}` });
    restartMutation.mutate(vmId);
  };

  const handleDelete = (vmId: string) => {
    toast.loading("Deleting VM...", { id: `vm-${vmId}` });
    deleteMutation.mutate(vmId);
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

  // Pagination
  const totalPages = Math.max(1, Math.ceil(filteredVms.length / UI.TABLE_PAGE_SIZE));
  const currentPageSafe = Math.min(currentPage, totalPages);
  const paginatedVms = filteredVms.slice(
    (currentPageSafe - 1) * UI.TABLE_PAGE_SIZE,
    currentPageSafe * UI.TABLE_PAGE_SIZE
  );

  // Reset to page 1 when filters change
  useEffect(() => {
    setCurrentPage(1);
  }, [statusFilter, sortField, sortOrder]);

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
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden" role="region" aria-label="VMs list">
          <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
            <thead className="bg-gray-50 dark:bg-gray-700">
              <tr>
                <th
                  scope="col"
                  onClick={() => handleSort("name")}
                  className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider cursor-pointer hover:text-gray-700 dark:hover:text-gray-100"
                  aria-sort={sortField === "name" ? (sortOrder === "asc" ? "ascending" : "descending") : "none"}
                  tabIndex={0}
                  onKeyDown={(e) => {
                    if (e.key === "Enter" || e.key === " ") {
                      e.preventDefault();
                      handleSort("name");
                    }
                  }}
                >
                  <span className="flex items-center gap-1">
                    Name
                    {sortField === "name" && (
                      <span aria-hidden="true">{sortOrder === "asc" ? "↑" : "↓"}</span>
                    )}
                  </span>
                </th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                  OS
                </th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                  Tier
                </th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                  Resources
                </th>
                <th
                  scope="col"
                  onClick={() => handleSort("status")}
                  className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider cursor-pointer hover:text-gray-700 dark:hover:text-gray-100"
                  aria-sort={sortField === "status" ? (sortOrder === "asc" ? "ascending" : "descending") : "none"}
                  tabIndex={0}
                  onKeyDown={(e) => {
                    if (e.key === "Enter" || e.key === " ") {
                      e.preventDefault();
                      handleSort("status");
                    }
                  }}
                >
                  <span className="flex items-center gap-1">
                    Status
                    {sortField === "status" && (
                      <span aria-hidden="true">{sortOrder === "asc" ? "↑" : "↓"}</span>
                    )}
                  </span>
                </th>
                <th
                  scope="col"
                  onClick={() => handleSort("created_at")}
                  className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider cursor-pointer hover:text-gray-700 dark:hover:text-gray-100"
                  aria-sort={sortField === "created_at" ? (sortOrder === "asc" ? "ascending" : "descending") : "none"}
                  tabIndex={0}
                  onKeyDown={(e) => {
                    if (e.key === "Enter" || e.key === " ") {
                      e.preventDefault();
                      handleSort("created_at");
                    }
                  }}
                >
                  <span className="flex items-center gap-1">
                    Created
                    {sortField === "created_at" && (
                      <span aria-hidden="true">{sortOrder === "asc" ? "↑" : "↓"}</span>
                    )}
                  </span>
                </th>
                <th scope="col" className="px-6 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody className="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
              {isLoading ? (
                // Skeleton loading state
                Array.from({ length: UI.TABLE_PAGE_SIZE }).map((_, idx) => (
                  <tr key={idx} className="animate-pulse">
                    <td className="px-6 py-4"><Skeleton className="h-4 w-32" /></td>
                    <td className="px-6 py-4"><Skeleton className="h-4 w-24" /></td>
                    <td className="px-6 py-4"><Skeleton className="h-4 w-20" /></td>
                    <td className="px-6 py-4"><Skeleton className="h-4 w-48" /></td>
                    <td className="px-6 py-4"><Skeleton className="h-6 w-20 rounded-full" /></td>
                    <td className="px-6 py-4"><Skeleton className="h-4 w-24" /></td>
                    <td className="px-6 py-4"><Skeleton className="h-8 w-32 ml-auto" /></td>
                  </tr>
                ))
              ) : error ? (
                <tr>
                  <td colSpan={7} className="px-6 py-8 text-center">
                    <div className="text-red-600 dark:text-red-400">
                      <p className="font-medium">Failed to load VMs</p>
                      <button
                        onClick={() => refetch()}
                        className="mt-2 text-primary hover:text-primary/80 font-medium"
                      >
                        Try again
                      </button>
                    </div>
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
                paginatedVms.map((vm) => (
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
                      <div className="flex items-center justify-end gap-2" role="group" aria-label={`Actions for ${vm.name}`}>
                        {vm.status === "stopped" && (
                          <button
                            onClick={() => handleStart(vm.id)}
                            disabled={startMutation.isPending}
                            className="min-h-[44px] min-w-[44px] p-2 text-green-600 hover:text-green-900 dark:text-green-400 dark:hover:text-green-300 hover:bg-green-50 dark:hover:bg-green-900/20 rounded-md transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                            aria-label={`Start ${vm.name}`}
                            title="Start VM"
                          >
                            ▶
                          </button>
                        )}
                        {vm.status === "running" && (
                          <>
                            <button
                              onClick={() => handleStop(vm.id)}
                              disabled={stopMutation.isPending}
                              className="min-h-[44px] min-w-[44px] p-2 text-yellow-600 hover:text-yellow-900 dark:text-yellow-400 dark:hover:text-yellow-300 hover:bg-yellow-50 dark:hover:bg-yellow-900/20 rounded-md transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                              aria-label={`Stop ${vm.name}`}
                              title="Stop VM"
                            >
                              ⏹
                            </button>
                            <button
                              onClick={() => handleRestart(vm.id)}
                              disabled={restartMutation.isPending}
                              className="min-h-[44px] min-w-[44px] p-2 text-blue-600 hover:text-blue-900 dark:text-blue-400 dark:hover:text-blue-300 hover:bg-blue-50 dark:hover:bg-blue-900/20 rounded-md transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                              aria-label={`Restart ${vm.name}`}
                              title="Restart VM"
                            >
                              ↻
                            </button>
                          </>
                        )}
                        <button
                          onClick={() => handleDelete(vm.id)}
                          disabled={deleteMutation.isPending}
                          className="min-h-[44px] min-w-[44px] p-2 text-red-600 hover:text-red-900 dark:text-red-400 dark:hover:text-red-300 hover:bg-red-50 dark:hover:bg-red-900/20 rounded-md transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                          aria-label={`Delete ${vm.name}`}
                          title="Delete VM"
                        >
                          🗑
                        </button>
                      </div>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>

        {/* Pagination Controls */}
        {totalPages > 1 && (
          <div className="flex items-center justify-between mt-4 px-4 py-3 bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700">
            <div className="text-sm text-gray-600 dark:text-gray-400">
              Showing <span className="font-medium">{(currentPageSafe - 1) * UI.TABLE_PAGE_SIZE + 1}</span> to{" "}
              <span className="font-medium">{Math.min(currentPageSafe * UI.TABLE_PAGE_SIZE, filteredVms.length)}</span> of{" "}
              <span className="font-medium">{filteredVms.length}</span> VMs
            </div>
            <div className="flex items-center gap-2" role="navigation" aria-label="Pagination">
              <button
                onClick={() => setCurrentPage((p) => Math.max(1, p - 1))}
                disabled={currentPageSafe === 1}
                className="px-3 py-1 text-sm font-medium rounded-md border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-600 disabled:opacity-50 disabled:cursor-not-allowed min-h-[36px] min-w-[36px]"
                aria-label="Previous page"
              >
                ← Prev
              </button>
              <div className="flex items-center gap-1">
                {Array.from({ length: Math.min(5, totalPages) }, (_, i) => {
                  let pageNum: number;
                  if (totalPages <= 5) {
                    pageNum = i + 1;
                  } else if (currentPageSafe <= 3) {
                    pageNum = i + 1;
                  } else if (currentPageSafe >= totalPages - 2) {
                    pageNum = totalPages - 4 + i;
                  } else {
                    pageNum = currentPageSafe - 2 + i;
                  }

                  return (
                    <button
                      key={pageNum}
                      onClick={() => setCurrentPage(pageNum)}
                      aria-current={currentPageSafe === pageNum ? "page" : undefined}
                      className={`px-3 py-1 text-sm font-medium rounded-md min-h-[36px] min-w-[36px] ${
                        currentPageSafe === pageNum
                          ? "bg-primary text-white"
                          : "border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-600"
                      }`}
                    >
                      {pageNum}
                    </button>
                  );
                })}
              </div>
              <button
                onClick={() => setCurrentPage((p) => Math.min(totalPages, p + 1))}
                disabled={currentPageSafe === totalPages}
                className="px-3 py-1 text-sm font-medium rounded-md border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-600 disabled:opacity-50 disabled:cursor-not-allowed min-h-[36px] min-w-[36px]"
                aria-label="Next page"
              >
                Next →
              </button>
            </div>
          </div>
        )}
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
