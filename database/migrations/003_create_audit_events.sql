-- Create the parent table for audit events
CREATE TABLE audit_events (
    id UUID PRIMARY KEY,
    org_id UUID,
    actor_id UUID NOT NULL,
    actor_ip INET,
    action TEXT NOT NULL,
    resource_type TEXT NOT NULL,
    resource_id UUID NOT NULL,
    old_values JSONB,
    new_values JSONB,
    correlation_id UUID,
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
) PARTITION BY HASH (org_id);

-- Create partitions (adjust the number based on your needs)
CREATE TABLE audit_events_p0 PARTITION OF audit_events FOR VALUES WITH (modulus 4, remainder 0);
CREATE TABLE audit_events_p1 PARTITION OF audit_events FOR VALUES WITH (modulus 4, remainder 1);
CREATE TABLE audit_events_p2 PARTITION OF audit_events FOR VALUES WITH (modulus 4, remainder 2);
CREATE TABLE audit_events_p3 PARTITION OF audit_events FOR VALUES WITH (modulus 4, remainder 3);

-- Add foreign key constraint after partitioning
ALTER TABLE audit_events
ADD CONSTRAINT fk_audit_events_org
FOREIGN KEY (org_id) REFERENCES organizations(id);

-- Indexes for each partition
CREATE INDEX audit_events_created_at_idx_p0 ON audit_events_p0(created_at);
CREATE INDEX audit_events_created_at_idx_p1 ON audit_events_p1(created_at);
CREATE INDEX audit_events_created_at_idx_p2 ON audit_events_p2(created_at);
CREATE INDEX audit_events_created_at_idx_p3 ON audit_events_p3(created_at);

CREATE INDEX audit_events_correlation_id_idx_p0 ON audit_events_p0(correlation_id);
CREATE INDEX audit_events_correlation_id_idx_p1 ON audit_events_p1(correlation_id);
CREATE INDEX audit_events_correlation_id_idx_p2 ON audit_events_p2(correlation_id);
CREATE INDEX audit_events_correlation_id_idx_p3 ON audit_events_p3(correlation_id);

-- Create a default partition for system-level audit events (where org_id is NULL)
CREATE TABLE audit_events_default PARTITION OF audit_events DEFAULT; 