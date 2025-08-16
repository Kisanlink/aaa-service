-- =====================================================
-- AAA Service Database Migration
-- Creates all tables for the AAA (Authentication, Authorization, and Accounting) service
-- =====================================================

-- Enable UUID extension for PostgreSQL
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- =====================================================
-- CORE TABLES
-- =====================================================

-- Organizations table
CREATE TABLE organizations (
    id VARCHAR(255) PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255),
    updated_by VARCHAR(255),
    deleted_at TIMESTAMP,
    deleted_by VARCHAR(255),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    parent_id VARCHAR(255),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    metadata JSONB,

    CONSTRAINT uk_organizations_name UNIQUE (name),
    CONSTRAINT fk_organizations_parent FOREIGN KEY (parent_id) REFERENCES organizations(id) ON DELETE SET NULL
);

-- Users table
CREATE TABLE users (
    id VARCHAR(255) PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255),
    updated_by VARCHAR(255),
    deleted_at TIMESTAMP,
    deleted_by VARCHAR(255),
    phone_number VARCHAR(10) NOT NULL,
    country_code VARCHAR(10) NOT NULL DEFAULT '+91',
    username VARCHAR(100),
    password VARCHAR(255) NOT NULL,
    mpin VARCHAR(255),
    is_validated BOOLEAN NOT NULL DEFAULT FALSE,
    status VARCHAR(50) DEFAULT 'pending',
    tokens INTEGER NOT NULL DEFAULT 1000,

    CONSTRAINT uk_users_phone_number UNIQUE (phone_number),
    CONSTRAINT uk_users_username UNIQUE (username),
    CONSTRAINT chk_users_status CHECK (status IN ('pending', 'active', 'suspended', 'blocked')),
    CONSTRAINT chk_users_tokens CHECK (tokens >= 0)
);

-- User Profiles table
CREATE TABLE user_profiles (
    id VARCHAR(255) PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255),
    updated_by VARCHAR(255),
    deleted_at TIMESTAMP,
    deleted_by VARCHAR(255),
    user_id VARCHAR(255) NOT NULL,
    name VARCHAR(255),
    care_of VARCHAR(255),
    date_of_birth VARCHAR(10),
    photo TEXT,
    year_of_birth VARCHAR(4),
    message TEXT,
    aadhaar_number VARCHAR(12),
    email_hash VARCHAR(255),
    share_code VARCHAR(50),
    address_id VARCHAR(255),

    CONSTRAINT uk_user_profiles_user_id UNIQUE (user_id),
    CONSTRAINT fk_user_profiles_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_user_profiles_address FOREIGN KEY (address_id) REFERENCES addresses(id) ON DELETE SET NULL
);

-- Addresses table
CREATE TABLE addresses (
    id VARCHAR(255) PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255),
    updated_by VARCHAR(255),
    deleted_at TIMESTAMP,
    deleted_by VARCHAR(255),
    house VARCHAR(255),
    street VARCHAR(255),
    landmark VARCHAR(255),
    post_office VARCHAR(255),
    subdistrict VARCHAR(255),
    district VARCHAR(255),
    vtc VARCHAR(255),
    state VARCHAR(255),
    country VARCHAR(255),
    pincode VARCHAR(10),
    full_address TEXT
);

-- Contacts table
CREATE TABLE contacts (
    id VARCHAR(255) PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255),
    updated_by VARCHAR(255),
    deleted_at TIMESTAMP,
    deleted_by VARCHAR(255),
    user_id VARCHAR(255) NOT NULL,
    mobile_number BIGINT NOT NULL,
    country_code VARCHAR(10) DEFAULT '+91',
    email_hash VARCHAR(255),
    share_code VARCHAR(50),
    address_id VARCHAR(255),

    CONSTRAINT fk_contacts_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_contacts_address FOREIGN KEY (address_id) REFERENCES addresses(id) ON DELETE SET NULL
);

-- =====================================================
-- AUTHORIZATION TABLES
-- =====================================================

-- Roles table
CREATE TABLE roles (
    id VARCHAR(255) PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255),
    updated_by VARCHAR(255),
    deleted_at TIMESTAMP,
    deleted_by VARCHAR(255),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    scope VARCHAR(20) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    version INTEGER NOT NULL DEFAULT 1,
    metadata JSONB,
    organization_id VARCHAR(255),
    group_id VARCHAR(255),
    parent_id VARCHAR(255),

    CONSTRAINT uk_roles_name UNIQUE (name),
    CONSTRAINT chk_roles_scope CHECK (scope IN ('GLOBAL', 'ORG')),
    CONSTRAINT fk_roles_organization FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE SET NULL,
    CONSTRAINT fk_roles_parent FOREIGN KEY (parent_id) REFERENCES roles(id) ON DELETE SET NULL
);

-- Actions table
CREATE TABLE actions (
    id VARCHAR(255) PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255),
    updated_by VARCHAR(255),
    deleted_at TIMESTAMP,
    deleted_by VARCHAR(255),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    category VARCHAR(50) NOT NULL,
    is_static BOOLEAN NOT NULL DEFAULT FALSE,
    service_id VARCHAR(255),
    metadata JSONB,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,

    CONSTRAINT uk_actions_name UNIQUE (name),
    CONSTRAINT chk_actions_category CHECK (category IN ('user', 'role', 'system', 'api', 'database', 'audit', 'general'))
);

-- Resources table
CREATE TABLE resources (
    id VARCHAR(255) PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255),
    updated_by VARCHAR(255),
    deleted_at TIMESTAMP,
    deleted_by VARCHAR(255),
    name VARCHAR(100) NOT NULL,
    type VARCHAR(100) NOT NULL,
    description TEXT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    parent_id VARCHAR(255),
    owner_id VARCHAR(255),

    CONSTRAINT uk_resources_name UNIQUE (name),
    CONSTRAINT fk_resources_parent FOREIGN KEY (parent_id) REFERENCES resources(id) ON DELETE SET NULL,
    CONSTRAINT fk_resources_owner FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE SET NULL
);

-- Permissions table
CREATE TABLE permissions (
    id VARCHAR(255) PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255),
    updated_by VARCHAR(255),
    deleted_at TIMESTAMP,
    deleted_by VARCHAR(255),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    resource_id VARCHAR(255),
    action_id VARCHAR(255),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,

    CONSTRAINT uk_permissions_name UNIQUE (name),
    CONSTRAINT fk_permissions_resource FOREIGN KEY (resource_id) REFERENCES resources(id) ON DELETE SET NULL,
    CONSTRAINT fk_permissions_action FOREIGN KEY (action_id) REFERENCES actions(id) ON DELETE SET NULL
);

-- User Roles table (many-to-many relationship)
CREATE TABLE user_roles (
    id VARCHAR(255) PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255),
    updated_by VARCHAR(255),
    deleted_at TIMESTAMP,
    deleted_by VARCHAR(255),
    user_id VARCHAR(255) NOT NULL,
    role_id VARCHAR(255) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,

    CONSTRAINT fk_user_roles_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_user_roles_role FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
    CONSTRAINT uk_user_roles_user_role UNIQUE (user_id, role_id)
);

-- Role Permissions table (many-to-many relationship)
CREATE TABLE role_permissions (
    id VARCHAR(255) PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255),
    updated_by VARCHAR(255),
    deleted_at TIMESTAMP,
    deleted_by VARCHAR(255),
    role_id VARCHAR(255) NOT NULL,
    permission_id VARCHAR(255) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,

    CONSTRAINT fk_role_permissions_role FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
    CONSTRAINT fk_role_permissions_permission FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE,
    CONSTRAINT uk_role_permissions_role_permission UNIQUE (role_id, permission_id)
);

-- Resource Permissions table
CREATE TABLE resource_permissions (
    id VARCHAR(255) PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255),
    updated_by VARCHAR(255),
    deleted_at TIMESTAMP,
    deleted_by VARCHAR(255),
    role_id VARCHAR(255) NOT NULL,
    resource_type VARCHAR(100) NOT NULL,
    resource_id VARCHAR(255) NOT NULL,
    action VARCHAR(50) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,

    CONSTRAINT fk_resource_permissions_role FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
    CONSTRAINT uk_resource_permissions_role_resource_action UNIQUE (role_id, resource_type, resource_id, action)
);

-- =====================================================
-- GROUP AND ORGANIZATION TABLES
-- =====================================================

-- Groups table
CREATE TABLE groups (
    id VARCHAR(255) PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255),
    updated_by VARCHAR(255),
    deleted_at TIMESTAMP,
    deleted_by VARCHAR(255),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    organization_id VARCHAR(255) NOT NULL,
    parent_id VARCHAR(255),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    metadata JSONB,

    CONSTRAINT fk_groups_organization FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_groups_parent FOREIGN KEY (parent_id) REFERENCES groups(id) ON DELETE SET NULL
);

-- Group Memberships table
CREATE TABLE group_memberships (
    id VARCHAR(255) PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255),
    updated_by VARCHAR(255),
    deleted_at TIMESTAMP,
    deleted_by VARCHAR(255),
    group_id VARCHAR(255) NOT NULL,
    principal_id VARCHAR(255) NOT NULL,
    principal_type VARCHAR(50) NOT NULL,
    starts_at TIMESTAMP,
    ends_at TIMESTAMP,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    added_by_id VARCHAR(255) NOT NULL,
    metadata JSONB,

    CONSTRAINT chk_group_memberships_principal_type CHECK (principal_type IN ('user', 'service')),
    CONSTRAINT fk_group_memberships_group FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE,
    CONSTRAINT fk_group_memberships_added_by FOREIGN KEY (added_by_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Group Inheritance table
CREATE TABLE group_inheritance (
    id VARCHAR(255) PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255),
    updated_by VARCHAR(255),
    deleted_at TIMESTAMP,
    deleted_by VARCHAR(255),
    parent_group_id VARCHAR(255) NOT NULL,
    child_group_id VARCHAR(255) NOT NULL,
    starts_at TIMESTAMP,
    ends_at TIMESTAMP,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,

    CONSTRAINT fk_group_inheritance_parent FOREIGN KEY (parent_group_id) REFERENCES groups(id) ON DELETE CASCADE,
    CONSTRAINT fk_group_inheritance_child FOREIGN KEY (child_group_id) REFERENCES groups(id) ON DELETE CASCADE,
    CONSTRAINT uk_group_inheritance_parent_child UNIQUE (parent_group_id, child_group_id)
);

-- =====================================================
-- PRINCIPAL AND SERVICE TABLES
-- =====================================================

-- Principals table
CREATE TABLE principals (
    id VARCHAR(255) PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255),
    updated_by VARCHAR(255),
    deleted_at TIMESTAMP,
    deleted_by VARCHAR(255),
    type VARCHAR(20) NOT NULL,
    user_id VARCHAR(255),
    service_id VARCHAR(255),
    name VARCHAR(100) NOT NULL,
    organization_id VARCHAR(255),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    metadata JSONB,

    CONSTRAINT chk_principals_type CHECK (type IN ('user', 'service')),
    CONSTRAINT uk_principals_user_id UNIQUE (user_id),
    CONSTRAINT uk_principals_service_id UNIQUE (service_id),
    CONSTRAINT fk_principals_organization FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE SET NULL
);

-- Services table
CREATE TABLE services (
    id VARCHAR(255) PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255),
    updated_by VARCHAR(255),
    deleted_at TIMESTAMP,
    deleted_by VARCHAR(255),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    organization_id VARCHAR(255) NOT NULL,
    api_key VARCHAR(255) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    metadata JSONB,

    CONSTRAINT uk_services_name UNIQUE (name),
    CONSTRAINT uk_services_api_key UNIQUE (api_key),
    CONSTRAINT fk_services_organization FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE
);

-- =====================================================
-- ATTRIBUTE AND BINDING TABLES
-- =====================================================

-- Attributes table
CREATE TABLE attributes (
    id VARCHAR(255) PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255),
    updated_by VARCHAR(255),
    deleted_at TIMESTAMP,
    deleted_by VARCHAR(255),
    subject_id VARCHAR(255) NOT NULL,
    subject_type VARCHAR(20) NOT NULL,
    key VARCHAR(100) NOT NULL,
    value JSONB NOT NULL,
    organization_id VARCHAR(255),
    expires_at TIMESTAMP,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    set_by_id VARCHAR(255) NOT NULL,
    metadata JSONB,

    CONSTRAINT chk_attributes_subject_type CHECK (subject_type IN ('principal', 'resource', 'organization')),
    CONSTRAINT fk_attributes_organization FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE SET NULL,
    CONSTRAINT fk_attributes_set_by FOREIGN KEY (set_by_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT uk_attributes_subject_key UNIQUE (subject_id, subject_type, key)
);

-- Attribute History table
CREATE TABLE attribute_history (
    id VARCHAR(255) PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255),
    updated_by VARCHAR(255),
    deleted_at TIMESTAMP,
    deleted_by VARCHAR(255),
    attribute_id VARCHAR(255) NOT NULL,
    subject_id VARCHAR(255) NOT NULL,
    subject_type VARCHAR(20) NOT NULL,
    key VARCHAR(100) NOT NULL,
    old_value JSONB,
    new_value JSONB,
    action VARCHAR(20) NOT NULL,
    changed_by_id VARCHAR(255) NOT NULL,
    changed_at TIMESTAMP NOT NULL,
    organization_id VARCHAR(255),

    CONSTRAINT chk_attribute_history_subject_type CHECK (subject_type IN ('principal', 'resource', 'organization')),
    CONSTRAINT chk_attribute_history_action CHECK (action IN ('SET', 'DELETE')),
    CONSTRAINT fk_attribute_history_attribute FOREIGN KEY (attribute_id) REFERENCES attributes(id) ON DELETE CASCADE,
    CONSTRAINT fk_attribute_history_changed_by FOREIGN KEY (changed_by_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_attribute_history_organization FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE SET NULL
);

-- Bindings table
CREATE TABLE bindings (
    id VARCHAR(255) PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255),
    updated_by VARCHAR(255),
    deleted_at TIMESTAMP,
    deleted_by VARCHAR(255),
    subject_id VARCHAR(255) NOT NULL,
    subject_type VARCHAR(20) NOT NULL,
    binding_type VARCHAR(20) NOT NULL,
    role_id VARCHAR(255),
    permission_id VARCHAR(255),
    resource_type VARCHAR(100) NOT NULL,
    resource_id VARCHAR(255),
    organization_id VARCHAR(255) NOT NULL,
    caveat JSONB,
    version INTEGER NOT NULL DEFAULT 1,
    created_by_id VARCHAR(255) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,

    CONSTRAINT chk_bindings_subject_type CHECK (subject_type IN ('user', 'group', 'service')),
    CONSTRAINT chk_bindings_binding_type CHECK (binding_type IN ('role', 'permission')),
    CONSTRAINT fk_bindings_role FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE SET NULL,
    CONSTRAINT fk_bindings_permission FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE SET NULL,
    CONSTRAINT fk_bindings_organization FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_bindings_created_by FOREIGN KEY (created_by_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Binding History table
CREATE TABLE binding_history (
    id VARCHAR(255) PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255),
    updated_by VARCHAR(255),
    deleted_at TIMESTAMP,
    deleted_by VARCHAR(255),
    binding_id VARCHAR(255) NOT NULL,
    subject_id VARCHAR(255) NOT NULL,
    subject_type VARCHAR(20) NOT NULL,
    binding_type VARCHAR(20) NOT NULL,
    role_id VARCHAR(255),
    permission_id VARCHAR(255),
    resource_type VARCHAR(100) NOT NULL,
    resource_id VARCHAR(255),
    organization_id VARCHAR(255) NOT NULL,
    caveat JSONB,
    version INTEGER NOT NULL,
    action VARCHAR(20) NOT NULL,
    changed_by_id VARCHAR(255) NOT NULL,
    changed_at TIMESTAMP NOT NULL,

    CONSTRAINT chk_binding_history_subject_type CHECK (subject_type IN ('user', 'group', 'service')),
    CONSTRAINT chk_binding_history_binding_type CHECK (binding_type IN ('role', 'permission')),
    CONSTRAINT chk_binding_history_action CHECK (action IN ('CREATE', 'UPDATE', 'DELETE')),
    CONSTRAINT fk_binding_history_binding FOREIGN KEY (binding_id) REFERENCES bindings(id) ON DELETE CASCADE,
    CONSTRAINT fk_binding_history_changed_by FOREIGN KEY (changed_by_id) REFERENCES users(id) ON DELETE CASCADE
);

-- =====================================================
-- COLUMN PERMISSION TABLES
-- =====================================================

-- Column Groups table
CREATE TABLE column_groups (
    id VARCHAR(255) PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255),
    updated_by VARCHAR(255),
    deleted_at TIMESTAMP,
    deleted_by VARCHAR(255),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    table_name VARCHAR(100) NOT NULL,
    organization_id VARCHAR(255) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    metadata JSONB,

    CONSTRAINT fk_column_groups_organization FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE
);

-- Column Group Members table
CREATE TABLE column_group_members (
    id VARCHAR(255) PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255),
    updated_by VARCHAR(255),
    deleted_at TIMESTAMP,
    deleted_by VARCHAR(255),
    column_group_id VARCHAR(255) NOT NULL,
    column_name VARCHAR(100) NOT NULL,
    column_position INTEGER NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,

    CONSTRAINT fk_column_group_members_column_group FOREIGN KEY (column_group_id) REFERENCES column_groups(id) ON DELETE CASCADE,
    CONSTRAINT uk_column_group_members_group_column UNIQUE (column_group_id, column_name)
);

-- Column Sets table
CREATE TABLE column_sets (
    id VARCHAR(255) PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255),
    updated_by VARCHAR(255),
    deleted_at TIMESTAMP,
    deleted_by VARCHAR(255),
    name VARCHAR(100) NOT NULL,
    table_name VARCHAR(100) NOT NULL,
    column_group_id VARCHAR(255),
    bitmap BYTEA NOT NULL,
    column_count INTEGER NOT NULL,
    organization_id VARCHAR(255) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,

    CONSTRAINT fk_column_sets_column_group FOREIGN KEY (column_group_id) REFERENCES column_groups(id) ON DELETE SET NULL,
    CONSTRAINT fk_column_sets_organization FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE
);

-- =====================================================
-- AUDIT AND EVENT TABLES
-- =====================================================

-- Audit Logs table
CREATE TABLE audit_logs (
    id VARCHAR(255) PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255),
    updated_by VARCHAR(255),
    deleted_at TIMESTAMP,
    deleted_by VARCHAR(255),
    user_id VARCHAR(255),
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(100) NOT NULL,
    resource_id VARCHAR(255),
    ip_address VARCHAR(45),
    user_agent TEXT,
    status VARCHAR(20) NOT NULL,
    message TEXT,
    details JSONB,
    timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT chk_audit_logs_status CHECK (status IN ('success', 'failure', 'warning')),
    CONSTRAINT fk_audit_logs_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
);

-- Events table
CREATE TABLE events (
    id VARCHAR(255) PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255),
    updated_by VARCHAR(255),
    deleted_at TIMESTAMP,
    deleted_by VARCHAR(255),
    occurred_at TIMESTAMP NOT NULL,
    actor_id VARCHAR(255) NOT NULL,
    actor_type VARCHAR(20) NOT NULL,
    kind VARCHAR(50) NOT NULL,
    resource_type VARCHAR(100) NOT NULL,
    resource_id VARCHAR(255) NOT NULL,
    organization_id VARCHAR(255),
    payload JSONB NOT NULL,
    prev_hash VARCHAR(64),
    hash VARCHAR(64) NOT NULL,
    sequence_num BIGINT NOT NULL,
    request_id VARCHAR(255),
    source_ip VARCHAR(45),
    user_agent TEXT,

    CONSTRAINT chk_events_actor_type CHECK (actor_type IN ('user', 'service', 'system')),
    CONSTRAINT uk_events_hash UNIQUE (hash),
    CONSTRAINT uk_events_sequence_num UNIQUE (sequence_num),
    CONSTRAINT fk_events_organization FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE SET NULL
);

-- Event Checkpoints table
CREATE TABLE event_checkpoints (
    id VARCHAR(255) PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255),
    updated_by VARCHAR(255),
    deleted_at TIMESTAMP,
    deleted_by VARCHAR(255),
    checkpoint_time TIMESTAMP NOT NULL,
    last_event_id VARCHAR(255) NOT NULL,
    last_sequence_num BIGINT NOT NULL,
    last_event_hash VARCHAR(64) NOT NULL,
    merkle_root VARCHAR(64) NOT NULL,
    event_count BIGINT NOT NULL,
    created_by_id VARCHAR(255) NOT NULL,

    CONSTRAINT fk_event_checkpoints_last_event FOREIGN KEY (last_event_id) REFERENCES events(id) ON DELETE CASCADE,
    CONSTRAINT fk_event_checkpoints_created_by FOREIGN KEY (created_by_id) REFERENCES users(id) ON DELETE CASCADE
);

-- =====================================================
-- INDEXES FOR PERFORMANCE
-- =====================================================

-- Users indexes
CREATE INDEX idx_users_phone_number ON users(phone_number);
CREATE INDEX idx_users_country_code ON users(country_code);
CREATE INDEX idx_users_status ON users(status);
CREATE INDEX idx_users_is_validated ON users(is_validated);

-- User profiles indexes
CREATE INDEX idx_user_profiles_user_id ON user_profiles(user_id);
CREATE INDEX idx_user_profiles_address_id ON user_profiles(address_id);

-- Addresses indexes
CREATE INDEX idx_addresses_district ON addresses(district);
CREATE INDEX idx_addresses_state ON addresses(state);
CREATE INDEX idx_addresses_pincode ON addresses(pincode);

-- Contacts indexes
CREATE INDEX idx_contacts_user_id ON contacts(user_id);
CREATE INDEX idx_contacts_mobile_number ON contacts(mobile_number);

-- Roles indexes
CREATE INDEX idx_roles_name ON roles(name);
CREATE INDEX idx_roles_scope ON roles(scope);
CREATE INDEX idx_roles_organization_id ON roles(organization_id);
CREATE INDEX idx_roles_parent_id ON roles(parent_id);

-- Actions indexes
CREATE INDEX idx_actions_name ON actions(name);
CREATE INDEX idx_actions_category ON actions(category);
CREATE INDEX idx_actions_is_static ON actions(is_static);

-- Resources indexes
CREATE INDEX idx_resources_name ON resources(name);
CREATE INDEX idx_resources_type ON resources(type);
CREATE INDEX idx_resources_parent_id ON resources(parent_id);
CREATE INDEX idx_resources_owner_id ON resources(owner_id);

-- Permissions indexes
CREATE INDEX idx_permissions_name ON permissions(name);
CREATE INDEX idx_permissions_resource_id ON permissions(resource_id);
CREATE INDEX idx_permissions_action_id ON permissions(action_id);

-- User roles indexes
CREATE INDEX idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX idx_user_roles_role_id ON user_roles(role_id);

-- Role permissions indexes
CREATE INDEX idx_role_permissions_role_id ON role_permissions(role_id);
CREATE INDEX idx_role_permissions_permission_id ON role_permissions(permission_id);

-- Resource permissions indexes
CREATE INDEX idx_resource_permissions_role_id ON resource_permissions(role_id);
CREATE INDEX idx_resource_permissions_resource_type ON resource_permissions(resource_type);
CREATE INDEX idx_resource_permissions_resource_id ON resource_permissions(resource_id);
CREATE INDEX idx_resource_permissions_action ON resource_permissions(action);

-- Groups indexes
CREATE INDEX idx_groups_name ON groups(name);
CREATE INDEX idx_groups_organization_id ON groups(organization_id);
CREATE INDEX idx_groups_parent_id ON groups(parent_id);

-- Group memberships indexes
CREATE INDEX idx_group_memberships_group_id ON group_memberships(group_id);
CREATE INDEX idx_group_memberships_principal_id ON group_memberships(principal_id);
CREATE INDEX idx_group_memberships_principal_type ON group_memberships(principal_type);

-- Group inheritance indexes
CREATE INDEX idx_group_inheritance_parent_group_id ON group_inheritance(parent_group_id);
CREATE INDEX idx_group_inheritance_child_group_id ON group_inheritance(child_group_id);

-- Principals indexes
CREATE INDEX idx_principals_type ON principals(type);
CREATE INDEX idx_principals_user_id ON principals(user_id);
CREATE INDEX idx_principals_service_id ON principals(service_id);
CREATE INDEX idx_principals_organization_id ON principals(organization_id);

-- Services indexes
CREATE INDEX idx_services_name ON services(name);
CREATE INDEX idx_services_organization_id ON services(organization_id);
CREATE INDEX idx_services_api_key ON services(api_key);

-- Attributes indexes
CREATE INDEX idx_attributes_subject_id ON attributes(subject_id);
CREATE INDEX idx_attributes_subject_type ON attributes(subject_type);
CREATE INDEX idx_attributes_key ON attributes(key);
CREATE INDEX idx_attributes_organization_id ON attributes(organization_id);
CREATE INDEX idx_attributes_expires_at ON attributes(expires_at);

-- Attribute history indexes
CREATE INDEX idx_attribute_history_attribute_id ON attribute_history(attribute_id);
CREATE INDEX idx_attribute_history_subject_id ON attribute_history(subject_id);
CREATE INDEX idx_attribute_history_changed_at ON attribute_history(changed_at);

-- Bindings indexes
CREATE INDEX idx_bindings_subject_id ON bindings(subject_id);
CREATE INDEX idx_bindings_subject_type ON bindings(subject_type);
CREATE INDEX idx_bindings_binding_type ON bindings(binding_type);
CREATE INDEX idx_bindings_role_id ON bindings(role_id);
CREATE INDEX idx_bindings_permission_id ON bindings(permission_id);
CREATE INDEX idx_bindings_resource_type ON bindings(resource_type);
CREATE INDEX idx_bindings_organization_id ON bindings(organization_id);

-- Binding history indexes
CREATE INDEX idx_binding_history_binding_id ON binding_history(binding_id);
CREATE INDEX idx_binding_history_subject_id ON binding_history(subject_id);
CREATE INDEX idx_binding_history_changed_at ON binding_history(changed_at);

-- Column groups indexes
CREATE INDEX idx_column_groups_name ON column_groups(name);
CREATE INDEX idx_column_groups_table_name ON column_groups(table_name);
CREATE INDEX idx_column_groups_organization_id ON column_groups(organization_id);

-- Column group members indexes
CREATE INDEX idx_column_group_members_column_group_id ON column_group_members(column_group_id);
CREATE INDEX idx_column_group_members_column_name ON column_group_members(column_name);

-- Column sets indexes
CREATE INDEX idx_column_sets_name ON column_sets(name);
CREATE INDEX idx_column_sets_table_name ON column_sets(table_name);
CREATE INDEX idx_column_sets_organization_id ON column_sets(organization_id);

-- Audit logs indexes
CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_resource_type ON audit_logs(resource_type);
CREATE INDEX idx_audit_logs_status ON audit_logs(status);
CREATE INDEX idx_audit_logs_timestamp ON audit_logs(timestamp);

-- Events indexes
CREATE INDEX idx_events_occurred_at ON events(occurred_at);
CREATE INDEX idx_events_actor_id ON events(actor_id);
CREATE INDEX idx_events_actor_type ON events(actor_type);
CREATE INDEX idx_events_kind ON events(kind);
CREATE INDEX idx_events_resource_type ON events(resource_type);
CREATE INDEX idx_events_resource_id ON events(resource_id);
CREATE INDEX idx_events_organization_id ON events(organization_id);
CREATE INDEX idx_events_sequence_num ON events(sequence_num);

-- Event checkpoints indexes
CREATE INDEX idx_event_checkpoints_checkpoint_time ON event_checkpoints(checkpoint_time);
CREATE INDEX idx_event_checkpoints_last_event_id ON event_checkpoints(last_event_id);

-- =====================================================
-- TRIGGERS FOR UPDATED_AT
-- =====================================================

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for all tables
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_user_profiles_updated_at BEFORE UPDATE ON user_profiles FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_addresses_updated_at BEFORE UPDATE ON addresses FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_contacts_updated_at BEFORE UPDATE ON contacts FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_organizations_updated_at BEFORE UPDATE ON organizations FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_roles_updated_at BEFORE UPDATE ON roles FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_actions_updated_at BEFORE UPDATE ON actions FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_resources_updated_at BEFORE UPDATE ON resources FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_permissions_updated_at BEFORE UPDATE ON permissions FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_user_roles_updated_at BEFORE UPDATE ON user_roles FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_role_permissions_updated_at BEFORE UPDATE ON role_permissions FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_resource_permissions_updated_at BEFORE UPDATE ON resource_permissions FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_groups_updated_at BEFORE UPDATE ON groups FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_group_memberships_updated_at BEFORE UPDATE ON group_memberships FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_group_inheritance_updated_at BEFORE UPDATE ON group_inheritance FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_principals_updated_at BEFORE UPDATE ON principals FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_services_updated_at BEFORE UPDATE ON services FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_attributes_updated_at BEFORE UPDATE ON attributes FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_attribute_history_updated_at BEFORE UPDATE ON attribute_history FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_bindings_updated_at BEFORE UPDATE ON bindings FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_binding_history_updated_at BEFORE UPDATE ON binding_history FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_column_groups_updated_at BEFORE UPDATE ON column_groups FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_column_group_members_updated_at BEFORE UPDATE ON column_group_members FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_column_sets_updated_at BEFORE UPDATE ON column_sets FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_audit_logs_updated_at BEFORE UPDATE ON audit_logs FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_events_updated_at BEFORE UPDATE ON events FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_event_checkpoints_updated_at BEFORE UPDATE ON event_checkpoints FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- =====================================================
-- COMMENTS FOR DOCUMENTATION
-- =====================================================

COMMENT ON TABLE organizations IS 'Organizations in the AAA service with hierarchical support';
COMMENT ON TABLE users IS 'User accounts with phone number authentication and token system';
COMMENT ON TABLE user_profiles IS 'Extended user profile information including Aadhaar and contact details';
COMMENT ON TABLE addresses IS 'Physical addresses for users and organizations';
COMMENT ON TABLE contacts IS 'Contact information for users including mobile numbers and email hashes';
COMMENT ON TABLE roles IS 'Roles with scope (global/organization) and hierarchical support';
COMMENT ON TABLE actions IS 'Actions that can be performed on resources';
COMMENT ON TABLE resources IS 'Resources that can be protected by permissions';
COMMENT ON TABLE permissions IS 'Permissions linking roles, resources, and actions';
COMMENT ON TABLE user_roles IS 'Many-to-many relationship between users and roles';
COMMENT ON TABLE role_permissions IS 'Many-to-many relationship between roles and permissions';
COMMENT ON TABLE resource_permissions IS 'Direct resource-level permissions for roles';
COMMENT ON TABLE groups IS 'Groups for organizing users and roles within organizations';
COMMENT ON TABLE group_memberships IS 'User and service memberships in groups with time bounds';
COMMENT ON TABLE group_inheritance IS 'Hierarchical relationships between groups';
COMMENT ON TABLE principals IS 'Unified identity representation for users and services';
COMMENT ON TABLE services IS 'Service accounts with API key authentication';
COMMENT ON TABLE attributes IS 'Key-value attributes for ABAC (Attribute-Based Access Control)';
COMMENT ON TABLE attribute_history IS 'Audit trail for attribute changes';
COMMENT ON TABLE bindings IS 'Subject-to-role/permission bindings with caveats';
COMMENT ON TABLE binding_history IS 'Audit trail for binding changes';
COMMENT ON TABLE column_groups IS 'Named groups of columns for column-level permissions';
COMMENT ON TABLE column_group_members IS 'Columns belonging to column groups';
COMMENT ON TABLE column_sets IS 'Optimized bitmap representations of allowed columns';
COMMENT ON TABLE audit_logs IS 'Audit trail for all system activities';
COMMENT ON TABLE events IS 'Immutable event log for system state changes';
COMMENT ON TABLE event_checkpoints IS 'Periodic checkpoints of the event chain';

COMMENT ON COLUMN users.phone_number IS 'User phone number (unique identifier)';
COMMENT ON COLUMN users.country_code IS 'Country code for phone number (default: +91)';
COMMENT ON COLUMN users.username IS 'Optional username for user identification';
COMMENT ON COLUMN users.password IS 'Hashed password for authentication';
COMMENT ON COLUMN users.mpin IS 'Hashed MPIN for additional authentication';
COMMENT ON COLUMN users.is_validated IS 'Whether user has completed Aadhaar validation';
COMMENT ON COLUMN users.status IS 'User account status: pending, active, suspended, blocked';
COMMENT ON COLUMN users.tokens IS 'Available tokens for service usage';

COMMENT ON COLUMN roles.scope IS 'Role scope: GLOBAL (across all orgs) or ORG (organization-specific)';
COMMENT ON COLUMN roles.version IS 'Role version for role evolution tracking';
COMMENT ON COLUMN roles.metadata IS 'Additional role metadata in JSON format';

COMMENT ON COLUMN actions.category IS 'Action category: user, role, system, api, database, audit, general';
COMMENT ON COLUMN actions.is_static IS 'Whether action is built-in (true) or service-defined (false)';

COMMENT ON COLUMN resources.type IS 'Resource type following format: aaa/resource_name';
COMMENT ON COLUMN resources.owner_id IS 'User who owns this resource';

COMMENT ON COLUMN bindings.caveat IS 'JSON constraints for the binding (time limits, attributes, etc.)';
COMMENT ON COLUMN bindings.version IS 'Binding version for tracking changes';

COMMENT ON COLUMN events.hash IS 'SHA256 hash of event data for integrity verification';
COMMENT ON COLUMN events.sequence_num IS 'Monotonic sequence number for event ordering';
COMMENT ON COLUMN events.prev_hash IS 'Hash of previous event for chain verification';

COMMENT ON COLUMN column_sets.bitmap IS 'Binary representation of allowed columns (1=allowed, 0=denied)';
COMMENT ON COLUMN column_sets.column_count IS 'Total number of columns in the table';

-- =====================================================
-- MIGRATION COMPLETE
-- =====================================================
