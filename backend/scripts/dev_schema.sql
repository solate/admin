-- ========================================
-- 开发环境数据库初始化脚本（SaaS多租户版）
-- 说明：
-- 1. 采用 "字段隔离" (Discriminator Column) 模式实现多租户架构
-- 2. tenant_id 为租户唯一标识，除 tenants 表外，核心业务表均需包含该字段
-- 3. 业务层需强制检查 tenant_id，防止越权访问
-- 4. 物理外键约束被移除，由应用层保证数据一致性（便于分库分表扩展）
-- ========================================


-- ========================================
-- 1. 租户表 (Tenants)
-- 核心表：存储租户的基础信息，全局唯一
-- ========================================
CREATE TABLE tenants (
    tenant_id VARCHAR(36) PRIMARY KEY, -- UUID作为主键，便于数据迁移和防碰撞
    tenant_code VARCHAR(50) NOT NULL, -- 租户编码（全局唯一，可用于二级域名或URL路径，如：tenant_shanghai）
    name VARCHAR(200) NOT NULL, -- 租户名称（企业/组织名称）
    description TEXT,
    status SMALLINT NOT NULL DEFAULT 1, -- 状态：1-正常, 2-冻结/停用 (影响该租户下所有用户访问)
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    deleted_at BIGINT DEFAULT 0 -- 软删除标识，0表示未删除
);
CREATE UNIQUE INDEX idx_tenants_tenant_code ON tenants(tenant_code);

COMMENT ON TABLE tenants IS '租户表(SaaS核心表)';
COMMENT ON COLUMN tenants.tenant_id IS '租户ID(UUID)';
COMMENT ON COLUMN tenants.tenant_code IS '租户编码(全局唯一业务标识)';
COMMENT ON COLUMN tenants.name IS '租户名称';
COMMENT ON COLUMN tenants.description IS '租户描述';
COMMENT ON COLUMN tenants.status IS '状态(1:启用, 2:禁用)';
COMMENT ON COLUMN tenants.created_at IS '创建时间戳(毫秒)';
COMMENT ON COLUMN tenants.updated_at IS '更新时间戳(毫秒)';
COMMENT ON COLUMN tenants.deleted_at IS '删除时间戳(软删除)';


-- ========================================
-- 2. 用户表 (Users)
-- 平台级用户表，支持同一手机号/邮箱在不同租户下存在（视业务需求而定）
-- 当前设计：用户名在租户内唯一
-- ========================================
CREATE TABLE users (
    user_id VARCHAR(255) PRIMARY KEY, -- 建议统一使用UUID (VARCHAR(36))
    user_name VARCHAR(255) NOT NULL, -- 登录账号
    password VARCHAR(255) NOT NULL, -- 密码 (Bcrypt加密)
    name VARCHAR(255) NOT NULL DEFAULT '', -- 真实姓名/昵称
    avatar VARCHAR(255), -- 头像URL
    phone VARCHAR(20), -- 手机号
    email VARCHAR(255), -- 邮箱
    tenant_id VARCHAR(36) NOT NULL, -- [多租户核心] 所属租户ID
    role_type INTEGER NOT NULL DEFAULT 1, -- 角色类型标识 (1:普通用户, 2:租户管理员, 3:平台超级管理员) - 用于快速权限判断
    status INTEGER NOT NULL DEFAULT 1, -- 状态 (1:正常, 2:冻结)
    remark TEXT,
    last_login_time BIGINT, -- 最后登录时间
    created_at BIGINT NOT NULL DEFAULT 0,
    updated_at BIGINT NOT NULL DEFAULT 0,
    deleted_at BIGINT DEFAULT 0
);

-- 租户内用户名唯一约束 (实现租户间账号隔离)
CREATE UNIQUE INDEX idx_users_tenant_username ON users(tenant_id, user_name) WHERE deleted_at = 0;

COMMENT ON TABLE users IS '用户表(多租户隔离)';
COMMENT ON COLUMN users.user_id IS '用户ID';
COMMENT ON COLUMN users.user_name IS '用户名(登录账号)';
COMMENT ON COLUMN users.password IS '加密密码';
COMMENT ON COLUMN users.name IS '姓名/昵称';
COMMENT ON COLUMN users.avatar IS '头像URL';
COMMENT ON COLUMN users.phone IS '手机号';
COMMENT ON COLUMN users.email IS '电子邮箱';
COMMENT ON COLUMN users.tenant_id IS '所属租户ID';
COMMENT ON COLUMN users.role_type IS '管理员类型(1:普通用户, 2:租户管理员, 3:平台超级管理员)';
COMMENT ON COLUMN users.status IS '状态(1:启用, 2:禁用)';
COMMENT ON COLUMN users.remark IS '备注信息';
COMMENT ON COLUMN users.last_login_time IS '最后登录时间戳';
COMMENT ON COLUMN users.created_at IS '创建时间戳';
COMMENT ON COLUMN users.updated_at IS '更新时间戳';
COMMENT ON COLUMN users.deleted_at IS '删除时间戳(软删除)';


-- ========================================
-- 3. 角色表 (Roles)
-- 租户自定义角色，实现RBAC模型
-- ========================================
CREATE TABLE roles (
    role_id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL, -- [多租户核心] 角色属于特定租户
    role_code VARCHAR(50) NOT NULL, -- 角色编码 (如: hr_manager)
    name VARCHAR(100) NOT NULL, -- 角色名称 (如: 人事经理)
    description TEXT,
    status INTEGER NOT NULL DEFAULT 1,
    created_at BIGINT NOT NULL DEFAULT 0,
    updated_at BIGINT NOT NULL DEFAULT 0,
    deleted_at BIGINT DEFAULT 0
);

-- 租户内角色编码唯一约束
CREATE UNIQUE INDEX idx_roles_tenant_role_code ON roles(tenant_id, role_code) WHERE deleted_at = 0;

COMMENT ON TABLE roles IS '角色表(租户隔离)';
COMMENT ON COLUMN roles.role_id IS '角色ID';
COMMENT ON COLUMN roles.tenant_id IS '所属租户ID';
COMMENT ON COLUMN roles.role_code IS '角色编码(租户内唯一)';
COMMENT ON COLUMN roles.name IS '角色名称';
COMMENT ON COLUMN roles.description IS '角色描述';
COMMENT ON COLUMN roles.status IS '状态(1:启用, 2:禁用)';
COMMENT ON COLUMN roles.created_at IS '创建时间戳';
COMMENT ON COLUMN roles.updated_at IS '更新时间戳';
COMMENT ON COLUMN roles.deleted_at IS '删除时间戳(软删除)';


-- ========================================
-- 4. 权限表 (Permissions)
-- 功能/菜单/API资源的定义
-- 注意：当前设计为"租户级权限"，即每个租户可定义自己的权限列表
-- 若系统权限是统一的，建议将tenant_id设为可空或移除，作为公共元数据
-- ========================================
CREATE TABLE permissions (
    permission_id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL, -- [多租户核心] 属于特定租户的权限定义
    name VARCHAR(100) NOT NULL, -- 权限名称
    type VARCHAR(20) NOT NULL, -- 资源类型: MENU(菜单), BUTTON(按钮), API(接口), DATA(数据)
    resource VARCHAR(255), -- 资源路径/路由/API地址
    action VARCHAR(20), -- 操作动词 (GET, POST, PUT, DELETE)
    sort INTEGER DEFAULT 0, -- 排序权重
    description TEXT,
    created_at BIGINT NOT NULL DEFAULT 0,
    updated_at BIGINT NOT NULL DEFAULT 0,
    deleted_at BIGINT DEFAULT 0
);

COMMENT ON TABLE permissions IS '权限/菜单定义表(租户隔离)';
COMMENT ON COLUMN permissions.permission_id IS '权限ID';
COMMENT ON COLUMN permissions.tenant_id IS '所属租户ID';
COMMENT ON COLUMN permissions.code IS '权限标识(如 sys:user:list)';
COMMENT ON COLUMN permissions.name IS '权限名称';
COMMENT ON COLUMN permissions.type IS '类型(MENU:菜单, BUTTON:按钮, API:接口, DATA:数据)';
COMMENT ON COLUMN permissions.resource IS '资源路径(路由/API)';
COMMENT ON COLUMN permissions.action IS '请求方法(仅API类型有效)';
COMMENT ON COLUMN permissions.sort IS '显示排序';
COMMENT ON COLUMN permissions.description IS '描述信息';
COMMENT ON COLUMN permissions.created_at IS '创建时间戳';
COMMENT ON COLUMN permissions.updated_at IS '更新时间戳';

-- ========================================
-- 5. 用户-租户-角色关联表 (User Tenant Role)
-- 实现多租户下用户角色分配
-- ========================================
CREATE TABLE user_tenant_role (
    user_tenant_role_id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL, -- 用户ID
    tenant_id VARCHAR(36) NOT NULL, -- 租户ID
    role_id VARCHAR(255) NOT NULL -- 角色ID
);

-- 租户内用户-角色关联唯一约束
CREATE UNIQUE INDEX idx_user_tenant_role_tenant_user_role ON user_tenant_role(tenant_id, user_id, role_id) WHERE deleted_at = 0;

COMMENT ON TABLE user_tenant_role IS '用户-租户-角色关联表(核心)';
COMMENT ON COLUMN user_tenant_role.user_tenant_role_id IS '关联ID';
COMMENT ON COLUMN user_tenant_role.user_id IS '用户ID';
COMMENT ON COLUMN user_tenant_role.tenant_id IS '租户ID';
COMMENT ON COLUMN user_tenant_role.role_id IS '角色ID';





-- -- ========================================
-- -- (会自动建表，无需自己建)
-- -- 5. Casbin 策略表 (Casbin Rules)
-- -- 用于 Casbin RBAC 模型持久化
-- -- 支持带租户的 RBAC 模型 (RBAC with Domains)
-- -- ========================================
-- CREATE TABLE casbin_rule (
--     id SERIAL PRIMARY KEY,
--     ptype VARCHAR(255), -- 'p', 'g', 'g2'
--     v0 VARCHAR(255),    -- subject (user/role)
--     v1 VARCHAR(255),    -- domain/role/object
--     v2 VARCHAR(255),    -- object/role/action
--     v3 VARCHAR(255),    -- action (p类型) / 空 (g,g2类型)
--     v4 VARCHAR(255),    -- 通常为空，保留字段
--     v5 VARCHAR(255)     -- 通常为空，保留字段
-- );
-- -- 索引优化查询性能
-- CREATE UNIQUE INDEX idx_casbin_rule ON casbin_rule (ptype,v0,v1,v2,v3,v4,v5)
-- CREATE INDEX idx_casbin_ptype ON casbin_rules(ptype);
-- -- 针对 RBAC with Domain 模型的特定查询优化
-- CREATE INDEX idx_casbin_g_lookup ON casbin_rules(ptype, v0, v2) WHERE ptype = 'g';  -- g策略查询: g, user, domain -> 找角色
-- CREATE INDEX idx_casbin_p_match ON casbin_rules(ptype, v1, v2, v3) WHERE ptype = 'p';  -- p策略匹配: p, domain, resource, action -> 找策略


-- COMMENT ON TABLE casbin_rules IS 'Casbin权限策略表 (RBAC with Domains)';
-- COMMENT ON COLUMN casbin_rules.id IS '主键自增ID';
-- COMMENT ON COLUMN casbin_rules.ptype IS '策略类型(p:策略, g:角色关联, g2:角色继承)';
-- COMMENT ON COLUMN casbin_rules.v0 IS 'v0: Subject (用户ID/角色ID)';
-- COMMENT ON COLUMN casbin_rules.v1 IS 'v1: Domain (租户ID) / Role (角色ID)';
-- COMMENT ON COLUMN casbin_rules.v2 IS 'v2: Object (资源) / Domain (租户ID) / Role (角色ID)';
-- COMMENT ON COLUMN casbin_rules.v3 IS 'v3: Action (操作)';
-- COMMENT ON COLUMN casbin_rules.v4 IS '扩展字段';
-- COMMENT ON COLUMN casbin_rules.v5 IS '扩展字段';


