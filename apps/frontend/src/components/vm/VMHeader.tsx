import { VM } from "@/hooks/useVMs";
import { Pin, PinOff, ArrowLeft, Server } from "lucide-react";

interface VMHeaderProps {
  vm?: VM;
  isLoading?: boolean;
  onPin?: () => void;
  onUnpin?: () => void;
  isPinning?: boolean;
  isUnpinning?: boolean;
}

export function VMHeader({ vm, isLoading, onPin, onUnpin, isPinning, isUnpinning }: VMHeaderProps) {
  if (isLoading) {
    return (
      <div className="flex flex-col sm:flex-row justify-between items-start gap-4 mb-6 animate-pulse">
        <div className="space-y-2">
          <div className="h-8 w-48 bg-gray-200 dark:bg-gray-700 rounded" />
          <div className="h-4 w-32 bg-gray-200 dark:bg-gray-700 rounded" />
        </div>
        <div className="flex items-center gap-3">
          <div className="h-8 w-24 bg-gray-200 dark:bg-gray-700 rounded-full" />
          <div className="h-8 w-20 bg-gray-200 dark:bg-gray-700 rounded-full" />
        </div>
      </div>
    );
  }

  if (!vm) return null;

  const getStatusColor = (status: string) => {
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

  return (
    <div className="flex flex-col sm:flex-row justify-between items-start gap-4 mb-6">
      <div>
        <div className="flex items-center gap-3">
          <a
            href="/dashboard/vms"
            className="inline-flex items-center text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200 transition-colors"
          >
            <ArrowLeft className="h-4 w-4 mr-1" />
            Back
          </a>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-white">{vm.name}</h1>
        </div>
        <p className="text-gray-600 dark:text-gray-400 mt-2 ml-12 flex items-center gap-2">
          <span className="flex items-center gap-1.5">
            <Server className="h-4 w-4" />
            {vm.os === "ubuntu-2204" ? "Ubuntu 22.04 LTS" : "Debian 12"}
          </span>
          <span className="text-gray-300 dark:text-gray-600">•</span>
          <span className="capitalize">{vm.tier} tier</span>
        </p>
      </div>
      <div className="flex items-center gap-3 flex-wrap">
        {vm.is_pinned ? (
          <>
            <span className="inline-flex items-center gap-1.5 px-3 py-1.5 bg-yellow-100 dark:bg-yellow-900/30 text-yellow-800 dark:text-yellow-400 rounded-full text-sm font-semibold">
              <Pin className="h-4 w-4 fill-current" />
              Pinned
            </span>
            <button
              onClick={onUnpin}
              disabled={isUnpinning}
              className="inline-flex items-center gap-1.5 px-3 py-1.5 bg-gray-100 hover:bg-gray-200 dark:bg-gray-700 dark:hover:bg-gray-600 text-gray-700 dark:text-gray-300 rounded-lg text-sm font-medium transition-all disabled:opacity-50"
            >
              <PinOff className="h-4 w-4" />
              {isUnpinning ? "..." : "Unpin"}
            </button>
          </>
        ) : (
          <button
            onClick={onPin}
            disabled={isPinning}
            className="inline-flex items-center gap-1.5 px-3 py-1.5 bg-gray-100 hover:bg-gray-200 dark:bg-gray-700 dark:hover:bg-gray-600 text-gray-700 dark:text-gray-300 rounded-lg text-sm font-medium transition-all disabled:opacity-50"
            title="Pin VM to prevent auto-deletion"
          >
            <Pin className="h-4 w-4" />
            {isPinning ? "..." : "Pin VM"}
          </button>
        )}
        <span className={`inline-flex items-center gap-2 px-4 py-1.5 text-sm font-semibold rounded-full ${getStatusColor(vm.status)}`}>
          {vm.status.toUpperCase()}
        </span>
      </div>
    </div>
  );
}
