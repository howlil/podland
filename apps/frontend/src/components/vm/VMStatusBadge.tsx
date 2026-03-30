import { VM } from "@/hooks/useVMs";
import { getVMStatusStyles, getVMStatusLabel } from "@/lib/vm-utils";

/**
 * VMStatusBadge component props
 */
export interface VMStatusBadgeProps {
  /** VM status */
  status: VM["status"];
  /** Show icon dot */
  showDot?: boolean;
  /** Size variant */
  size?: "sm" | "md" | "lg";
  /** Additional CSS classes */
  className?: string;
}

/**
 * VMStatusBadge component for consistent status display
 * Replaces duplicate status badge logic across the app
 */
export function VMStatusBadge({
  status,
  showDot = true,
  size = "sm",
  className = "",
}: VMStatusBadgeProps) {
  const sizeClasses = {
    sm: "text-xs px-2 py-1",
    md: "text-sm px-3 py-1.5",
    lg: "text-base px-4 py-2",
  };

  const dotSizeClasses = {
    sm: "h-1.5 w-1.5",
    md: "h-2 w-2",
    lg: "h-2.5 w-2.5",
  };

  return (
    <span
      className={`inline-flex items-center gap-1.5 rounded-full font-semibold ${getVMStatusStyles(status)} ${sizeClasses[size]} ${className}`}
    >
      {showDot && (
        <span className={`rounded-full bg-current ${dotSizeClasses[size]}`} />
      )}
      {getVMStatusLabel(status)}
    </span>
  );
}
