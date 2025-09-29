-- Rollback initial schema
-- This will drop all tables and indexes

-- Drop indexes
DROP INDEX IF EXISTS idx_clusters_name;
DROP INDEX IF EXISTS idx_clusters_status;
DROP INDEX IF EXISTS idx_clusters_created_at;
DROP INDEX IF EXISTS idx_cluster_agents_cluster_id;
DROP INDEX IF EXISTS idx_cluster_agents_last_heartbeat;
DROP INDEX IF EXISTS idx_operations_cluster_id;
DROP INDEX IF EXISTS idx_operations_status;
DROP INDEX IF EXISTS idx_operations_type;
DROP INDEX IF EXISTS idx_operations_created_at;
DROP INDEX IF EXISTS idx_audit_logs_user_id;
DROP INDEX IF EXISTS idx_audit_logs_action;
DROP INDEX IF EXISTS idx_audit_logs_resource_type;
DROP INDEX IF EXISTS idx_audit_logs_created_at;
DROP INDEX IF EXISTS idx_users_username;
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_roles_name;

-- Remove initial data first
DELETE FROM users WHERE id = '00000000-0000-0000-0000-000000000001';
DELETE FROM roles WHERE id IN ('00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000003');

-- Drop tables (in reverse order due to foreign key constraints)
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS operations;
DROP TABLE IF EXISTS cluster_agents;
DROP TABLE IF EXISTS clusters;

-- Drop extensions
DROP EXTENSION IF EXISTS "pgcrypto";
DROP EXTENSION IF EXISTS "uuid-ossp";
