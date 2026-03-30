import * as React from "react";
import { cn } from "@/lib/utils";

export interface EmptyStateProps {
  icon?: React.ElementType;
  title: string;
  description?: string;
  action?: React.ReactNode;
  actionLabel?: string;
  actionHref?: string;
  className?: string;
}

export function EmptyState({ 
  icon: Icon, 
  title, 
  description, 
  action, 
  actionLabel,
  actionHref,
  className 
}: EmptyStateProps) {
  return (
    <div className={cn("text-center py-12", className)}>
      {Icon && (
        <Icon className="h-16 w-16 mx-auto mb-4 text-gray-400 dark:text-gray-600" aria-hidden="true" />
      )}
      <h3 className="text-lg font-medium text-gray-900 dark:text-white mb-2">{title}</h3>
      {description && (
        <p className="text-gray-600 dark:text-gray-400 mb-6">{description}</p>
      )}
      {action && <div className="flex justify-center">{action}</div>}
      {actionLabel && actionHref && (
        <a
          href={actionHref}
          className="mt-4 inline-block text-blue-600 hover:text-blue-700 dark:text-blue-400"
        >
          {actionLabel} →
        </a>
      )}
    </div>
  );
}
