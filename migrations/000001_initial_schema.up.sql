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
    endpoint text,
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
    password_hash text,
    auth_source text DEFAULT 'password',
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
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

-- Create permissions table
CREATE TABLE IF NOT EXISTS permissions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text NOT NULL UNIQUE,
    resource text NOT NULL,
    action text NOT NULL,
    description text,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

-- Create user_roles junction table
CREATE TABLE IF NOT EXISTS user_roles (
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id uuid NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    assigned_at timestamptz DEFAULT now(),
    assigned_by uuid REFERENCES users(id),
    PRIMARY KEY (user_id, role_id)
);

-- Create role_permissions junction table
CREATE TABLE IF NOT EXISTS role_permissions (
    role_id uuid NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id uuid NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    granted_at timestamptz DEFAULT now(),
    granted_by uuid REFERENCES users(id),
    PRIMARY KEY (role_id, permission_id)
);

-- Create user_clusters table for resource ownership (ABAC)
CREATE TABLE IF NOT EXISTS user_clusters (
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    cluster_id uuid NOT NULL REFERENCES clusters(id) ON DELETE CASCADE,
    access_level text DEFAULT 'read' CHECK (access_level IN ('read', 'write', 'admin')),
    granted_at timestamptz DEFAULT now(),
    granted_by uuid REFERENCES users(id),
    PRIMARY KEY (user_id, cluster_id)
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_clusters_name ON clusters(name);
CREATE INDEX IF NOT EXISTS idx_clusters_status ON clusters(status);
CREATE INDEX IF NOT EXISTS idx_clusters_endpoint ON clusters(endpoint);
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
CREATE INDEX IF NOT EXISTS idx_users_password_hash ON users(password_hash);
CREATE INDEX IF NOT EXISTS idx_users_auth_source ON users(auth_source);

CREATE INDEX IF NOT EXISTS idx_roles_name ON roles(name);
CREATE INDEX IF NOT EXISTS idx_permissions_name ON permissions(name);
CREATE INDEX IF NOT EXISTS idx_permissions_resource_action ON permissions(resource, action);
CREATE INDEX IF NOT EXISTS idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_role_id ON user_roles(role_id);
CREATE INDEX IF NOT EXISTS idx_role_permissions_role_id ON role_permissions(role_id);
CREATE INDEX IF NOT EXISTS idx_role_permissions_permission_id ON role_permissions(permission_id);
CREATE INDEX IF NOT EXISTS idx_user_clusters_user_id ON user_clusters(user_id);
CREATE INDEX IF NOT EXISTS idx_user_clusters_cluster_id ON user_clusters(cluster_id);

-- Insert initial data
-- Create default permissions (wildcards first for better ordering)
INSERT INTO permissions (id, name, resource, action, description, created_at, updated_at) VALUES
('00000000-0000-0000-0000-000000000001', '*:*', '*', '*', 'Full access to all resources and actions', NOW(), NOW()),
('00000000-0000-0000-0000-000000000002', '*:read', '*', 'read', 'Read access to all resources', NOW(), NOW()),
('00000000-0000-0000-0000-000000000003', 'clusters:*', 'clusters', '*', 'All actions on cluster resources', NOW(), NOW()),
('00000000-0000-0000-0000-000000000004', 'operations:*', 'operations', '*', 'All actions on operation resources', NOW(), NOW()),
('00000000-0000-0000-0000-000000000005', 'users:*', 'users', '*', 'All actions on user resources', NOW(), NOW()),
('00000000-0000-0000-0000-000000000006', 'system:*', 'system', '*', 'All actions on system resources', NOW(), NOW()),
('00000000-0000-0000-0000-000000000007', 'clusters:read', 'clusters', 'read', 'Read cluster information', NOW(), NOW()),
('00000000-0000-0000-0000-000000000008', 'clusters:write', 'clusters', 'write', 'Create and update clusters', NOW(), NOW()),
('00000000-0000-0000-0000-000000000009', 'clusters:delete', 'clusters', 'delete', 'Delete clusters', NOW(), NOW()),
('00000000-0000-0000-0000-000000000010', 'clusters:manage', 'clusters', 'manage', 'Manage cluster resources and manifests', NOW(), NOW()),
('00000000-0000-0000-0000-000000000011', 'operations:read', 'operations', 'read', 'Read operation information', NOW(), NOW()),
('00000000-0000-0000-0000-000000000012', 'operations:write', 'operations', 'write', 'Create operations', NOW(), NOW()),
('00000000-0000-0000-0000-000000000013', 'operations:cancel', 'operations', 'cancel', 'Cancel operations', NOW(), NOW()),
('00000000-0000-0000-0000-000000000014', 'users:read', 'users', 'read', 'Read user information', NOW(), NOW()),
('00000000-0000-0000-0000-000000000015', 'users:write', 'users', 'write', 'Create and update users', NOW(), NOW()),
('00000000-0000-0000-0000-000000000016', 'users:delete', 'users', 'delete', 'Delete users', NOW(), NOW()),
('00000000-0000-0000-0000-000000000017', 'system:read', 'system', 'read', 'Read system information', NOW(), NOW()),
('00000000-0000-0000-0000-000000000018', 'system:write', 'system', 'write', 'Manage system settings', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- Create default roles
INSERT INTO roles (id, name, description, created_at, updated_at) VALUES
('00000000-0000-0000-0000-000000000001', 'admin', 'Administrator role with full access', NOW(), NOW()),
('00000000-0000-0000-0000-000000000002', 'operator', 'Operator role with cluster management access', NOW(), NOW()),
('00000000-0000-0000-0000-000000000003', 'viewer', 'Viewer role with read-only access', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- Assign permissions to admin role (using wildcard for full access)
INSERT INTO role_permissions (role_id, permission_id) VALUES
('00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001')  -- *:* (full access)
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Assign permissions to operator role (using wildcards for cluster and operation access)
INSERT INTO role_permissions (role_id, permission_id) VALUES
('00000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000003'), -- clusters:* (all cluster actions)
('00000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000004'), -- operations:* (all operation actions)
('00000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000002')  -- *:read (read access to all resources)
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Assign permissions to viewer role (using wildcard for read-only access)
INSERT INTO role_permissions (role_id, permission_id) VALUES
('00000000-0000-0000-0000-000000000003', '00000000-0000-0000-0000-000000000002')  -- *:read (read access to all resources)
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Create default admin user with password: admin123
INSERT INTO users (id, username, email, password_hash, auth_source, active, created_at, updated_at) VALUES
('00000000-0000-0000-0000-000000000001', 'admin', 'admin@mckmt.local', 
 '$argon2id$v=19$m=65536,t=4,p=2$eHtOigMBSF8OqtSnq43o/A$SRP0Y1+V+a6npQdERQirv5mJoAIgOEeZ+BMJRq9aEKc', 'password', true, NOW(), NOW()) 
ON CONFLICT (id) DO NOTHING;

-- Assign admin role to admin user
INSERT INTO user_roles (user_id, role_id) VALUES
('00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001')
ON CONFLICT (user_id, role_id) DO NOTHING;
