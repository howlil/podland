/**
 * Application-wide constants
 */

// Refresh intervals (milliseconds)
export const REFRESH_INTERVALS = {
  VM_STATUS: 5000,        // VM status polling (5s)
  VM_DETAIL: 5000,        // Single VM detail polling (5s)
  QUOTA_USAGE: 30000,     // Quota usage refresh (30s)
  ACTIVITY_LOG: 10000,    // Activity log refresh (10s)
  METRICS: 30000,         // Metrics refresh (30s)
  LOGS_LIVE: 5000,        // Live logs refresh (5s)
  HEALTH: 30000,          // System health refresh (30s)
  AUDIT_LOG: 60000,       // Audit log refresh (60s)
} as const;

// Polling intervals (deprecated - use REFRESH_INTERVALS)
export const POLLING_INTERVALS = REFRESH_INTERVALS;

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
