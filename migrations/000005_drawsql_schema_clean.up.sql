-- AAA Service Database Schema (Clean Version)
-- PostgreSQL DDL optimized for drawSQL visualization
-- No sample data, only schema definitions

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
    vtc VARCHAR(255),
    state VARCHAR(255),
    country VARCHAR(255),
    pincode VARCHAR(10),
    full_address TEXT
);

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
    tokens INTEGER DEFAULT 1000
);

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
    actions TEXT[]
);

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

-- Role Permissions junction table (Many-to-Many between Roles and Permissions)
CREATE TABLE role_permissions (
    role_id VARCHAR(255) NOT NULL,
    permission_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    PRIMARY KEY (role_id, permission_id),
    CONSTRAINT fk_role_permissions_role_id FOREIGN KEY (role_id) REFERENCES roles(id) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_role_permissions_permission_id FOREIGN KEY (permission_id) REFERENCES permissions(id) ON UPDATE CASCADE ON DELETE CASCADE
);

-- =============================================================================
-- Indexes for Performance
-- =============================================================================

-- Soft delete indexes
CREATE INDEX idx_addresses_deleted_at ON addresses(deleted_at);
CREATE INDEX idx_users_deleted_at ON users(deleted_at);
CREATE INDEX idx_roles_deleted_at ON roles(deleted_at);
CREATE INDEX idx_permissions_deleted_at ON permissions(deleted_at);
CREATE INDEX idx_user_profiles_deleted_at ON user_profiles(deleted_at);
CREATE INDEX idx_contacts_deleted_at ON contacts(deleted_at);
CREATE INDEX idx_user_roles_deleted_at ON user_roles(deleted_at);

-- Business logic indexes
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_status ON users(status);
CREATE UNIQUE INDEX idx_roles_name ON roles(name);
CREATE INDEX idx_permissions_resource ON permissions(resource);
CREATE UNIQUE INDEX idx_user_profiles_user_id ON user_profiles(user_id);
CREATE INDEX idx_user_profiles_address_id ON user_profiles(address_id);
CREATE INDEX idx_contacts_user_id ON contacts(user_id);
CREATE INDEX idx_contacts_mobile_number ON contacts(mobile_number);
CREATE INDEX idx_contacts_address_id ON contacts(address_id);
CREATE INDEX idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX idx_user_roles_role_id ON user_roles(role_id);
CREATE INDEX idx_user_roles_is_active ON user_roles(is_active);
CREATE INDEX idx_role_permissions_role_id ON role_permissions(role_id);
CREATE INDEX idx_role_permissions_permission_id ON role_permissions(permission_id);
