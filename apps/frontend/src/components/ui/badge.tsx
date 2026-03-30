import * as React from "react";
import { cva } from "class-variance-authority";
import { cn } from "@/lib/utils";

export interface BadgeProps extends React.HTMLAttributes<HTMLDivElement> {
  variant?: "default" | "secondary" | "destructive" | "outline";
}

const badgeVariants = cva(
  "inline-flex items-center rounded-full px-3 py-1 text-xs font-semibold transition-colors focus:outline-none focus:ring-2 focus:ring-primary focus:ring-offset-2",
  {
    variants: {
      variant: {
        default: "bg-primary text-primary-foreground hover:bg-primary/80",
        secondary: "bg-gray-100 text-gray-900 dark:bg-gray-800 dark:text-gray-100",
        destructive: "bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400",
        outline: "border border-gray-300 dark:border-gray-600 text-gray-900 dark:text-gray-100",
      },
    },
    defaultVariants: {
      variant: "default",
    },
  }
);

function Badge({ className, variant, ...props }: BadgeProps) {
  return <div className={cn(badgeVariants({ variant }), className)} {...props} />;
}

export { Badge, badgeVariants };
