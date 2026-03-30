import * as React from "react";
import { cn } from "@/lib/utils";

export interface SkeletonProps extends React.HTMLAttributes<HTMLDivElement> {
  variant?: "text" | "circular" | "rectangular";
  width?: string | number;
  height?: string | number;
}

const Skeleton = React.forwardRef<HTMLDivElement, SkeletonProps>(
  ({ className, variant = "text", width, height, ...props }, ref) => {
    const baseStyles = "animate-pulse bg-gray-200 dark:bg-gray-700";
    
    const variants = {
      text: "rounded",
      circular: "rounded-full",
      rectangular: "rounded-xl",
    };

    const style: React.CSSProperties = {};
    if (width) style.width = typeof width === "string" ? width : `${width}px`;
    if (height) style.height = typeof height === "string" ? height : `${height}px`;

    return (
      <div
        ref={ref}
        className={cn(baseStyles, variants[variant], className)}
        style={style}
        {...props}
      />
    );
  }
);
Skeleton.displayName = "Skeleton";

// Preset skeletons for common use cases
export function SkeletonText({ lines = 1, className }: { lines?: number; className?: string }) {
  return (
    <div className={cn("space-y-2", className)}>
      {Array.from({ length: lines }).map((_, i) => (
        <Skeleton key={i} variant="text" className="h-4 w-full" />
      ))}
    </div>
  );
}

export function SkeletonCard({ className }: { className?: string }) {
  return (
    <div className={cn("bg-white dark:bg-gray-800 rounded-2xl p-6 border border-gray-200 dark:border-gray-700", className)}>
      <Skeleton variant="circular" width={48} height={48} className="mb-4" />
      <SkeletonText lines={2} />
    </div>
  );
}

export function SkeletonTable({ rows = 5 }: { rows?: number }) {
  return (
    <div className="space-y-3">
      {Array.from({ length: rows }).map((_, i) => (
        <div key={i} className="flex items-center gap-4 p-4 bg-white dark:bg-gray-800 rounded-xl animate-pulse">
          <Skeleton variant="circular" width={40} height={40} />
          <div className="flex-1 space-y-2">
            <Skeleton variant="text" className="h-4 w-32" />
            <Skeleton variant="text" className="h-3 w-24" />
          </div>
          <Skeleton variant="rectangular" width={80} height={32} />
        </div>
      ))}
    </div>
  );
}

export { Skeleton };
