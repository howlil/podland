import * as React from "react"
import { cn } from "@/lib/utils"
import { AlertCircle, CheckCircle2, AlertTriangle, Info } from "lucide-react"

export interface AlertProps
  extends React.HTMLAttributes<HTMLDivElement> {
  variant?: "default" | "destructive" | "success" | "warning"
}

const Alert = React.forwardRef<HTMLDivElement, AlertProps>(
  ({ className, variant = "default", ...props }, ref) => {
    const variants = {
      default: "bg-blue-50 dark:bg-blue-900/20 border-blue-200 dark:border-blue-800 text-blue-800 dark:text-blue-300",
      destructive: "bg-red-50 dark:bg-red-900/20 border-red-200 dark:border-red-800 text-red-800 dark:text-red-300",
      success: "bg-green-50 dark:bg-green-900/20 border-green-200 dark:border-green-800 text-green-800 dark:text-green-300",
      warning: "bg-yellow-50 dark:bg-yellow-900/20 border-yellow-200 dark:border-yellow-800 text-yellow-800 dark:text-yellow-300",
    }

    return (
      <div
        ref={ref}
        role="alert"
        className={cn(
          "relative w-full rounded-lg border p-4",
          variants[variant],
          className
        )}
        {...props}
      />
    )
  }
)
Alert.displayName = "Alert"

const AlertTitle = React.forwardRef<
  HTMLParagraphElement,
  React.HTMLAttributes<HTMLHeadingElement>
>(({ className, ...props }, ref) => (
  <h5
    ref={ref}
    className={cn("mb-1 font-medium leading-none tracking-tight", className)}
    {...props}
  />
))
AlertTitle.displayName = "AlertTitle"

const AlertDescription = React.forwardRef<
  HTMLParagraphElement,
  React.HTMLAttributes<HTMLParagraphElement>
>(({ className, ...props }, ref) => (
  <div
    ref={ref}
    className={cn("text-sm [&_p]:leading-relaxed", className)}
    {...props}
  />
))
AlertDescription.displayName = "AlertDescription"

const AlertIcon = React.forwardRef<
  HTMLSpanElement,
  React.HTMLAttributes<HTMLSpanElement> & { variant?: "default" | "destructive" | "success" | "warning" }
>(({ className, variant = "default", ...props }, ref) => {
  const icons = {
    default: Info,
    destructive: AlertCircle,
    success: CheckCircle2,
    warning: AlertTriangle,
  }
  const Icon = icons[variant]

  return (
    <span
      ref={ref}
      className={cn("absolute left-3 top-3", className)}
      {...props}
    >
      <Icon className="h-4 w-4" />
    </span>
  )
})
AlertIcon.displayName = "AlertIcon"

export { Alert, AlertTitle, AlertDescription, AlertIcon }
