-- ========================================
-- 开发环境数据库初始化脚本
-- 包含所有表结构（优化版）
-- ========================================


-- ========================================
-- 1. 用户表 (全局用户)
-- ========================================
CREATE TABLE users (
    user_id VARCHAR(255) PRIMARY KEY,
    user_name VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL, -- bcrypt
    name VARCHAR(255) NOT NULL DEFAULT '',
    avatar VARCHAR(255),
    phone VARCHAR(20),
    email VARCHAR(255),
    status INTEGER NOT NULL DEFAULT 1, -- 1:启用, 2:禁用
    remark TEXT,
    last_login_time BIGINT,
    created_at BIGINT NOT NULL DEFAULT 0,
    updated_at BIGINT NOT NULL DEFAULT 0,
    deleted_at BIGINT DEFAULT 0
);
CREATE UNIQUE INDEX idx_users_username ON users(user_name);

COMMENT ON TABLE users IS '用户表';
COMMENT ON COLUMN users.user_id IS '用户ID';
COMMENT ON COLUMN users.user_name IS '用户名(登录账号)';
COMMENT ON COLUMN users.password IS '加密密码';
COMMENT ON COLUMN users.name IS '姓名/昵称';
COMMENT ON COLUMN users.avatar IS '头像URL';
COMMENT ON COLUMN users.phone IS '手机号';
COMMENT ON COLUMN users.email IS '电子邮箱';
COMMENT ON COLUMN users.status IS '状态(1:启用, 2:禁用)';
COMMENT ON COLUMN users.remark IS '备注信息';
COMMENT ON COLUMN users.last_login_time IS '最后登录时间戳';
COMMENT ON COLUMN users.created_at IS '创建时间戳';
COMMENT ON COLUMN users.updated_at IS '更新时间戳';
COMMENT ON COLUMN users.deleted_at IS '删除时间戳(软删除)';

-- ========================================
-- 2. 租户表
-- ========================================
CREATE TABLE tenants (
    tenant_id VARCHAR(255) PRIMARY KEY,
    tenant_name VARCHAR(200) NOT NULL,
    description TEXT,
    status INTEGER NOT NULL DEFAULT 1, -- 1:启用, 2:禁用
    created_at BIGINT NOT NULL DEFAULT 0,
    updated_at BIGINT NOT NULL DEFAULT 0,
    deleted_at BIGINT DEFAULT 0
);
-- CREATE UNIQUE INDEX idx_tenants_code ON tenants(tenant_code);

COMMENT ON TABLE tenants IS '租户表';
COMMENT ON COLUMN tenants.tenant_id IS '租户ID';
COMMENT ON COLUMN tenants.tenant_name IS '租户名称';
COMMENT ON COLUMN tenants.description IS '租户描述';
COMMENT ON COLUMN tenants.status IS '状态(1:启用, 2:禁用)';
COMMENT ON COLUMN tenants.created_at IS '创建时间戳';
COMMENT ON COLUMN tenants.updated_at IS '更新时间戳';
COMMENT ON COLUMN tenants.deleted_at IS '删除时间戳(软删除)';

-- ========================================
-- 3. 用户-租户关联表
-- ========================================
CREATE TABLE user_tenants (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    tenant_id VARCHAR(255) NOT NULL REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    is_admin BOOLEAN DEFAULT FALSE, -- 是否为该租户管理员
    status INTEGER NOT NULL DEFAULT 1,
    created_at BIGINT NOT NULL DEFAULT 0,
    updated_at BIGINT NOT NULL DEFAULT 0,
    UNIQUE(user_id, tenant_id)
);
CREATE INDEX idx_user_tenants_uid ON user_tenants(user_id);

COMMENT ON TABLE user_tenants IS '用户-租户关联表';
COMMENT ON COLUMN user_tenants.id IS '关联ID';
COMMENT ON COLUMN user_tenants.user_id IS '用户ID';
COMMENT ON COLUMN user_tenants.tenant_id IS '租户ID';
COMMENT ON COLUMN user_tenants.is_admin IS '是否为租户管理员';
COMMENT ON COLUMN user_tenants.status IS '状态(1:正常, 2:冻结)';
COMMENT ON COLUMN user_tenants.created_at IS '创建时间戳';
COMMENT ON COLUMN user_tenants.updated_at IS '更新时间戳';

-- ========================================
-- 4. 权限表 (系统功能定义)
-- ========================================
CREATE TABLE permissions (
    permission_id VARCHAR(255) PRIMARY KEY,
    parent_id VARCHAR(255) DEFAULT '',
    code VARCHAR(100) NOT NULL, -- 权限标识 e.g. 'sys:user:list'
    name VARCHAR(100) NOT NULL,
    type VARCHAR(20) NOT NULL, -- 'MENU', 'BUTTON', 'API'
    path VARCHAR(255), -- 路由路径 or API路径
    method VARCHAR(20), -- API方法 GET/POST...
    sort INTEGER DEFAULT 0,
    description TEXT,
    created_at BIGINT NOT NULL DEFAULT 0,
    updated_at BIGINT NOT NULL DEFAULT 0
);
CREATE UNIQUE INDEX idx_permissions_code ON permissions(code);

COMMENT ON TABLE permissions IS '权限/菜单定义表';
COMMENT ON COLUMN permissions.permission_id IS '权限ID';
COMMENT ON COLUMN permissions.parent_id IS '父权限ID';
COMMENT ON COLUMN permissions.code IS '权限标识(如 sys:user:list)';
COMMENT ON COLUMN permissions.name IS '权限名称';
COMMENT ON COLUMN permissions.type IS '类型(MENU:菜单, BUTTON:按钮, API:接口)';
COMMENT ON COLUMN permissions.path IS '路由路径或API路径';
COMMENT ON COLUMN permissions.method IS 'HTTP方法(仅API类型有效)';
COMMENT ON COLUMN permissions.sort IS '排序号';
COMMENT ON COLUMN permissions.description IS '描述信息';
COMMENT ON COLUMN permissions.created_at IS '创建时间戳';
COMMENT ON COLUMN permissions.updated_at IS '更新时间戳';

-- ========================================
-- 5. 角色表 (租户隔离)
-- ========================================
CREATE TABLE roles (
    role_id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    role_name VARCHAR(100) NOT NULL,
    description TEXT,
    status INTEGER NOT NULL DEFAULT 1,
    created_at BIGINT NOT NULL DEFAULT 0,
    updated_at BIGINT NOT NULL DEFAULT 0,
    deleted_at BIGINT DEFAULT 0
);

COMMENT ON TABLE roles IS '角色表';
COMMENT ON COLUMN roles.role_id IS '角色ID';
COMMENT ON COLUMN roles.tenant_id IS '所属租户ID';
COMMENT ON COLUMN roles.role_name IS '角色名称';
COMMENT ON COLUMN roles.description IS '角色描述';
COMMENT ON COLUMN roles.status IS '状态(1:启用, 2:禁用)';
COMMENT ON COLUMN roles.created_at IS '创建时间戳';
COMMENT ON COLUMN roles.updated_at IS '更新时间戳';
COMMENT ON COLUMN roles.deleted_at IS '删除时间戳(软删除)';

-- ========================================
-- 6. 角色-权限关联表
-- ========================================
CREATE TABLE role_permissions (
    id VARCHAR(255) PRIMARY KEY,
    role_id VARCHAR(255) NOT NULL REFERENCES roles(role_id) ON DELETE CASCADE,
    permission_id VARCHAR(255) NOT NULL REFERENCES permissions(permission_id) ON DELETE CASCADE,
    created_at BIGINT NOT NULL DEFAULT 0,
    UNIQUE(role_id, permission_id)
);

COMMENT ON TABLE role_permissions IS '角色-权限关联表';
COMMENT ON COLUMN role_permissions.id IS '关联ID';
COMMENT ON COLUMN role_permissions.role_id IS '角色ID';
COMMENT ON COLUMN role_permissions.permission_id IS '权限ID';
COMMENT ON COLUMN role_permissions.created_at IS '创建时间戳';

-- ========================================
-- 7. 用户组表 (租户隔离)
-- ========================================
CREATE TABLE user_groups (
    group_id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    group_name VARCHAR(100) NOT NULL,
    parent_id VARCHAR(255) DEFAULT '',
    description TEXT,
    status INTEGER NOT NULL DEFAULT 1,
    created_at BIGINT NOT NULL DEFAULT 0,
    updated_at BIGINT NOT NULL DEFAULT 0,
    deleted_at BIGINT DEFAULT 0
);

COMMENT ON TABLE user_groups IS '用户组表';
COMMENT ON COLUMN user_groups.group_id IS '用户组ID';
COMMENT ON COLUMN user_groups.tenant_id IS '所属租户ID';
COMMENT ON COLUMN user_groups.group_name IS '用户组名称';
COMMENT ON COLUMN user_groups.parent_id IS '父用户组ID';
COMMENT ON COLUMN user_groups.description IS '描述信息';
COMMENT ON COLUMN user_groups.status IS '状态(1:启用, 2:禁用)';
COMMENT ON COLUMN user_groups.created_at IS '创建时间戳';
COMMENT ON COLUMN user_groups.updated_at IS '更新时间戳';
COMMENT ON COLUMN user_groups.deleted_at IS '删除时间戳(软删除)';

-- ========================================
-- 8. 用户组-角色关联表 (组拥有角色)
-- ========================================
CREATE TABLE group_roles (
    id VARCHAR(255) PRIMARY KEY,
    group_id VARCHAR(255) NOT NULL REFERENCES user_groups(group_id) ON DELETE CASCADE,
    role_id VARCHAR(255) NOT NULL REFERENCES roles(role_id) ON DELETE CASCADE,
    created_at BIGINT NOT NULL DEFAULT 0,
    UNIQUE(group_id, role_id)
);

COMMENT ON TABLE group_roles IS '用户组-角色关联表';
COMMENT ON COLUMN group_roles.id IS '关联ID';
COMMENT ON COLUMN group_roles.group_id IS '用户组ID';
COMMENT ON COLUMN group_roles.role_id IS '角色ID';
COMMENT ON COLUMN group_roles.created_at IS '创建时间戳';

-- ========================================
-- 9. 用户组-用户关联表 (用户加入组)
-- ========================================
CREATE TABLE group_users (
    id VARCHAR(255) PRIMARY KEY,
    group_id VARCHAR(255) NOT NULL REFERENCES user_groups(group_id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    created_at BIGINT NOT NULL DEFAULT 0,
    UNIQUE(group_id, user_id)
);

COMMENT ON TABLE group_users IS '用户组-用户关联表';
COMMENT ON COLUMN group_users.id IS '关联ID';
COMMENT ON COLUMN group_users.group_id IS '用户组ID';
COMMENT ON COLUMN group_users.user_id IS '用户ID';
COMMENT ON COLUMN group_users.created_at IS '创建时间戳';

-- ========================================
-- 10. Casbin 策略表 (运行时权限控制)
-- ========================================
CREATE TABLE casbin_rules (
    id SERIAL PRIMARY KEY,
    ptype VARCHAR(100),
    v0 VARCHAR(100),
    v1 VARCHAR(100),
    v2 VARCHAR(100),
    v3 VARCHAR(100),
    v4 VARCHAR(100),
    v5 VARCHAR(100)
);
CREATE INDEX idx_casbin_ptype ON casbin_rules(ptype);
CREATE INDEX idx_casbin_v0 ON casbin_rules(v0);
CREATE INDEX idx_casbin_v1 ON casbin_rules(v1);
CREATE INDEX idx_casbin_v2 ON casbin_rules(v2);
CREATE INDEX idx_casbin_v3 ON casbin_rules(v3);

COMMENT ON TABLE casbin_rules IS 'Casbin权限策略表 (支持 RBAC with Domains)';
COMMENT ON COLUMN casbin_rules.id IS '主键ID';
COMMENT ON COLUMN casbin_rules.ptype IS '策略类型: p(policy) / g(group)';
COMMENT ON COLUMN casbin_rules.v0 IS 'v0: sub (主体: 用户ID/角色ID)';
COMMENT ON COLUMN casbin_rules.v1 IS 'v1: dom (租户ID - p策略) / role (角色ID - g策略)';
COMMENT ON COLUMN casbin_rules.v2 IS 'v2: obj (资源标识 - p策略) / dom (租户ID - g策略)';
COMMENT ON COLUMN casbin_rules.v3 IS 'v3: act (操作 - p策略)';
COMMENT ON COLUMN casbin_rules.v4 IS 'v4: 保留字段';
COMMENT ON COLUMN casbin_rules.v5 IS 'v5: 保留字段';

-- ========================================
-- 预置数据初始化
-- ========================================

-- 1. 创建默认系统租户 (ID示例: 100000000000000001)
INSERT INTO tenants (tenant_id, tenant_name, description, created_at, updated_at) VALUES 
('100000000000000001', '系统默认租户', '系统管理租户', EXTRACT(EPOCH FROM NOW())::BIGINT, EXTRACT(EPOCH FROM NOW())::BIGINT);

-- 2. 创建超级管理员用户 (密码: admin123, ID示例: 200000000000000001)
INSERT INTO users (user_id, user_name, password, name, status, created_at, updated_at) VALUES 
('200000000000000001', 'admin', '$2a$10$oZjrI9lUNaF3Vn1jTYKqPupglNiQb8SIh8ApThsUcVmY.JlgaBY3G', '超级管理员', 1, EXTRACT(EPOCH FROM NOW())::BIGINT, EXTRACT(EPOCH FROM NOW())::BIGINT);

-- 3. 关联用户到租户
INSERT INTO user_tenants (id, user_id, tenant_id, is_admin, created_at) VALUES 
('300000000000000001', '200000000000000001', '100000000000000001', TRUE, EXTRACT(EPOCH FROM NOW())::BIGINT);

-- 4. 创建超级管理员角色
INSERT INTO roles (role_id, tenant_id, role_name, description, created_at, updated_at) VALUES 
('400000000000000001', '100000000000000001', '超级管理员', '拥有所有权限', EXTRACT(EPOCH FROM NOW())::BIGINT, EXTRACT(EPOCH FROM NOW())::BIGINT);

-- 5. 创建管理员组
INSERT INTO user_groups (group_id, tenant_id, group_name, description, created_at, updated_at) VALUES 
('500000000000000001', '100000000000000001', '管理组', '系统管理人员组', EXTRACT(EPOCH FROM NOW())::BIGINT, EXTRACT(EPOCH FROM NOW())::BIGINT);

-- 6. 关联组和角色
INSERT INTO group_roles (id, group_id, role_id, created_at) VALUES 
('600000000000000001', '500000000000000001', '400000000000000001', EXTRACT(EPOCH FROM NOW())::BIGINT);

-- 7. 关联用户到组
INSERT INTO group_users (id, group_id, user_id, created_at) VALUES 
('700000000000000001', '500000000000000001', '200000000000000001', EXTRACT(EPOCH FROM NOW())::BIGINT);

-- 8. 插入一些基础权限 (示例)
INSERT INTO permissions (permission_id, code, name, type, path, created_at, updated_at) VALUES
('800000000000000001', 'sys:mgt', '系统管理', 'MENU', '/system', EXTRACT(EPOCH FROM NOW())::BIGINT, EXTRACT(EPOCH FROM NOW())::BIGINT),
('800000000000000002', 'sys:user:view', '查看用户', 'BUTTON', '', EXTRACT(EPOCH FROM NOW())::BIGINT, EXTRACT(EPOCH FROM NOW())::BIGINT),
('800000000000000003', 'sys:user:edit', '编辑用户', 'BUTTON', '', EXTRACT(EPOCH FROM NOW())::BIGINT, EXTRACT(EPOCH FROM NOW())::BIGINT);

-- 9. 赋予超级管理员角色所有权限 (这里只演示关联，实际业务中超级管理员可能跳过权限检查)
INSERT INTO role_permissions (id, role_id, permission_id, created_at) VALUES 
('900000000000000001', '400000000000000001', '800000000000000001', EXTRACT(EPOCH FROM NOW())::BIGINT),
('900000000000000002', '400000000000000001', '800000000000000002', EXTRACT(EPOCH FROM NOW())::BIGINT),
('900000000000000003', '400000000000000001', '800000000000000003', EXTRACT(EPOCH FROM NOW())::BIGINT);
