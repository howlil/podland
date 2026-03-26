// User types
export interface User {
  id: string
  githubId: string
  email: string
  displayName: string
  avatarUrl: string
  nim: string
  role: 'internal' | 'external' | 'superadmin'
  createdAt: string
  updatedAt: string
}

export interface UserCreateInput {
  githubId: string
  email: string
  displayName: string
  avatarUrl: string
  nim: string
  role: 'internal' | 'external'
}

export interface UserUpdateInput {
  displayName?: string
  avatarUrl?: string
}

// Session types
export interface Session {
  id: string
  userId: string
  refreshTokenHash: string
  jti: string
  deviceInfo: DeviceInfo
  createdAt: string
  expiresAt: string
  revokedAt?: string
  replacedBy?: string
}

export interface DeviceInfo {
  userAgent: string
  ip: string
}

// Activity types
export type ActivityAction =
  | 'account_created'
  | 'signed_in'
  | 'signed_out'
  | 'nim_confirmed'

export interface ActivityLog {
  id: string
  userId: string
  action: ActivityAction
  metadata: Record<string, unknown> | null
  createdAt: string
}

// Quota types
export interface Quota {
  cpu: number
  ram: number // in MB
  storage: number // in GB
  vmCount: number
}

export interface QuotaUsage {
  usedCpu: number
  maxCpu: number
  usedRam: number
  maxRam: number
  usedStorage: number
  maxStorage: number
  usedVmCount: number
  maxVmCount: number
}

// Role-based quotas
export const ROLE_QUOTAS: Record<string, Quota> = {
  internal: {
    cpu: 1,
    ram: 2048,
    storage: 10,
    vmCount: 5,
  },
  external: {
    cpu: 0.5,
    ram: 1024,
    storage: 5,
    vmCount: 2,
  },
  superadmin: {
    cpu: 4,
    ram: 8192,
    storage: 50,
    vmCount: 20,
  },
}

// API types
export interface ApiResponse<T> {
  data?: T
  error?: string
  message?: string
}

export interface PaginatedResponse<T> {
  data: T[]
  total: number
  page: number
  pageSize: number
  hasMore: boolean
}

// Auth types
export interface LoginRequest {
  code: string
  state: string
}

export interface LoginResponse {
  accessToken: string
  user: User
}

export interface RefreshResponse {
  accessToken: string
}

// VM types (for Phase 2)
export type VMStatus = 'pending' | 'running' | 'stopped' | 'error' | 'deleting'

export interface VM {
  id: string
  userId: string
  name: string
  os: 'ubuntu-22.04' | 'debian-12'
  cpu: number
  ram: number
  storage: number
  status: VMStatus
  domain?: string
  createdAt: string
  updatedAt: string
}

export interface VMCreateInput {
  name: string
  os: 'ubuntu-22.04' | 'debian-12'
  cpu: number
  ram: number
  storage: number
}

// Domain types (for Phase 3)
export interface Domain {
  id: string
  vmId: string
  userId: string
  subdomain: string
  domain: string
  status: 'pending' | 'active' | 'error'
  createdAt: string
}

// Error types
export class AppError extends Error {
  constructor(
    public code: string,
    message: string,
    public status: number = 500,
  ) {
    super(message)
    this.name = 'AppError'
  }
}

export const ErrorCodes = {
  // Auth errors
  INVALID_CREDENTIALS: 'INVALID_CREDENTIALS',
  INVALID_TOKEN: 'INVALID_TOKEN',
  TOKEN_EXPIRED: 'TOKEN_EXPIRED',
  INVALID_EMAIL: 'INVALID_EMAIL',
  EMAIL_NOT_VERIFIED: 'EMAIL_NOT_VERIFIED',
  
  // User errors
  USER_NOT_FOUND: 'USER_NOT_FOUND',
  USER_ALREADY_EXISTS: 'USER_ALREADY_EXISTS',
  INVALID_NIM: 'INVALID_NIM',
  
  // Session errors
  SESSION_NOT_FOUND: 'SESSION_NOT_FOUND',
  SESSION_EXPIRED: 'SESSION_EXPIRED',
  SESSION_LIMIT_REACHED: 'SESSION_LIMIT_REACHED',
  
  // VM errors
  VM_NOT_FOUND: 'VM_NOT_FOUND',
  VM_ALREADY_EXISTS: 'VM_ALREADY_EXISTS',
  QUOTA_EXCEEDED: 'QUOTA_EXCEEDED',
  INVALID_VM_STATUS: 'INVALID_VM_STATUS',
  
  // Domain errors
  DOMAIN_NOT_FOUND: 'DOMAIN_NOT_FOUND',
  DOMAIN_ALREADY_EXISTS: 'DOMAIN_ALREADY_EXISTS',
  DOMAIN_PENDING: 'DOMAIN_PENDING',
  
  // System errors
  INTERNAL_ERROR: 'INTERNAL_ERROR',
  SERVICE_UNAVAILABLE: 'SERVICE_UNAVAILABLE',
  RATE_LIMIT_EXCEEDED: 'RATE_LIMIT_EXCEEDED',
} as const
