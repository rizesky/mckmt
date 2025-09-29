-- Initial database schema for MCKMT
-- This migration creates all the core tables and indexes

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Create clusters table
CREATE TABLE IF NOT EXISTS clusters (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text NOT NULL UNIQUE,
    description text,
    labels jsonb DEFAULT '{}'::jsonb,
    encrypted_credentials bytea,
    status text DEFAULT 'pending' CHECK (status IN ('pending', 'connected', 'disconnected', 'error')),
    last_seen_at timestamptz,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

-- Create cluster_agents table
CREATE TABLE IF NOT EXISTS cluster_agents (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    cluster_id uuid NOT NULL REFERENCES clusters(id) ON DELETE CASCADE,
    agent_version text NOT NULL,
    connected_at timestamptz,
    last_heartbeat timestamptz,
    fingerprint text,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

-- Create operations table
CREATE TABLE IF NOT EXISTS operations (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    cluster_id uuid NOT NULL REFERENCES clusters(id) ON DELETE CASCADE,
    type text NOT NULL,
    status text DEFAULT 'queued' CHECK (status IN ('queued', 'running', 'success', 'failed', 'cancelled')),
    payload jsonb DEFAULT '{}'::jsonb,
    result jsonb,
    started_at timestamptz,
    finished_at timestamptz,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

-- Create audit_logs table
CREATE TABLE IF NOT EXISTS audit_logs (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id text NOT NULL,
    action text NOT NULL,
    resource_type text NOT NULL,
    resource_id text NOT NULL,
    request_payload jsonb,
    response_payload jsonb,
    ip_address text,
    user_agent text,
    created_at timestamptz DEFAULT now()
);

-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    username text NOT NULL UNIQUE,
    email text NOT NULL UNIQUE,
    roles text[] DEFAULT '{}',
    active boolean DEFAULT true,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

-- Create roles table
CREATE TABLE IF NOT EXISTS roles (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text NOT NULL UNIQUE,
    description text,
    permissions text[] DEFAULT '{}',
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_clusters_name ON clusters(name);
CREATE INDEX IF NOT EXISTS idx_clusters_status ON clusters(status);
CREATE INDEX IF NOT EXISTS idx_clusters_created_at ON clusters(created_at);

CREATE INDEX IF NOT EXISTS idx_cluster_agents_cluster_id ON cluster_agents(cluster_id);
CREATE INDEX IF NOT EXISTS idx_cluster_agents_last_heartbeat ON cluster_agents(last_heartbeat);

CREATE INDEX IF NOT EXISTS idx_operations_cluster_id ON operations(cluster_id);
CREATE INDEX IF NOT EXISTS idx_operations_status ON operations(status);
CREATE INDEX IF NOT EXISTS idx_operations_type ON operations(type);
CREATE INDEX IF NOT EXISTS idx_operations_created_at ON operations(created_at);

CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_resource_type ON audit_logs(resource_type);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at);

CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

CREATE INDEX IF NOT EXISTS idx_roles_name ON roles(name);

-- Insert initial data
-- Create default roles
INSERT INTO roles (id, name, description, permissions, created_at, updated_at) VALUES
('00000000-0000-0000-0000-000000000001', 'admin', 'Administrator role with full access', 
 ARRAY['clusters:read', 'clusters:write', 'clusters:delete', 'operations:read', 'operations:write', 'users:read', 'users:write', 'users:delete'], 
 NOW(), NOW()),
('00000000-0000-0000-0000-000000000002', 'operator', 'Operator role with cluster management access', 
 ARRAY['clusters:read', 'clusters:write', 'operations:read', 'operations:write'], 
 NOW(), NOW()),
('00000000-0000-0000-0000-000000000003', 'viewer', 'Viewer role with read-only access', 
 ARRAY['clusters:read', 'operations:read'], 
 NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- Create default admin user
INSERT INTO users (id, username, email, roles, active, created_at, updated_at) VALUES
('00000000-0000-0000-0000-000000000001', 'admin', 'admin@mckmt.local', 
 ARRAY['admin'], true, NOW(), NOW())
ON CONFLICT (id) DO NOTHING;
