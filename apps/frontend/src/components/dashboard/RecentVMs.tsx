import { Server, ArrowRight, Plus } from "lucide-react";
import { VM } from "@/hooks/useVMs";

interface RecentVMsProps {
  vms?: VM[];
  isLoading?: boolean;
  onCreateVM?: () => void;
}

export function RecentVMs({ vms, isLoading, onCreateVM }: RecentVMsProps) {
  if (isLoading) {
    return (
      <div className="bg-white dark:bg-gray-800 rounded-2xl shadow-sm border border-gray-200 dark:border-gray-700 p-6 animate-pulse">
        <div className="h-6 w-32 bg-gray-200 dark:bg-gray-700 rounded mb-6" />
        <div className="space-y-3">
          {Array.from({ length: 3 }).map((_, i) => (
            <div key={i} className="flex items-center gap-4">
              <div className="h-10 w-10 bg-gray-200 dark:bg-gray-700 rounded-lg" />
              <div className="flex-1 space-y-2">
                <div className="h-4 w-32 bg-gray-200 dark:bg-gray-700 rounded" />
                <div className="h-3 w-24 bg-gray-200 dark:bg-gray-700 rounded" />
              </div>
              <div className="h-4 w-4 bg-gray-200 dark:bg-gray-700 rounded" />
            </div>
          ))}
        </div>
      </div>
    );
  }

  if (!vms || vms.length === 0) {
    return (
      <div className="bg-white dark:bg-gray-800 rounded-2xl shadow-sm border border-gray-200 dark:border-gray-700 p-6 text-center">
        <Server className="h-16 w-16 mx-auto mb-4 text-gray-400 dark:text-gray-600" />
        <p className="text-lg font-medium text-gray-900 dark:text-white mb-2">
          No VMs yet
        </p>
        <p className="text-gray-600 dark:text-gray-400 mb-6">
          Create your first VM to get started!
        </p>
        {onCreateVM && (
          <button
            onClick={onCreateVM}
            className="inline-flex items-center gap-2 px-6 py-3 bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700 text-white rounded-xl font-semibold shadow-lg hover:shadow-xl transition-all"
          >
            <Plus className="h-5 w-5" />
            Create Your First VM
          </button>
        )}
      </div>
    );
  }

  return (
    <div className="bg-white dark:bg-gray-800 rounded-2xl shadow-sm border border-gray-200 dark:border-gray-700 p-6">
      <div className="flex justify-between items-center mb-4">
        <h2 className="text-lg font-semibold text-gray-900 dark:text-white">
          Recent VMs
        </h2>
        <a
          href="/dashboard/vms"
          className="inline-flex items-center gap-1 text-sm text-blue-600 hover:text-blue-700 dark:text-blue-400 font-medium"
        >
          View All
          <ArrowRight className="h-4 w-4" />
        </a>
      </div>
      <div className="space-y-3">
        {vms.slice(0, 5).map((vm) => (
          <VMRow key={vm.id} vm={vm} />
        ))}
      </div>
    </div>
  );
}

interface VMRowProps {
  vm: VM;
}

function VMRow({ vm }: VMRowProps) {
  const getStatusColor = (status: string) => {
    switch (status) {
      case "running":
        return "bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400";
      case "stopped":
        return "bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300";
      case "pending":
        return "bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400";
      default:
        return "bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300";
    }
  };

  return (
    <a
      href={`/dashboard/vms/${vm.id}`}
      className="flex items-center justify-between p-4 bg-gray-50 dark:bg-gray-700/50 rounded-xl hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors group"
    >
      <div className="flex items-center gap-4">
        <div className="p-2 bg-blue-100 dark:bg-blue-900/30 rounded-lg">
          <Server className="h-5 w-5 text-blue-600 dark:text-blue-400" />
        </div>
        <div>
          <p className="font-medium text-gray-900 dark:text-white group-hover:text-blue-600 dark:group-hover:text-blue-400">
            {vm.name}
          </p>
          <p className="text-sm text-gray-600 dark:text-gray-400">
            {vm.os === "ubuntu-2204" ? "Ubuntu 22.04" : "Debian 12"}
          </p>
        </div>
      </div>
      <div className="flex items-center gap-4">
        <span
          className={`px-3 py-1 rounded-full text-xs font-semibold ${getStatusColor(
            vm.status
          )}`}
        >
          {vm.status}
        </span>
        <ArrowRight className="h-4 w-4 text-gray-400 group-hover:text-blue-600 dark:group-hover:text-blue-400" />
      </div>
    </a>
  );
}
