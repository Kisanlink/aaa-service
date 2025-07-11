-- Remove Sample Data

-- Remove audit logs
DELETE FROM audit_logs WHERE id IN ('audit_001', 'audit_002', 'audit_003', 'audit_004');

-- Remove V1 sample data
DELETE FROM users_v1 WHERE email IN ('admin@kisanlink.com', 'farmer@kisanlink.com', 'support@kisanlink.com');
DELETE FROM organizations WHERE name IN ('KisanLink', 'FarmTech Solutions', 'AgriSupport');

-- Remove V2 sample data (in reverse dependency order)
DELETE FROM user_roles WHERE id IN ('ur_001', 'ur_002', 'ur_003');
DELETE FROM contacts WHERE id IN ('cont_001', 'cont_002', 'cont_003');
DELETE FROM user_profiles WHERE id IN ('prof_001', 'prof_002', 'prof_003');
DELETE FROM users WHERE id IN ('usr_admin_001', 'usr_test_002', 'usr_viewer_003');
DELETE FROM addresses WHERE id IN ('addr_001', 'addr_002', 'addr_003');
DELETE FROM role_permissions WHERE id IN ('rp_001', 'rp_002', 'rp_003', 'rp_004', 'rp_005', 'rp_006', 'rp_007', 'rp_008', 'rp_009', 'rp_010');
DELETE FROM permissions WHERE id IN ('perm_users_001', 'perm_users_002', 'perm_users_003', 'perm_roles_001', 'perm_roles_002', 'perm_content_001', 'perm_content_002', 'perm_content_003');
DELETE FROM roles WHERE id IN ('rol_admin_001', 'rol_user_002', 'rol_viewer_003', 'rol_moderator_004');
