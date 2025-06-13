-- Add validity ranges to user_roles
ALTER TABLE user_roles 
ADD COLUMN validity tsrange;

CREATE INDEX user_roles_validity_idx ON user_roles USING GIST (validity);

-- Add exclusion constraint to prevent overlapping ranges
ALTER TABLE user_roles
ADD CONSTRAINT user_roles_no_overlap 
EXCLUDE USING GIST (
    user_id WITH =,
    role_id WITH =,
    validity WITH &&
) WHERE (deleted_at IS NULL);

-- Add audit fields to users
ALTER TABLE users
ADD COLUMN created_by UUID,
ADD COLUMN updated_by UUID,
ADD COLUMN deleted_by UUID;

-- Add audit fields to roles
ALTER TABLE roles
ADD COLUMN created_by UUID,
ADD COLUMN updated_by UUID,
ADD COLUMN deleted_by UUID;

-- Add audit fields to permissions
ALTER TABLE permissions
ADD COLUMN created_by UUID,
ADD COLUMN updated_by UUID,
ADD COLUMN deleted_by UUID;

-- Add audit fields to resources
ALTER TABLE resources
ADD COLUMN created_by UUID,
ADD COLUMN updated_by UUID,
ADD COLUMN deleted_by UUID;

-- Add audit fields to actions
ALTER TABLE actions
ADD COLUMN created_by UUID,
ADD COLUMN updated_by UUID,
ADD COLUMN deleted_by UUID;

-- Add audit fields to role_permissions
ALTER TABLE role_permissions
ADD COLUMN created_by UUID,
ADD COLUMN updated_by UUID,
ADD COLUMN deleted_by UUID;

-- Add audit fields to addresses
ALTER TABLE addresses
ADD COLUMN created_by UUID,
ADD COLUMN updated_by UUID,
ADD COLUMN deleted_by UUID;

-- Add partial index for active roles
CREATE INDEX user_roles_active_idx ON user_roles(user_id, role_id) 
WHERE (validity @> now()::timestamp AND deleted_at IS NULL); 