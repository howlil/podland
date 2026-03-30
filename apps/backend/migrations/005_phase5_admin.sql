-- Phase 5: Admin Panel, Idle Detection, and VM Pinning
-- Creates tables for audit logging and adds pinning support to VMs

-- audit_logs table: Stores all admin actions for auditing
CREATE TABLE IF NOT EXISTS audit_logs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID REFERENCES users(id),
  action VARCHAR(255) NOT NULL,
  ip_address INET,
  user_agent TEXT,
  created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at);

-- Add is_pinned column to vms table (prevents auto-deletion)
ALTER TABLE vms ADD COLUMN IF NOT EXISTS is_pinned BOOLEAN NOT NULL DEFAULT false;
CREATE INDEX IF NOT EXISTS idx_vms_is_pinned ON vms(is_pinned);

-- Add idle_warned_at column to vms table (tracks when warning was sent)
ALTER TABLE vms ADD COLUMN IF NOT EXISTS idle_warned_at TIMESTAMP;
CREATE INDEX IF NOT EXISTS idx_vms_idle_warned_at ON vms(idle_warned_at);

-- Add superadmin role to users (if not already present in check constraint)
-- Note: This is handled in application logic; no DB constraint needed
