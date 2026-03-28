-- Phase 3: Domain Status Migration
-- Adds domain_status column to vms table for tracking DNS propagation status

-- Add domain_status column to vms table
ALTER TABLE vms
ADD COLUMN IF NOT EXISTS domain_status VARCHAR(20) DEFAULT 'pending' CHECK (domain_status IN ('pending', 'active', 'error'));

-- Create index on domain_status for efficient filtering
CREATE INDEX IF NOT EXISTS idx_vms_domain_status ON vms(domain_status);

-- Create index on domain for fast lookups (if not already exists)
CREATE INDEX IF NOT EXISTS idx_vms_domain ON vms(domain);

-- Add unique constraint on domain (optional, can be enabled if needed)
-- ALTER TABLE vms ADD CONSTRAINT uk_vms_domain UNIQUE (domain);
