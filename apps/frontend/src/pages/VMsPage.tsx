import { useState } from "react";
import { useVMs } from "@/hooks/useVMs";
import { useUIStore } from "@/stores/uiStore";
import { VMTable } from "@/components/vm/VMTable";
import { StatsCard } from "@/components/ui/StatsCard";
import { EmptyState } from "@/components/ui/StatsCard";
import { CreateVMWizard } from "@/components/vm/CreateVMWizard";
import { Server, Zap, HardDrive, Plus } from "lucide-react";
import { UI } from "@/lib/constants";

export default function VMsPage() {
  const [isWizardOpen, setIsWizardOpen] = useState(false);

  // Use custom hook for VM data and actions
  const {
    vms,
    isLoading,
    startVM,
    stopVM,
    deleteVM,
  } = useVMs();

  // Use Zustand for UI state
  const {
    vmStatusFilter,
    vmCurrentPage,
    setVMStatusFilter,
    setVMCurrentPage,
  } = useUIStore();

  // Filter and sort VMs
  let filteredVms = vms.filter((vm) => {
    if (vmStatusFilter === "all") return true;
    return vm.status === vmStatusFilter;
  });

  // Pagination
  const totalPages = Math.max(1, Math.ceil(filteredVms.length / UI.TABLE_PAGE_SIZE));
  const currentPageSafe = Math.min(vmCurrentPage, totalPages);
  const paginatedVms = filteredVms.slice(
    (currentPageSafe - 1) * UI.TABLE_PAGE_SIZE,
    currentPageSafe * UI.TABLE_PAGE_SIZE
  );

  const runningVMs = vms.filter((vm) => vm.status === "running").length;
  const totalRAM = vms.reduce((sum, vm) => sum + vm.ram, 0);

  return (
    <div className="max-w-7xl mx-auto space-y-6">
      {/* Header */}
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-white">My VMs</h1>
          <p className="text-gray-600 dark:text-gray-400 mt-1">Manage your virtual machines</p>
        </div>
        <button
          onClick={() => setIsWizardOpen(true)}
          className="inline-flex items-center gap-2 px-6 py-3 bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700 text-white rounded-xl font-semibold shadow-lg hover:shadow-xl transition-all"
        >
          <Plus className="h-5 w-5" />
          Create VM
        </button>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <StatsCard icon={Server} label="Total VMs" value={vms.length} color="from-blue-500 to-cyan-500" />
        <StatsCard icon={Zap} label="Running" value={runningVMs} color="from-green-500 to-emerald-500" />
        <StatsCard icon={HardDrive} label="Total RAM" value={formatBytes(totalRAM)} color="from-purple-500 to-pink-500" />
      </div>

      {/* Filters */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-4">
        <div className="flex flex-wrap gap-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Status</label>
            <select
              value={vmStatusFilter}
              onChange={(e) => setVMStatusFilter(e.target.value)}
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

      {/* VM List */}
      {filteredVms.length === 0 ? (
        <EmptyState
          icon={Server}
          title={vmStatusFilter === "all" ? "No VMs yet" : `No ${vmStatusFilter} VMs`}
          description={vmStatusFilter === "all" ? "Create your first VM to get started!" : "Try changing the filter"}
          action={
            vmStatusFilter === "all" && (
              <button
                onClick={() => setIsWizardOpen(true)}
                className="inline-flex items-center gap-2 px-6 py-3 bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700 text-white rounded-xl font-semibold"
              >
                <Plus className="h-5 w-5" />
                Create VM
              </button>
            )
          }
        />
      ) : (
        <>
          <VMTable
            vms={paginatedVms}
            isLoading={isLoading}
            onStart={startVM}
            onStop={stopVM}
            onDelete={deleteVM}
          />

          {/* Pagination */}
          {totalPages > 1 && (
            <Pagination
              currentPage={currentPageSafe}
              totalPages={totalPages}
              onPageChange={setVMCurrentPage}
            />
          )}
        </>
      )}

      {isWizardOpen && (
        <CreateVMWizard
          onClose={() => setIsWizardOpen(false)}
          onSuccess={() => setIsWizardOpen(false)}
        />
      )}
    </div>
  );
}

function Pagination({ currentPage, totalPages, onPageChange }: { currentPage: number; totalPages: number; onPageChange: (page: number) => void }) {
  return (
    <div className="flex justify-center gap-2 mt-6">
      <button
        onClick={() => onPageChange(currentPage - 1)}
        disabled={currentPage === 1}
        className="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg disabled:opacity-50"
      >
        Previous
      </button>
      <span className="px-4 py-2">Page {currentPage} of {totalPages}</span>
      <button
        onClick={() => onPageChange(currentPage + 1)}
        disabled={currentPage === totalPages}
        className="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg disabled:opacity-50"
      >
        Next
      </button>
    </div>
  );
}

function formatBytes(bytes: number) {
  const k = 1024;
  const sizes = ["B", "KB", "MB", "GB", "TB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + " " + sizes[i];
}
