import { type ClassValue, clsx } from "clsx"

export function cn(...inputs: ClassValue[]) {
  return clsx(inputs)
}

/**
 * Format bytes to human-readable string
 * @param bytes - Number of bytes
 * @param decimals - Number of decimal places (default: 2)
 * @returns Formatted string (e.g., "1.5 GB")
 */
export function formatBytes(bytes: number, decimals = 2): string {
  if (bytes === 0) return "0 B"
  const k = 1024
  const sizes = ["B", "KB", "MB", "GB", "TB", "PB"]
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(decimals)) + " " + sizes[i]
}

/**
 * Format date to relative time (e.g., "2 hours ago")
 * @param date - Date string or Date object
 * @returns Relative time string
 */
export function formatRelativeTime(date: string | Date): string {
  const now = new Date()
  const then = new Date(date)
  const diffMs = now.getTime() - then.getTime()
  const diffSecs = Math.floor(diffMs / 1000)
  const diffMins = Math.floor(diffSecs / 60)
  const diffHours = Math.floor(diffMins / 60)
  const diffDays = Math.floor(diffHours / 24)

  if (diffSecs < 60) return "just now"
  if (diffMins < 60) return `${diffMins}m ago`
  if (diffHours < 24) return `${diffHours}h ago`
  if (diffDays < 7) return `${diffDays}d ago`
  return then.toLocaleDateString()
}

/**
 * Sanitize string for safe display (basic XSS prevention)
 * @param str - String to sanitize
 * @returns Sanitized string
 */
export function sanitizeString(str: string): string {
  const div = document.createElement("div")
  div.textContent = str
  return div.innerHTML
}

/**
 * Check if element is visible in viewport
 * @param element - DOM element to check
 * @returns True if element is visible
 */
export function isElementInViewport(element: Element): boolean {
  const rect = element.getBoundingClientRect()
  return (
    rect.top >= 0 &&
    rect.left >= 0 &&
    rect.bottom <= (window.innerHeight || document.documentElement.clientHeight) &&
    rect.right <= (window.innerWidth || document.documentElement.clientWidth)
  )
}
