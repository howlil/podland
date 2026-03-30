import { VM } from "@/hooks/useVMs";
import { Play, Square, RotateCcw, Trash2 } from "lucide-react";

interface VMActionsProps {
  vm?: VM;
  isLoading?: boolean;
  onStart?: () => void;
  onStop?: () => void;
  onRestart?: () => void;
  onDelete?: () => void;
  isStarting?: boolean;
  isStopping?: boolean;
  isRestarting?: boolean;
  isDeleting?: boolean;
}

export function VMActions({ vm, isLoading, onStart, onStop, onRestart, onDelete, isStarting, isStopping, isRestarting, isDeleting }: VMActionsProps) {
  if (isLoading || !vm) {
    return (
      <div className="bg-white dark:bg-gray-800 rounded-2xl shadow-sm border border-gray-200 dark:border-gray-700 p-6 animate-pulse">
        <div className="h-6 w-32 bg-gray-200 dark:bg-gray-700 rounded mb-4" />
        <div className="flex flex-wrap gap-3">
          {Array.from({ length: 3 }).map((_, i) => (
            <div key={i} className="h-10 w-28 bg-gray-200 dark:bg-gray-700 rounded-xl" />
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className="bg-white dark:bg-gray-800 rounded-2xl shadow-sm border border-gray-200 dark:border-gray-700 p-6">
      <h2 className="text-lg font-semibold text-gray-900 dark:text-white mb-4 flex items-center gap-2">
        VM Actions
      </h2>
      <div className="flex flex-wrap gap-3">
        {vm.status === "stopped" && (
          <button
            onClick={onStart}
            disabled={isStarting}
            className="inline-flex items-center gap-2 px-5 py-2.5 bg-gradient-to-r from-green-600 to-green-700 hover:from-green-700 hover:to-green-800 text-white rounded-xl font-medium transition-all disabled:opacity-50 shadow-md hover:shadow-lg"
          >
            <Play className="h-4 w-4" />
            {isStarting ? "Starting..." : "Start VM"}
          </button>
        )}
        {vm.status === "running" && (
          <>
            <button
              onClick={onStop}
              disabled={isStopping}
              className="inline-flex items-center gap-2 px-5 py-2.5 bg-gradient-to-r from-yellow-600 to-yellow-700 hover:from-yellow-700 hover:to-yellow-800 text-white rounded-xl font-medium transition-all disabled:opacity-50 shadow-md hover:shadow-lg"
            >
              <Square className="h-4 w-4" />
              {isStopping ? "Stopping..." : "Stop VM"}
            </button>
            <button
              onClick={onRestart}
              disabled={isRestarting}
              className="inline-flex items-center gap-2 px-5 py-2.5 bg-gradient-to-r from-blue-600 to-blue-700 hover:from-blue-700 hover:to-blue-800 text-white rounded-xl font-medium transition-all disabled:opacity-50 shadow-md hover:shadow-lg"
            >
              <RotateCcw className="h-4 w-4" />
              {isRestarting ? "Restarting..." : "Restart VM"}
            </button>
          </>
        )}
        <button
          onClick={onDelete}
          disabled={isDeleting}
          className="inline-flex items-center gap-2 px-5 py-2.5 bg-gradient-to-r from-red-600 to-red-700 hover:from-red-700 hover:to-red-800 text-white rounded-xl font-medium transition-all disabled:opacity-50 shadow-md hover:shadow-lg"
        >
          <Trash2 className="h-4 w-4" />
          {isDeleting ? "Deleting..." : "Delete VM"}
        </button>
      </div>
    </div>
  );
}
