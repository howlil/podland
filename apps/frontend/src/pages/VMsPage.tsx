// TODO: Refactor to use useVMs hook and VMTable component
import { useState, useMemo } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import api from "@/lib/api";
import { DashboardLayout } from "@/components/layout/DashboardLayout";
import { CreateVMWizard } from "@/components/vm/CreateVMWizard";
import { VMFilters } from "@/components/vm/VMFilters";
import { Pagination } from "@/components/ui/Pagination";
import { formatBytes, formatDate } from "@/lib/utils";
import { POLLING_INTERVALS, UI } from "@/lib/constants";
import { VMStatusBadge } from "@/components/vm/VMStatusBadge";
import { getErrorMessage } from "@/lib/errorHandler";

export default function VMsPage() {
  const [statusFilter, setStatusFilter] = useState<string>("all");
  const [isWizardOpen, setIsWizardOpen] = useState(false);
  const [currentPage, setCurrentPage] = useState(1);

  const { data: vms = [], isLoading, refetch } = useQuery<any[]>({
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
      toast.error(getErrorMessage(error, "Failed to start VM"));
    },
  });

  const stopMutation = useMutation({
    mutationFn: (vmId: string) => api.post(`/vms/${vmId}/stop`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["vms"] });
      toast.success("VM stopped successfully");
    },
    onError: (error: any) => {
      toast.error(getErrorMessage(error, "Failed to stop VM"));
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (vmId: string) => api.delete(`/vms/${vmId}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["vms"] });
      toast.success("VM deleted successfully");
    },
    onError: (error: any) => {
      toast.error(getErrorMessage(error, "Failed to delete VM"));
    },
  });

  // Filter
  const filteredVms = useMemo(() => {
    return vms.filter((vm) => {
      if (statusFilter === "all") return true;
      return vm.status === statusFilter;
    });
  }, [vms, statusFilter]);

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
        <VMFilters
          statusFilter={statusFilter}
          onStatusFilterChange={setStatusFilter}
          onCreateVM={() => setIsWizardOpen(true)}
        />

        {/* VM List */}
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
          <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
            <thead className="bg-gray-50 dark:bg-gray-700">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase">
                  Name
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase">
                  OS
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase">
                  Tier
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase">
                  Resources
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase">
                  Status
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase">
                  Created
                </th>
                <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-300 uppercase">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody className="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
              {isLoading ? (
                <tr>
                  <td colSpan={7} className="px-6 py-8 text-center text-gray-500">
                    Loading...
                  </td>
                </tr>
              ) : filteredVms.length === 0 ? (
                <tr>
                  <td colSpan={7} className="px-6 py-8 text-center text-gray-500">
                    No VMs found
                  </td>
                </tr>
              ) : (
                paginatedVms.map((vm) => (
                  <tr key={vm.id} className="hover:bg-gray-50 dark:hover:bg-gray-700/50">
                    <td className="px-6 py-4">
                      <a href={`/dashboard/vms/${vm.id}`} className="text-primary font-medium">
                        {vm.name}
                      </a>
                    </td>
                    <td className="px-6 py-4 text-gray-600 dark:text-gray-300">
                      {vm.os === "ubuntu-2204" ? "Ubuntu 22.04" : "Debian 12"}
                    </td>
                    <td className="px-6 py-4 text-gray-600 dark:text-gray-300">
                      {vm.tier}
                    </td>
                    <td className="px-6 py-4 text-gray-600 dark:text-gray-300">
                      {vm.cpu} CPU · {formatBytes(vm.ram)} RAM
                    </td>
                    <td className="px-6 py-4">
                      <VMStatusBadge status={vm.status} />
                    </td>
                    <td className="px-6 py-4 text-gray-600 dark:text-gray-300">
                      {formatDate(vm.created_at)}
                    </td>
                    <td className="px-6 py-4 text-right">
                      {vm.status === "stopped" && (
                        <button
                          onClick={() => startMutation.mutate(vm.id)}
                          className="text-green-600 hover:text-green-900 mr-3"
                        >
                          Start
                        </button>
                      )}
                      {vm.status === "running" && (
                        <>
                          <button
                            onClick={() => stopMutation.mutate(vm.id)}
                            className="text-yellow-600 hover:text-yellow-900 mr-3"
                          >
                            Stop
                          </button>
                          <button
                            onClick={() => deleteMutation.mutate(vm.id)}
                            className="text-red-600 hover:text-red-900"
                          >
                            Delete
                          </button>
                        </>
                      )}
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>

        <Pagination
          currentPage={currentPage}
          totalPages={totalPages}
          onPageChange={setCurrentPage}
          totalItems={filteredVms.length}
          itemsPerPage={UI.TABLE_PAGE_SIZE}
        />
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
