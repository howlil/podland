/**
 * Error handling utilities for consistent error messages
 */

import { toast } from "sonner";

/**
 * Extract error message from API error response
 * @param error - Error object from API call
 * @param defaultAction - Default action description for context
 * @returns Formatted error message
 */
export function getErrorMessage(error: any, defaultAction: string = "Operation failed"): string {
  const serverMessage = error.response?.data?.message;
  const networkError = !error.response && error.message;
  
  if (serverMessage) {
    return `${defaultAction}: ${serverMessage}`;
  }
  
  if (networkError) {
    return `${defaultAction}: ${error.message}`;
  }
  
  return `${defaultAction}: Unknown error`;
}

/**
 * Show error toast with consistent formatting
 * @param error - Error object from API call
 * @param defaultAction - Default action description for context
 */
export function showErrorToast(error: any, defaultAction: string = "Operation failed"): void {
  toast.error(getErrorMessage(error, defaultAction));
}

/**
 * Handle mutation error with optional callback
 * @param error - Error object from API call
 * @param defaultAction - Default action description for context
 * @param onError - Optional additional error handler
 */
export function handleMutationError(
  error: any,
  defaultAction: string,
  onError?: (errorMessage: string) => void
): void {
  const errorMessage = getErrorMessage(error, defaultAction);
  toast.error(errorMessage);
  onError?.(errorMessage);
}
