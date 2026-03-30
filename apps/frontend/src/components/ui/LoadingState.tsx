import * as React from "react";
import { Loader2 } from "lucide-react";

/**
 * LoadingState component props
 */
export interface LoadingStateProps {
  /** Loading message */
  message?: string;
  /** Size of the spinner */
  size?: "sm" | "md" | "lg";
  /** Additional CSS classes */
  className?: string;
  /** Show full screen overlay */
  fullscreen?: boolean;
}

/**
 * Standardized loading state component
 * Replaces ad-hoc animate-pulse patterns throughout the app
 */
export function LoadingState({
  message = "Loading...",
  size = "md",
  className = "",
  fullscreen = false,
}: LoadingStateProps) {
  const sizeClasses = {
    sm: "h-4 w-4",
    md: "h-8 w-8",
    lg: "h-12 w-12",
  };

  const content = (
    <div className="flex flex-col items-center justify-center gap-3">
      <Loader2 className={`${sizeClasses[size]} animate-spin text-blue-600 dark:text-blue-400`} />
      {message && (
        <p className="text-sm text-gray-600 dark:text-gray-400">{message}</p>
      )}
    </div>
  );

  if (fullscreen) {
    return (
      <div className={`fixed inset-0 bg-white/80 dark:bg-gray-900/80 backdrop-blur-sm flex items-center justify-center ${className}`}>
        {content}
      </div>
    );
  }

  return (
    <div className={`flex items-center justify-center py-12 ${className}`}>
      {content}
    </div>
  );
}

/**
 * CardLoading component for loading states inside cards
 */
export interface CardLoadingProps {
  /** Number of skeleton lines */
  lines?: number;
  /** Show header skeleton */
  showHeader?: boolean;
  /** Additional CSS classes */
  className?: string;
}

export function CardLoading({
  lines = 3,
  showHeader = true,
  className = "",
}: CardLoadingProps) {
  return (
    <div className={`bg-white dark:bg-gray-800 rounded-2xl shadow-sm border border-gray-200 dark:border-gray-700 p-6 animate-pulse ${className}`}>
      {showHeader && (
        <div className="h-6 w-32 bg-gray-200 dark:bg-gray-700 rounded mb-4" />
      )}
      <div className="space-y-3">
        {Array.from({ length: lines }).map((_, i) => (
          <div
            key={i}
            className="h-4 bg-gray-200 dark:bg-gray-700 rounded"
            style={{ width: `${100 - (i * 10)}%` }}
          />
        ))}
      </div>
    </div>
  );
}

/**
 * TableLoading component for loading states in tables
 */
export interface TableLoadingProps {
  /** Number of columns */
  columns?: number;
  /** Number of loading rows */
  rows?: number;
  /** Additional CSS classes */
  className?: string;
}

export function TableLoading({
  columns = 7,
  rows = 5,
  className = "",
}: TableLoadingProps) {
  return (
    <tbody className={`bg-white dark:bg-gray-800 animate-pulse ${className}`}>
      {Array.from({ length: rows }).map((_, rowIndex) => (
        <tr key={rowIndex} className="border-b border-gray-200 dark:border-gray-700">
          {Array.from({ length: columns }).map((_, colIndex) => (
            <td key={colIndex} className="px-6 py-4">
              <div className="h-4 bg-gray-200 dark:bg-gray-700 rounded" />
            </td>
          ))}
        </tr>
      ))}
    </tbody>
  );
}

/**
 * EmptyState component for empty data states
 * @deprecated Use @/components/ui/EmptyState instead
 */
export interface EmptyStateProps {
  /** Icon to display */
  icon?: React.ReactNode;
  /** Title text */
  title: string;
  /** Description text */
  description?: string;
  /** Action button */
  action?: React.ReactNode;
  /** Additional CSS classes */
  className?: string;
}

export function EmptyState({
  icon,
  title,
  description,
  action,
  className = "",
}: EmptyStateProps) {
  return (
    <div className={`bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-12 text-center ${className}`}>
      {icon && (
        <div className="mb-4 flex justify-center text-gray-400 dark:text-gray-600">
          {icon}
        </div>
      )}
      <h3 className="text-lg font-medium text-gray-900 dark:text-white mb-2">
        {title}
      </h3>
      {description && (
        <p className="text-gray-600 dark:text-gray-400 mb-4">{description}</p>
      )}
      {action && <div className="mt-4">{action}</div>}
    </div>
  );
}
