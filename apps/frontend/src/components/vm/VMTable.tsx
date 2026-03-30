import { VM } from "@/hooks/useVMs";
import { formatBytes } from "@/lib/utils";
import { Server, Trash2, Play, Square, RotateCcw } from "lucide-react";
import { VMStatusBadge } from "@/components/vm/VMStatusBadge";
import { canStartVM, canStopVM } from "@/lib/vm-utils";

interface VMTableProps {
  vms: VM[];
  isLoading: boolean;
  onStart?: (vmId: string) => void;
  onStop?: (vmId: string) => void;
  onRestart?: (vmId: string) => void;
  onDelete?: (vmId: string) => void;
}

export function VMTable({
  vms,
  isLoading,
  onStart,
  onStop,
  onRestart,
  onDelete,
}: VMTableProps) {
  if (isLoading) {
    return (
      <div className="space-y-3">
        {Array.from({ length: 5 }).map((_, i) => (
          <div key={i} className="animate-pulse flex items-center justify-between p-4 bg-gray-100 dark:bg-gray-800 rounded-xl">
            <div className="flex items-center gap-4">
              <div className="h-10 w-10 bg-gray-200 dark:bg-gray-700 rounded-lg" />
              <div className="space-y-2">
                <div className="h-4 w-32 bg-gray-200 dark:bg-gray-700 rounded" />
                <div className="h-3 w-24 bg-gray-200 dark:bg-gray-700 rounded" />
              </div>
            </div>
            <div className="h-8 w-24 bg-gray-200 dark:bg-gray-700 rounded" />
          </div>
        ))}
      </div>
    );
  }

  if (vms.length === 0) {
    return (
      <div className="text-center py-12">
        <Server className="h-16 w-16 mx-auto mb-4 text-gray-400 dark:text-gray-600" />
        <p className="text-gray-600 dark:text-gray-400">No VMs found</p>
      </div>
    );
  }

  return (
    <div className="space-y-3">
      {vms.map((vm) => (
        <VMRow
          key={vm.id}
          vm={vm}
          onStart={onStart}
          onStop={onStop}
          onRestart={onRestart}
          onDelete={onDelete}
        />
      ))}
    </div>
  );
}

interface VMRowProps {
  vm: VM;
  onStart?: (vmId: string) => void;
  onStop?: (vmId: string) => void;
  onRestart?: (vmId: string) => void;
  onDelete?: (vmId: string) => void;
}

function VMRow({ vm, onStart, onStop, onRestart, onDelete }: VMRowProps) {
  return (
    <a
      href={`/dashboard/vms/${vm.id}`}
      className="group flex items-center justify-between p-4 bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 hover:border-blue-300 dark:hover:border-blue-700 transition-all"
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
            {vm.os === "ubuntu-2204" ? "Ubuntu 22.04" : "Debian 12"} · {vm.tier} · {formatBytes(vm.ram)} RAM
          </p>
        </div>
      </div>
      <div className="flex items-center gap-3">
        <VMStatusBadge status={vm.status} />
        <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
          {canStartVM(vm.status) && (
            <button onClick={(e) => { e.preventDefault(); onStart?.(vm.id); }} className="p-2 hover:bg-green-50 dark:hover:bg-green-900/20 rounded-lg" title="Start">
              <Play className="h-4 w-4 text-green-600" />
            </button>
          )}
          {canStopVM(vm.status) && (
            <>
              <button onClick={(e) => { e.preventDefault(); onStop?.(vm.id); }} className="p-2 hover:bg-yellow-50 dark:hover:bg-yellow-900/20 rounded-lg" title="Stop">
                <Square className="h-4 w-4 text-yellow-600" />
              </button>
              <button onClick={(e) => { e.preventDefault(); onRestart?.(vm.id); }} className="p-2 hover:bg-blue-50 dark:hover:bg-blue-900/20 rounded-lg" title="Restart">
                <RotateCcw className="h-4 w-4 text-blue-600" />
              </button>
            </>
          )}
          <button onClick={(e) => { e.preventDefault(); onDelete?.(vm.id); }} className="p-2 hover:bg-red-50 dark:hover:bg-red-900/20 rounded-lg" title="Delete">
            <Trash2 className="h-4 w-4 text-red-600" />
          </button>
        </div>
      </div>
    </a>
  );
}
