CREATE EXTENSION IF NOT EXISTS "ltree";

-- Create the parent table for organizations
CREATE TABLE organizations (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    parent_id UUID,
    path ltree,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ,
    created_by UUID,
    updated_by UUID,
    deleted_by UUID
) PARTITION BY HASH (id);

-- Create partitions (adjust the number based on your needs)
CREATE TABLE organizations_p0 PARTITION OF organizations FOR VALUES WITH (modulus 4, remainder 0);
CREATE TABLE organizations_p1 PARTITION OF organizations FOR VALUES WITH (modulus 4, remainder 1);
CREATE TABLE organizations_p2 PARTITION OF organizations FOR VALUES WITH (modulus 4, remainder 2);
CREATE TABLE organizations_p3 PARTITION OF organizations FOR VALUES WITH (modulus 4, remainder 3);

-- Add foreign key constraint after partitioning
ALTER TABLE organizations
ADD CONSTRAINT fk_organizations_parent
FOREIGN KEY (parent_id) REFERENCES organizations(id);

-- Function to prevent cycles
CREATE OR REPLACE FUNCTION check_organization_cycle()
RETURNS TRIGGER AS $$
BEGIN
    IF EXISTS (
        WITH RECURSIVE org_tree AS (
            SELECT id, parent_id, ARRAY[id] as path
            FROM organizations
            WHERE id = NEW.parent_id
            UNION ALL
            SELECT o.id, o.parent_id, ot.path || o.id
            FROM organizations o
            JOIN org_tree ot ON o.id = ot.parent_id
            WHERE NOT o.id = ANY(ot.path)
        )
        SELECT 1 FROM org_tree WHERE id = NEW.id
    ) THEN
        RAISE EXCEPTION 'Cycle detected in organization hierarchy';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER prevent_organization_cycle
    BEFORE INSERT OR UPDATE ON organizations
    FOR EACH ROW
    EXECUTE FUNCTION check_organization_cycle();

-- Indexes for each partition
CREATE INDEX organizations_path_idx_p0 ON organizations_p0 USING GIST (path);
CREATE INDEX organizations_path_idx_p1 ON organizations_p1 USING GIST (path);
CREATE INDEX organizations_path_idx_p2 ON organizations_p2 USING GIST (path);
CREATE INDEX organizations_path_idx_p3 ON organizations_p3 USING GIST (path);

-- Indexes for parent_id lookups in each partition
CREATE INDEX organizations_parent_id_idx_p0 ON organizations_p0(parent_id);
CREATE INDEX organizations_parent_id_idx_p1 ON organizations_p1(parent_id);
CREATE INDEX organizations_parent_id_idx_p2 ON organizations_p2(parent_id);
CREATE INDEX organizations_parent_id_idx_p3 ON organizations_p3(parent_id); 