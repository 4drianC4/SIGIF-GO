-- SIGIF Database Initialization Script
-- This script runs on first database creation

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Enable pgcrypto for gen_random_uuid
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Create schemas if needed
-- All tables will be in public schema by default

-- Set timezone
SET timezone = 'UTC';

-- Create initial admin tenant (will be managed by application)
-- INSERT INTO tenants (id, name, slug, business_type, is_active, settings, created_at, updated_at)
-- VALUES (gen_random_uuid(), 'SIGIF Demo', 'sigif-demo', 'minimarket', true, '{"currency": "USD", "timezone": "UTC", "language": "es", "tax_rate": 0}', NOW(), NOW())
-- ON CONFLICT (slug) DO NOTHING;