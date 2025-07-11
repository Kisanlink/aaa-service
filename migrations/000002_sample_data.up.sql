-- Sample Data Migration for V1/V2 Compatibility Testing

-- Insert sample roles
INSERT INTO roles (id, name, description) VALUES
('rol_admin_001', 'admin', 'System administrator with full access'),
('rol_user_002', 'user', 'Standard user with basic access'),
('rol_viewer_003', 'viewer', 'Read-only access user'),
('rol_moderator_004', 'moderator', 'Content moderator with limited admin access');

-- Insert sample permissions
INSERT INTO permissions (id, resource, effect, actions) VALUES
('perm_users_001', 'users', 'allow', ARRAY['create', 'read', 'update', 'delete']),
('perm_users_002', 'users', 'allow', ARRAY['read', 'update']),
('perm_users_003', 'users', 'allow', ARRAY['read']),
('perm_roles_001', 'roles', 'allow', ARRAY['create', 'read', 'update', 'delete']),
('perm_roles_002', 'roles', 'allow', ARRAY['read']),
('perm_content_001', 'content', 'allow', ARRAY['create', 'read', 'update', 'delete', 'moderate']),
('perm_content_002', 'content', 'allow', ARRAY['create', 'read', 'update']),
('perm_content_003', 'content', 'allow', ARRAY['read']);

-- Link roles to permissions
INSERT INTO role_permissions (id, role_id, permission_id) VALUES
('rp_001', 'rol_admin_001', 'perm_users_001'),
('rp_002', 'rol_admin_001', 'perm_roles_001'),
('rp_003', 'rol_admin_001', 'perm_content_001'),
('rp_004', 'rol_user_002', 'perm_users_002'),
('rp_005', 'rol_user_002', 'perm_content_002'),
('rp_006', 'rol_viewer_003', 'perm_users_003'),
('rp_007', 'rol_viewer_003', 'perm_roles_002'),
('rp_008', 'rol_viewer_003', 'perm_content_003'),
('rp_009', 'rol_moderator_004', 'perm_users_003'),
('rp_010', 'rol_moderator_004', 'perm_content_001');

-- Insert sample addresses
INSERT INTO addresses (id, house, street, district, state, country, pincode) VALUES
('addr_001', '123', 'Main Street', 'Central District', 'Maharashtra', 'India', '400001'),
('addr_002', '456', 'Park Road', 'North District', 'Karnataka', 'India', '560001'),
('addr_003', '789', 'Lake View', 'South District', 'Tamil Nadu', 'India', '600001');

-- Insert sample users
INSERT INTO users (id, username, password, is_validated, status, tokens) VALUES
('usr_admin_001', 'admin', '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj/VF7aGdHuO', TRUE, 'active', 5000),
('usr_test_002', 'testuser', '$2a$12$EXROn55aiVVj1fCGHu7w/u2sxAiLM.TpOx/GlXlGl1zY1sX3QKQku', TRUE, 'active', 1000),
('usr_viewer_003', 'viewer', '$2a$12$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/XGF5v8GQlY5rUPE7G', TRUE, 'active', 500);

-- Insert sample user profiles
INSERT INTO user_profiles (id, user_id, name, address_id) VALUES
('prof_001', 'usr_admin_001', 'System Administrator', 'addr_001'),
('prof_002', 'usr_test_002', 'Test User', 'addr_002'),
('prof_003', 'usr_viewer_003', 'Viewer User', 'addr_003');

-- Insert sample contacts
INSERT INTO contacts (id, user_id, mobile_number, country_code, address_id) VALUES
('cont_001', 'usr_admin_001', 9876543210, '+91', 'addr_001'),
('cont_002', 'usr_test_002', 9876543211, '+91', 'addr_002'),
('cont_003', 'usr_viewer_003', 9876543212, '+91', 'addr_003');

-- Assign roles to users
INSERT INTO user_roles (id, user_id, role_id, is_active) VALUES
('ur_001', 'usr_admin_001', 'rol_admin_001', TRUE),
('ur_002', 'usr_test_002', 'rol_user_002', TRUE),
('ur_003', 'usr_viewer_003', 'rol_viewer_003', TRUE);

-- Insert sample organizations (V1 compatibility)
INSERT INTO organizations (id, name, description) VALUES
(uuid_generate_v4(), 'KisanLink', 'Main agricultural platform organization'),
(uuid_generate_v4(), 'FarmTech Solutions', 'Technology solutions for farmers'),
(uuid_generate_v4(), 'AgriSupport', 'Agricultural support services');

-- Insert sample v1 users
INSERT INTO users_v1 (id, email, first_name, last_name, phone) VALUES
(uuid_generate_v4(), 'admin@kisanlink.com', 'Admin', 'User', '+919876543210'),
(uuid_generate_v4(), 'farmer@kisanlink.com', 'Farmer', 'Singh', '+919876543211'),
(uuid_generate_v4(), 'support@kisanlink.com', 'Support', 'Team', '+919876543212');

-- Insert sample audit logs
INSERT INTO audit_logs (id, user_id, action, resource_type, resource_id, details) VALUES
('audit_001', 'usr_admin_001', 'CREATE', 'user', 'usr_test_002', '{"created_user": "testuser"}'),
('audit_002', 'usr_admin_001', 'ASSIGN_ROLE', 'user_role', 'ur_002', '{"role": "user", "user": "testuser"}'),
('audit_003', 'usr_test_002', 'LOGIN', 'session', 'sess_001', '{"login_time": "2024-01-01T00:00:00Z"}'),
('audit_004', 'usr_viewer_003', 'VIEW', 'users', 'usr_test_002', '{"viewed_user": "testuser"}');
