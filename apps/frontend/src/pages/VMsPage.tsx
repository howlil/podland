// TODO: Refactor to use useVMs hook and VMTable component
// For now, keeping existing implementation to maintain functionality
import { useState, useEffect } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import api from "@/lib/api";
import { DashboardLayout } from "@/components/layout/DashboardLayout";
import { CreateVMWizard } from "@/components/vm/CreateVMWizard";
import { Skeleton } from "@/components/ui/skeleton";
import { formatBytes } from "@/lib/utils";
import { POLLING_INTERVALS, VM_STATUS, UI } from "@/lib/constants";

export default function VMsPage() {
  const [statusFilter, setStatusFilter] = useState<string>("all");
  const [sortField, setSortField] = useState<string>("created_at");
  const [sortOrder, setSortOrder] = useState<"asc" | "desc">("desc");
  const [isWizardOpen, setIsWizardOpen] = useState(false);
  const [currentPage, setCurrentPage] = useState(1);

  const { data: vms = [], isLoading, error, refetch } = useQuery<any[]>({
    queryKey: ["vms"],
    queryFn: async () => {
      const { data } = await api.get("/vms");
      return data;
    },
    refetchInterval: POLLING_INTERVALS.VM_STATUS,
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

  // Filter and sort
  let filteredVms = vms.filter((vm) => {
    if (statusFilter === "all") return true;
    return vm.status === statusFilter;
  });

  filteredVms.sort((a, b) => {
    let aVal = a[sortField];
    let bVal = b[sortField];
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

  return (
    <DashboardLayout>
      <div className="max-w-7xl mx-auto">
        <div className="flex justify-between items-center mb-6">
          <div>
            <h1 className="text-2xl font-bold text-gray-900 dark:text-white">My VMs</h1>
            <p className="text-gray-600 dark:text-gray-400 mt-1">Manage your virtual machines</p>
          </div>
          <button
            onClick={() => setIsWizardOpen(true)}
            className="px-6 py-3 bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700 text-white rounded-xl font-semibold shadow-lg transition-all"
          >
            <span className="inline-flex items-center gap-2">
              <span>+</span> Create VM
            </span>
          </button>
        </div>

        {/* Filters */}
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-4 mb-6">
          <div className="flex flex-wrap gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Status</label>
              <select
                value={statusFilter}
                onChange={(e) => setStatusFilter(e.target.value)}
                className="px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
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

        {/* VM List */}
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
          <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
            <thead className="bg-gray-50 dark:bg-gray-700">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase">Name</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase">OS</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase">Tier</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase">Resources</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase">Status</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase">Created</th>
                <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-300 uppercase">Actions</th>
              </tr>
            </thead>
            <tbody className="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
              {isLoading ? (
                <tr><td colSpan={7} className="px-6 py-8 text-center text-gray-500">Loading...</td></tr>
              ) : filteredVms.length === 0 ? (
                <tr><td colSpan={7} className="px-6 py-8 text-center text-gray-500">No VMs found</td></tr>
              ) : (
                paginatedVms.map((vm) => (
                  <tr key={vm.id} className="hover:bg-gray-50 dark:hover:bg-gray-700/50">
                    <td className="px-6 py-4">
                      <a href={`/dashboard/vms/${vm.id}`} className="text-primary font-medium">{vm.name}</a>
                    </td>
                    <td className="px-6 py-4 text-gray-600 dark:text-gray-300">{vm.os === "ubuntu-2204" ? "Ubuntu 22.04" : "Debian 12"}</td>
                    <td className="px-6 py-4 text-gray-600 dark:text-gray-300">{vm.tier}</td>
                    <td className="px-6 py-4 text-gray-600 dark:text-gray-300">{vm.cpu} CPU · {formatBytes(vm.ram)} RAM</td>
                    <td className="px-6 py-4">
                      <span className={`px-2 py-1 rounded-full text-xs font-semibold ${
                        vm.status === "running" ? "bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400" :
                        vm.status === "stopped" ? "bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300" :
                        vm.status === "pending" ? "bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400" :
                        "bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400"
                      }`}>{vm.status}</span>
                    </td>
                    <td className="px-6 py-4 text-gray-600 dark:text-gray-300">{new Date(vm.created_at).toLocaleDateString()}</td>
                    <td className="px-6 py-4 text-right">
                      {vm.status === "stopped" && (
                        <button onClick={() => startMutation.mutate(vm.id)} className="text-green-600 hover:text-green-900 mr-3">Start</button>
                      )}
                      {vm.status === "running" && (
                        <>
                          <button onClick={() => stopMutation.mutate(vm.id)} className="text-yellow-600 hover:text-yellow-900 mr-3">Stop</button>
                          <button onClick={() => deleteMutation.mutate(vm.id)} className="text-red-600 hover:text-red-900">Delete</button>
                        </>
                      )}
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>

        {/* Pagination */}
        {totalPages > 1 && (
          <div className="flex justify-center gap-2 mt-6">
            <button onClick={() => setCurrentPage(p => Math.max(1, p - 1))} disabled={currentPageSafe === 1} className="px-4 py-2 border rounded-lg disabled:opacity-50">Previous</button>
            <span className="px-4 py-2">Page {currentPageSafe} of {totalPages}</span>
            <button onClick={() => setCurrentPage(p => Math.min(totalPages, p + 1))} disabled={currentPageSafe === totalPages} className="px-4 py-2 border rounded-lg disabled:opacity-50">Next</button>
          </div>
        )}
      </div>

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
