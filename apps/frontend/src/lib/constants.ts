/**
 * Application-wide constants
 */

// Polling intervals (milliseconds)
export const POLLING_INTERVALS = {
  VM_STATUS: 5000,        // VM status polling
  QUOTA_USAGE: 30000,     // Quota usage refresh
  ACTIVITY_LOG: 10000,    // Activity log refresh
  METRICS: 60000,         // Metrics refresh
} as const;

// Token refresh timing (milliseconds)
export const AUTH = {
  REFRESH_AT: 7.5 * 60 * 1000,  // Refresh at 50% of 15-min token expiry
} as const;

// VM statuses
export const VM_STATUS = {
  PENDING: "pending",
  RUNNING: "running",
  STOPPED: "stopped",
  ERROR: "error",
} as const;

// OS options
export const OS_OPTIONS = {
  "ubuntu-2204": "Ubuntu 22.04 LTS",
  "debian-12": "Debian 12",
} as const;

// Domain suffix
export const DOMAIN_SUFFIX = ".podland.app";

// API endpoints
export const API = {
  VMS: "/vms",
  USERS_ME: "/users/me",
  AUTH_LOGIN: "/auth/login",
  AUTH_LOGOUT: "/auth/logout",
  AUTH_REFRESH: "/auth/refresh",
  TIERS: "/tiers",
  QUOTA: "/quota",
} as const;

// UI constants
export const UI = {
  TABLE_PAGE_SIZE: 10,
  LOG_LIMIT: 1000,
  TOAST_DURATION: 4000,
} as const;
