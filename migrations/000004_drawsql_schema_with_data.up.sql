-- AAA Service Database Schema
-- PostgreSQL DDL for drawSQL visualization
-- Generated from Go models in aaa-service

-- Enable necessary extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Drop tables if they exist (reverse dependency order)
DROP TABLE IF EXISTS role_permissions CASCADE;
DROP TABLE IF EXISTS user_roles CASCADE;
DROP TABLE IF EXISTS contacts CASCADE;
DROP TABLE IF EXISTS user_profiles CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS permissions CASCADE;
DROP TABLE IF EXISTS roles CASCADE;
DROP TABLE IF EXISTS addresses CASCADE;

-- =============================================================================
-- Core Tables
-- =============================================================================

-- Addresses table
CREATE TABLE addresses (
    id VARCHAR(255) PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255),
    updated_by VARCHAR(255),
    deleted_at TIMESTAMP NULL,
    deleted_by VARCHAR(255),
    house VARCHAR(255),
    street VARCHAR(255),
    landmark VARCHAR(255),
    post_office VARCHAR(255),
    subdistrict VARCHAR(255),
    district VARCHAR(255),
    vtc VARCHAR(255), -- Village/Town/City
    state VARCHAR(255),
    country VARCHAR(255),
    pincode VARCHAR(10),
    full_address TEXT
);

-- Create index for soft deletes
CREATE INDEX idx_addresses_deleted_at ON addresses(deleted_at);

-- Users table
CREATE TABLE users (
    id VARCHAR(255) PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255),
    updated_by VARCHAR(255),
    deleted_at TIMESTAMP NULL,
    deleted_by VARCHAR(255),
    username VARCHAR(100) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    is_validated BOOLEAN DEFAULT FALSE,
    status VARCHAR(50) DEFAULT 'pending',
    tokens INTEGER DEFAULT 1000,

    CONSTRAINT chk_user_status CHECK (status IN ('pending', 'active', 'suspended', 'blocked'))
);

-- Create indexes
CREATE INDEX idx_users_deleted_at ON users(deleted_at);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_status ON users(status);

-- Roles table
CREATE TABLE roles (
    id VARCHAR(255) PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255),
    updated_by VARCHAR(255),
    deleted_at TIMESTAMP NULL,
    deleted_by VARCHAR(255),
    name VARCHAR(50) UNIQUE NOT NULL,
    description TEXT
);

-- Create indexes
CREATE INDEX idx_roles_deleted_at ON roles(deleted_at);
CREATE UNIQUE INDEX idx_roles_name ON roles(name);

-- Permissions table
CREATE TABLE permissions (
    id VARCHAR(255) PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255),
    updated_by VARCHAR(255),
    deleted_at TIMESTAMP NULL,
    deleted_by VARCHAR(255),
    resource VARCHAR(100) NOT NULL,
    effect TEXT,
    actions TEXT[] -- PostgreSQL array for actions
);

-- Create indexes
CREATE INDEX idx_permissions_deleted_at ON permissions(deleted_at);
CREATE INDEX idx_permissions_resource ON permissions(resource);

-- =============================================================================
-- Relationship Tables
-- =============================================================================

-- User Profiles table (One-to-One with Users)
CREATE TABLE user_profiles (
    id VARCHAR(255) PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255),
    updated_by VARCHAR(255),
    deleted_at TIMESTAMP NULL,
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

    CONSTRAINT fk_user_profiles_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_user_profiles_address_id FOREIGN KEY (address_id) REFERENCES addresses(id) ON UPDATE CASCADE ON DELETE SET NULL
);

-- Create indexes
CREATE INDEX idx_user_profiles_deleted_at ON user_profiles(deleted_at);
CREATE UNIQUE INDEX idx_user_profiles_user_id ON user_profiles(user_id);
CREATE INDEX idx_user_profiles_address_id ON user_profiles(address_id);

-- Contacts table (One-to-Many with Users)
CREATE TABLE contacts (
    id VARCHAR(255) PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255),
    updated_by VARCHAR(255),
    deleted_at TIMESTAMP NULL,
    deleted_by VARCHAR(255),
    user_id VARCHAR(255) NOT NULL,
    mobile_number BIGINT NOT NULL,
    country_code VARCHAR(10) DEFAULT '+91',
    email_hash VARCHAR(255),
    share_code VARCHAR(50),
    address_id VARCHAR(255),

    CONSTRAINT fk_contacts_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_contacts_address_id FOREIGN KEY (address_id) REFERENCES addresses(id) ON UPDATE CASCADE ON DELETE SET NULL
);

-- Create indexes
CREATE INDEX idx_contacts_deleted_at ON contacts(deleted_at);
CREATE INDEX idx_contacts_user_id ON contacts(user_id);
CREATE INDEX idx_contacts_mobile_number ON contacts(mobile_number);
CREATE INDEX idx_contacts_address_id ON contacts(address_id);

-- User Roles junction table (Many-to-Many between Users and Roles)
CREATE TABLE user_roles (
    id VARCHAR(255) PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255),
    updated_by VARCHAR(255),
    deleted_at TIMESTAMP NULL,
    deleted_by VARCHAR(255),
    user_id VARCHAR(255) NOT NULL,
    role_id VARCHAR(255) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,

    CONSTRAINT fk_user_roles_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_user_roles_role_id FOREIGN KEY (role_id) REFERENCES roles(id) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT uk_user_roles_user_role UNIQUE (user_id, role_id)
);

-- Create indexes
CREATE INDEX idx_user_roles_deleted_at ON user_roles(deleted_at);
CREATE INDEX idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX idx_user_roles_role_id ON user_roles(role_id);
CREATE INDEX idx_user_roles_is_active ON user_roles(is_active);

-- Role Permissions junction table (Many-to-Many between Roles and Permissions)
CREATE TABLE role_permissions (
    role_id VARCHAR(255) NOT NULL,
    permission_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    PRIMARY KEY (role_id, permission_id),
    CONSTRAINT fk_role_permissions_role_id FOREIGN KEY (role_id) REFERENCES roles(id) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_role_permissions_permission_id FOREIGN KEY (permission_id) REFERENCES permissions(id) ON UPDATE CASCADE ON DELETE CASCADE
);

-- Create indexes
CREATE INDEX idx_role_permissions_role_id ON role_permissions(role_id);
CREATE INDEX idx_role_permissions_permission_id ON role_permissions(permission_id);

-- =============================================================================
-- Sample Data (Optional - for testing)
-- =============================================================================

-- Insert sample address
INSERT INTO addresses (id, house, street, district, state, country, pincode) VALUES
('ADDR001', '123', 'Main Street', 'Central District', 'Maharashtra', 'India', '400001');

-- Insert sample roles
INSERT INTO roles (id, name, description) VALUES
('ROLE001', 'admin', 'Administrator with full access'),
('ROLE002', 'user', 'Regular user with limited access'),
('ROLE003', 'moderator', 'Moderator with intermediate access');

-- Insert sample permissions
INSERT INTO permissions (id, resource, effect, actions) VALUES
('PERM001', 'users', 'allow', ARRAY['create', 'read', 'update', 'delete']),
('PERM002', 'users', 'allow', ARRAY['read', 'update']),
('PERM003', 'users', 'allow', ARRAY['read']);

-- Insert sample user
INSERT INTO users (id, username, password, status) VALUES
('USR001', 'admin_user', '$2a$10$hash...', 'active');

-- Insert sample user profile
INSERT INTO user_profiles (id, user_id, name, address_id) VALUES
('PROF001', 'USR001', 'Admin User', 'ADDR001');

-- Insert sample contact
INSERT INTO contacts (id, user_id, mobile_number, address_id) VALUES
('CONT001', 'USR001', 9876543210, 'ADDR001');

-- Link user to role
INSERT INTO user_roles (id, user_id, role_id) VALUES
('UR001', 'USR001', 'ROLE001');

-- Link role to permissions
INSERT INTO role_permissions (role_id, permission_id) VALUES
('ROLE001', 'PERM001'),
('ROLE002', 'PERM002'),
('ROLE003', 'PERM003');

-- =============================================================================
-- Views for easier querying
-- =============================================================================

-- View for active users with their profiles
CREATE VIEW active_users_with_profiles AS
SELECT
    u.id as user_id,
    u.username,
    u.status,
    u.tokens,
    up.name,
    up.aadhaar_number,
    a.district,
    a.state,
    a.country
FROM users u
LEFT JOIN user_profiles up ON u.id = up.user_id
LEFT JOIN addresses a ON up.address_id = a.id
WHERE u.deleted_at IS NULL
  AND u.status = 'active';

-- View for user roles and permissions
CREATE VIEW user_permissions AS
SELECT
    u.id as user_id,
    u.username,
    r.name as role_name,
    p.resource,
    p.effect,
    p.actions
FROM users u
JOIN user_roles ur ON u.id = ur.user_id
JOIN roles r ON ur.role_id = r.id
JOIN role_permissions rp ON r.id = rp.role_id
JOIN permissions p ON rp.permission_id = p.id
WHERE u.deleted_at IS NULL
  AND ur.deleted_at IS NULL
  AND ur.is_active = TRUE;

-- =============================================================================
-- Comments for documentation
-- =============================================================================

COMMENT ON TABLE users IS 'Core user accounts with authentication information';
COMMENT ON TABLE user_profiles IS 'Extended user profile information including personal details';
COMMENT ON TABLE addresses IS 'Physical addresses that can be shared across users and contacts';
COMMENT ON TABLE contacts IS 'Contact information for users including mobile numbers';
COMMENT ON TABLE roles IS 'System roles that define user capabilities';
COMMENT ON TABLE permissions IS 'Granular permissions for specific resources and actions';
COMMENT ON TABLE user_roles IS 'Junction table linking users to their assigned roles';
COMMENT ON TABLE role_permissions IS 'Junction table linking roles to their permissions';

COMMENT ON COLUMN users.status IS 'User account status: pending, active, suspended, blocked';
COMMENT ON COLUMN users.tokens IS 'User token balance for paid operations';
COMMENT ON COLUMN permissions.actions IS 'Array of allowed actions for this permission';
COMMENT ON COLUMN addresses.vtc IS 'Village/Town/City designation';
