import type { LucideIcon } from "lucide-react";
import { Play, RotateCcw, Trash2, Power } from "lucide-react";
import { VM } from "@/hooks/useVMs";

/**
 * VM Action configuration interface
 */
export interface VMActionConfig {
  id: string;
  label: string;
  icon: LucideIcon;
  variant: "default" | "destructive" | "outline" | "secondary";
  isVisible: (vm: VM) => boolean;
  isEnabled: (vm: VM, isLoading?: boolean) => boolean;
  onClick: (vm: VM) => void;
  isLoading?: boolean;
  loadingText?: string;
}

/**
 * Default VM actions configuration
 */
export function getDefaultVMActions(options: {
  onStart?: (vm: VM) => void;
  onStop?: (vm: VM) => void;
  onRestart?: (vm: VM) => void;
  onDelete?: (vm: VM) => void;
  isStarting?: boolean;
  isStopping?: boolean;
  isRestarting?: boolean;
  isDeleting?: boolean;
}): VMActionConfig[] {
  const { onStart, onStop, onRestart, onDelete, isStarting, isStopping, isRestarting, isDeleting } = options;

  const actions: VMActionConfig[] = [];

  // Start action
  if (onStart) {
    actions.push({
      id: "start",
      label: "Start VM",
      icon: Play,
      variant: "default",
      isVisible: (vm) => vm.status === "stopped",
      isEnabled: () => !isStarting,
      onClick: onStart,
      isLoading: isStarting,
      loadingText: "Starting...",
    });
  }

  // Stop action
  if (onStop) {
    actions.push({
      id: "stop",
      label: "Stop VM",
      icon: Power,
      variant: "secondary",
      isVisible: (vm) => vm.status === "running",
      isEnabled: () => !isStopping,
      onClick: onStop,
      isLoading: isStopping,
      loadingText: "Stopping...",
    });
  }

  // Restart action
  if (onRestart) {
    actions.push({
      id: "restart",
      label: "Restart VM",
      icon: RotateCcw,
      variant: "outline",
      isVisible: (vm) => vm.status === "running",
      isEnabled: () => !isRestarting,
      onClick: onRestart,
      isLoading: isRestarting,
      loadingText: "Restarting...",
    });
  }

  // Delete action
  if (onDelete) {
    actions.push({
      id: "delete",
      label: "Delete VM",
      icon: Trash2,
      variant: "destructive",
      isVisible: () => true,
      isEnabled: () => !isDeleting,
      onClick: onDelete,
      isLoading: isDeleting,
      loadingText: "Deleting...",
    });
  }

  return actions;
}

/**
 * VMActions component props
 */
export interface VMActionsProps {
  vm?: VM;
  isLoading?: boolean;
  actions?: VMActionConfig[];
  className?: string;
}

/**
 * VMActions component with configurable actions
 */
export function VMActions({ vm, isLoading, actions, className = "" }: VMActionsProps) {
  // Use default actions if none provided
  const vmActions = actions || [];

  if (isLoading || !vm) {
    return (
      <div className={`bg-white dark:bg-gray-800 rounded-2xl shadow-sm border border-gray-200 dark:border-gray-700 p-6 animate-pulse ${className}`}>
        <div className="h-6 w-32 bg-gray-200 dark:bg-gray-700 rounded mb-4" />
        <div className="flex flex-wrap gap-3">
          {Array.from({ length: 3 }).map((_, i) => (
            <div key={i} className="h-10 w-28 bg-gray-200 dark:bg-gray-700 rounded-xl" />
          ))}
        </div>
      </div>
    );
  }

  // Filter visible actions
  const visibleActions = vmActions.filter((action) => action.isVisible(vm));

  if (visibleActions.length === 0) {
    return null;
  }

  return (
    <div className={`bg-white dark:bg-gray-800 rounded-2xl shadow-sm border border-gray-200 dark:border-gray-700 p-6 ${className}`}>
      <h2 className="text-lg font-semibold text-gray-900 dark:text-white mb-4 flex items-center gap-2">
        VM Actions
      </h2>
      <div className="flex flex-wrap gap-3">
        {visibleActions.map((action) => {
          const Icon = action.icon;
          const enabled = action.isEnabled(vm, action.isLoading);
          
          return (
            <button
              key={action.id}
              onClick={() => action.onClick(vm)}
              disabled={!enabled}
              className={`inline-flex items-center gap-2 px-5 py-2.5 rounded-xl font-medium transition-all shadow-md hover:shadow-lg disabled:opacity-50 disabled:cursor-not-allowed
                ${action.variant === "default" 
                  ? "bg-gradient-to-r from-green-600 to-green-700 hover:from-green-700 hover:to-green-800 text-white"
                  : action.variant === "destructive"
                  ? "bg-gradient-to-r from-red-600 to-red-700 hover:from-red-700 hover:to-red-800 text-white"
                  : action.variant === "secondary"
                  ? "bg-gradient-to-r from-yellow-600 to-yellow-700 hover:from-yellow-700 hover:to-yellow-800 text-white"
                  : "bg-gradient-to-r from-blue-600 to-blue-700 hover:from-blue-700 hover:to-blue-800 text-white"
                }
              `}
            >
              <Icon className="h-4 w-4" />
              {action.isLoading ? (action.loadingText || action.label) : action.label}
            </button>
          );
        })}
      </div>
    </div>
  );
}
