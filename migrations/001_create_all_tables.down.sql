-- =====================================================
-- AAA Service Database Migration - Rollback
-- Drops all tables created in the up migration
-- =====================================================

-- Drop triggers first
DROP TRIGGER IF EXISTS update_event_checkpoints_updated_at ON event_checkpoints;
DROP TRIGGER IF EXISTS update_events_updated_at ON events;
DROP TRIGGER IF EXISTS update_audit_logs_updated_at ON audit_logs;
DROP TRIGGER IF EXISTS update_column_sets_updated_at ON column_sets;
DROP TRIGGER IF EXISTS update_column_group_members_updated_at ON column_group_members;
DROP TRIGGER IF EXISTS update_column_groups_updated_at ON column_groups;
DROP TRIGGER IF EXISTS update_binding_history_updated_at ON binding_history;
DROP TRIGGER IF EXISTS update_bindings_updated_at ON bindings;
DROP TRIGGER IF EXISTS update_attribute_history_updated_at ON attribute_history;
DROP TRIGGER IF EXISTS update_attributes_updated_at ON attributes;
DROP TRIGGER IF EXISTS update_services_updated_at ON services;
DROP TRIGGER IF EXISTS update_principals_updated_at ON principals;
DROP TRIGGER IF EXISTS update_group_inheritance_updated_at ON group_inheritance;
DROP TRIGGER IF EXISTS update_group_memberships_updated_at ON group_memberships;
DROP TRIGGER IF EXISTS update_groups_updated_at ON groups;
DROP TRIGGER IF EXISTS update_resource_permissions_updated_at ON resource_permissions;
DROP TRIGGER IF EXISTS update_role_permissions_updated_at ON role_permissions;
DROP TRIGGER IF EXISTS update_user_roles_updated_at ON user_roles;
DROP TRIGGER IF EXISTS update_permissions_updated_at ON permissions;
DROP TRIGGER IF EXISTS update_resources_updated_at ON resources;
DROP TRIGGER IF EXISTS update_actions_updated_at ON actions;
DROP TRIGGER IF EXISTS update_roles_updated_at ON roles;
DROP TRIGGER IF EXISTS update_organizations_updated_at ON organizations;
DROP TRIGGER IF EXISTS update_contacts_updated_at ON contacts;
DROP TRIGGER IF EXISTS update_addresses_updated_at ON addresses;
DROP TRIGGER IF EXISTS update_user_profiles_updated_at ON user_profiles;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- Drop the function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables in reverse order (child tables first)
DROP TABLE IF EXISTS event_checkpoints CASCADE;
DROP TABLE IF EXISTS events CASCADE;
DROP TABLE IF EXISTS audit_logs CASCADE;
DROP TABLE IF EXISTS column_sets CASCADE;
DROP TABLE IF EXISTS column_group_members CASCADE;
DROP TABLE IF EXISTS column_groups CASCADE;
DROP TABLE IF EXISTS binding_history CASCADE;
DROP TABLE IF EXISTS bindings CASCADE;
DROP TABLE IF EXISTS attribute_history CASCADE;
DROP TABLE IF EXISTS attributes CASCADE;
DROP TABLE IF EXISTS services CASCADE;
DROP TABLE IF EXISTS principals CASCADE;
DROP TABLE IF EXISTS group_inheritance CASCADE;
DROP TABLE IF EXISTS group_memberships CASCADE;
DROP TABLE IF EXISTS groups CASCADE;
DROP TABLE IF EXISTS resource_permissions CASCADE;
DROP TABLE IF EXISTS role_permissions CASCADE;
DROP TABLE IF EXISTS user_roles CASCADE;
DROP TABLE IF EXISTS permissions CASCADE;
DROP TABLE IF EXISTS resources CASCADE;
DROP TABLE IF EXISTS actions CASCADE;
DROP TABLE IF EXISTS roles CASCADE;
DROP TABLE IF EXISTS contacts CASCADE;
DROP TABLE IF EXISTS user_profiles CASCADE;
DROP TABLE IF EXISTS addresses CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS organizations CASCADE;

-- =====================================================
-- ROLLBACK COMPLETE
-- =====================================================
