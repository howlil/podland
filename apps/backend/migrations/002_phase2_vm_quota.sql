-- Phase 2: VM and Quota Schema Migration
-- Creates tables for VM management, user quotas, tiers, and quota usage tracking

-- vms table: Stores VM instances
CREATE TABLE IF NOT EXISTS vms (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id),
  name VARCHAR(100) NOT NULL,
  os VARCHAR(50) NOT NULL DEFAULT 'ubuntu-2204',
  tier VARCHAR(20) NOT NULL,
  cpu DECIMAL(4,2) NOT NULL,
  ram BIGINT NOT NULL,
  storage BIGINT NOT NULL,
  status VARCHAR(20) NOT NULL DEFAULT 'pending',
  k8s_namespace VARCHAR(100),
  k8s_deployment VARCHAR(100),
  k8s_service VARCHAR(100),
  k8s_pvc VARCHAR(100),
  domain VARCHAR(255),
  ssh_public_key TEXT,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  started_at TIMESTAMP,
  stopped_at TIMESTAMP,
  deleted_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_vms_user_id ON vms(user_id);
CREATE INDEX IF NOT EXISTS idx_vms_status ON vms(status);

-- user_quotas table: Stores quota limits per user
CREATE TABLE IF NOT EXISTS user_quotas (
  user_id UUID PRIMARY KEY REFERENCES users(id),
  cpu_limit DECIMAL(4,2) NOT NULL DEFAULT 0.5,
  ram_limit BIGINT NOT NULL DEFAULT 1073741824,
  storage_limit BIGINT NOT NULL DEFAULT 10737418240,
  vm_count_limit INTEGER NOT NULL DEFAULT 2,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- user_quota_usage table: Tracks current quota usage per user
CREATE TABLE IF NOT EXISTS user_quota_usage (
  user_id UUID PRIMARY KEY REFERENCES users(id),
  cpu_used DECIMAL(4,2) NOT NULL DEFAULT 0,
  ram_used BIGINT NOT NULL DEFAULT 0,
  storage_used BIGINT NOT NULL DEFAULT 0,
  vm_count INTEGER NOT NULL DEFAULT 0,
  updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- tiers table: Defines available VM tiers
CREATE TABLE IF NOT EXISTS tiers (
  name VARCHAR(20) PRIMARY KEY,
  cpu DECIMAL(4,2) NOT NULL,
  ram BIGINT NOT NULL,
  storage BIGINT NOT NULL,
  min_role VARCHAR(20) NOT NULL DEFAULT 'external'
);

-- Insert default tiers (nano → xlarge)
INSERT INTO tiers (name, cpu, ram, storage, min_role) VALUES
  ('nano', 0.25, 536870912, 5368709120, 'external'),
  ('micro', 0.5, 1073741824, 10737418240, 'external'),
  ('small', 1.0, 2147483648, 21474836480, 'internal'),
  ('medium', 2.0, 4294967296, 42949672960, 'internal'),
  ('large', 4.0, 8589934592, 85899345920, 'internal'),
  ('xlarge', 4.0, 8589934592, 107374182400, 'internal')
ON CONFLICT (name) DO NOTHING;

-- Trigger function: Auto-create quota on user creation
CREATE OR REPLACE FUNCTION create_user_quota()
RETURNS TRIGGER AS $$
BEGIN
  -- Insert quota limits based on user role (internal vs external)
  INSERT INTO user_quotas (user_id, cpu_limit, ram_limit, storage_limit, vm_count_limit)
  VALUES (
    NEW.id,
    CASE 
      WHEN NEW.role = 'internal' OR NEW.nim LIKE '%1152%' THEN 4.0 
      ELSE 0.5 
    END,
    CASE 
      WHEN NEW.role = 'internal' OR NEW.nim LIKE '%1152%' THEN 8589934592 
      ELSE 1073741824 
    END,
    CASE 
      WHEN NEW.role = 'internal' OR NEW.nim LIKE '%1152%' THEN 107374182400 
      ELSE 10737418240 
    END,
    CASE 
      WHEN NEW.role = 'internal' OR NEW.nim LIKE '%1152%' THEN 5 
      ELSE 2 
    END
  );

  -- Initialize usage tracking
  INSERT INTO user_quota_usage (user_id, cpu_used, ram_used, storage_used, vm_count)
  VALUES (NEW.id, 0, 0, 0, 0);

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger on users table
DROP TRIGGER IF EXISTS trigger_create_user_quota ON users;
CREATE TRIGGER trigger_create_user_quota
AFTER INSERT ON users
FOR EACH ROW
EXECUTE FUNCTION create_user_quota();

-- Updated_at trigger for vms table
DROP TRIGGER IF EXISTS update_vms_updated_at ON vms;
CREATE TRIGGER update_vms_updated_at
BEFORE UPDATE ON vms
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- Updated_at trigger for user_quotas table
DROP TRIGGER IF EXISTS update_user_quotas_updated_at ON user_quotas;
CREATE TRIGGER update_user_quotas_updated_at
BEFORE UPDATE ON user_quotas
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- Updated_at trigger for user_quota_usage table
DROP TRIGGER IF EXISTS update_user_quota_usage_updated_at ON user_quota_usage;
CREATE TRIGGER update_user_quota_usage_updated_at
BEFORE UPDATE ON user_quota_usage
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();
