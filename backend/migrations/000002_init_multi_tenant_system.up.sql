-- =====================================================
-- RBAC 改造：添加 user_roles、role_permissions 表
-- 为 roles 表添加 parent_role_id 列（角色继承）
-- =====================================================

-- 1. roles 表添加 parent_role_id 列（替代 Casbin g2 策略）
ALTER TABLE roles ADD COLUMN IF NOT EXISTS parent_role_id VARCHAR(20) NOT NULL DEFAULT '';

-- 2. 用户角色关联表（替代 Casbin g 策略）
CREATE TABLE IF NOT EXISTS user_roles (
    id         BIGSERIAL    PRIMARY KEY,
    user_id    VARCHAR(20)  NOT NULL,
    role_id    VARCHAR(20)  NOT NULL,
    tenant_id  VARCHAR(20)  NOT NULL,
    created_at BIGINT       NOT NULL DEFAULT 0,
    UNIQUE(user_id, role_id, tenant_id)
);
CREATE INDEX IF NOT EXISTS idx_user_roles_user_tenant ON user_roles(user_id, tenant_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_role ON user_roles(role_id);

-- 3. 角色权限关联表（替代 Casbin p 策略）
CREATE TABLE IF NOT EXISTS role_permissions (
    id            BIGSERIAL    PRIMARY KEY,
    role_id       VARCHAR(20)  NOT NULL,
    permission_id VARCHAR(20)  NOT NULL,
    tenant_id     VARCHAR(20)  NOT NULL,
    created_at    BIGINT       NOT NULL DEFAULT 0,
    UNIQUE(role_id, permission_id, tenant_id)
);
CREATE INDEX IF NOT EXISTS idx_role_permissions_role ON role_permissions(role_id, tenant_id);
CREATE INDEX IF NOT EXISTS idx_role_permissions_permission ON role_permissions(permission_id);
