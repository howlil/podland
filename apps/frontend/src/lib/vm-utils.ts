import { VM } from "@/hooks/useVMs";

/**
 * Get CSS classes for VM status badge
 * @param status - VM status
 * @returns Tailwind CSS classes for status styling
 */
export function getVMStatusStyles(status: VM["status"]): string {
  const styles: Record<VM["status"], string> = {
    running:
      "bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400",
    stopped:
      "bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300",
    pending:
      "bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400",
    error:
      "bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400",
  };
  return styles[status] || styles.stopped;
}

/**
 * Get status label for display
 * @param status - VM status
 * @returns Human-readable status label
 */
export function getVMStatusLabel(status: VM["status"]): string {
  const labels: Record<VM["status"], string> = {
    running: "Running",
    stopped: "Stopped",
    pending: "Pending",
    error: "Error",
  };
  return labels[status] || status;
}

/**
 * Check if VM is in active state (can perform actions)
 * @param status - VM status
 * @returns True if VM is running
 */
export function isVMActive(status: VM["status"]): boolean {
  return status === "running";
}

/**
 * Check if VM can be started
 * @param status - VM status
 * @returns True if VM can be started
 */
export function canStartVM(status: VM["status"]): boolean {
  return status === "stopped";
}

/**
 * Check if VM can be stopped
 * @param status - VM status
 * @returns True if VM can be stopped
 */
export function canStopVM(status: VM["status"]): boolean {
  return status === "running";
}
