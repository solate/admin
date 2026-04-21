-- 回滚 RBAC 改造

DROP TABLE IF EXISTS role_permissions;
DROP TABLE IF EXISTS user_roles;
ALTER TABLE roles DROP COLUMN IF EXISTS parent_role_id;
