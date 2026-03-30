-- Phase 4: Observability - Notifications Schema
-- Creates table for storing in-app notifications from alerts

-- notifications table: Stores alert notifications for users
CREATE TABLE IF NOT EXISTS notifications (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id),
  vm_id UUID NOT NULL REFERENCES vms(id),
  alert_name VARCHAR(100) NOT NULL,
  severity VARCHAR(20) NOT NULL DEFAULT 'warning',
  title VARCHAR(255) NOT NULL,
  message TEXT NOT NULL,
  is_read BOOLEAN NOT NULL DEFAULT false,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  resolved_at TIMESTAMP
);

-- Indexes for common queries
CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_vm_id ON notifications(vm_id);
CREATE INDEX IF NOT EXISTS idx_notifications_is_read ON notifications(is_read);
CREATE INDEX IF NOT EXISTS idx_notifications_created_at ON notifications(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_notifications_user_unread ON notifications(user_id, is_read) WHERE is_read = false;

-- Updated_at trigger not needed since we don't update notifications frequently
-- Resolved_at is set directly on insert for resolved alerts
