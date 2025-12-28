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
    tenant_id VARCHAR(20) PRIMARY KEY, -- 18位字符串ID（默认租户：000000000000000000）
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
COMMENT ON COLUMN tenants.tenant_id IS '租户ID(18位字符串)';
COMMENT ON COLUMN tenants.tenant_code IS '租户编码(全局唯一业务标识)';
COMMENT ON COLUMN tenants.name IS '租户名称';
COMMENT ON COLUMN tenants.description IS '租户描述';
COMMENT ON COLUMN tenants.status IS '状态(1:启用, 2:禁用)';
COMMENT ON COLUMN tenants.created_at IS '创建时间戳(毫秒)';
COMMENT ON COLUMN tenants.updated_at IS '更新时间戳(毫秒)';
COMMENT ON COLUMN tenants.deleted_at IS '删除时间戳(毫秒,软删除)';


-- ========================================
-- 2. 用户表 (Users)
-- 所有用户都绑定租户，权限由 Casbin 的角色策略控制
-- ========================================
CREATE TABLE users (
    user_id VARCHAR(20) PRIMARY KEY, -- 18位字符串ID
    tenant_id VARCHAR(20) NOT NULL, -- 租户ID（所有用户都有值，包括超管）
    user_name VARCHAR(100) NOT NULL, -- 登录账号（租户内唯一）
    password VARCHAR(100) NOT NULL, -- 密码 (Bcrypt加密)
    name VARCHAR(100) NOT NULL DEFAULT '', -- 真实姓名/昵称
    avatar VARCHAR(255), -- 头像URL
    phone VARCHAR(20), -- 手机号
    email VARCHAR(100), -- 邮箱
    department_id VARCHAR(20), -- 所属部门ID
    position_id VARCHAR(20), -- 主岗位ID
    status SMALLINT NOT NULL DEFAULT 1, -- 状态 (1:正常, 2:冻结)
    remark TEXT,
    last_login_time BIGINT, -- 最后登录时间
    created_at BIGINT NOT NULL DEFAULT 0,
    updated_at BIGINT NOT NULL DEFAULT 0,
    deleted_at BIGINT DEFAULT 0
);

-- 租户内用户名唯一约束
-- ('tenant-001', 'admin') 租户内唯一
CREATE UNIQUE INDEX uk_tenant_username ON users(tenant_id, user_name) WHERE deleted_at = 0;

COMMENT ON TABLE users IS '用户表(所有用户都绑定租户，权限由Casbin角色策略控制)';
COMMENT ON COLUMN users.user_id IS '用户ID';
COMMENT ON COLUMN users.tenant_id IS '租户ID(所有用户都有值)';
COMMENT ON COLUMN users.user_name IS '用户名(登录账号，租户内唯一)';
COMMENT ON COLUMN users.password IS '加密密码';
COMMENT ON COLUMN users.name IS '姓名/昵称';
COMMENT ON COLUMN users.avatar IS '头像URL';
COMMENT ON COLUMN users.phone IS '手机号';
COMMENT ON COLUMN users.email IS '电子邮箱';
COMMENT ON COLUMN users.department_id IS '所属部门ID';
COMMENT ON COLUMN users.position_id IS '主岗位ID';
COMMENT ON COLUMN users.status IS '状态(1:启用, 2:禁用)';
COMMENT ON COLUMN users.remark IS '备注信息';
COMMENT ON COLUMN users.last_login_time IS '最后登录时间戳(毫秒)';
COMMENT ON COLUMN users.created_at IS '创建时间戳(毫秒)';
COMMENT ON COLUMN users.updated_at IS '更新时间戳(毫秒)';
COMMENT ON COLUMN users.deleted_at IS '删除时间戳(毫秒,软删除)';


-- ========================================
-- 3. 角色表 (Roles)
-- 租户自定义角色，继承关系由 Casbin g2 策略管理
-- ========================================
CREATE TABLE roles (
    role_id VARCHAR(20) PRIMARY KEY,
    tenant_id VARCHAR(20) NOT NULL,  -- [多租户核心] 角色属于特定租户
    code VARCHAR(50) NOT NULL,       -- 角色编码 (如: sales, manager)
    name VARCHAR(100) NOT NULL,      -- 角色名称 (如: 销售角色)
    description TEXT,
    status SMALLINT NOT NULL DEFAULT 1,
    created_at BIGINT NOT NULL DEFAULT 0,
    updated_at BIGINT NOT NULL DEFAULT 0,
    deleted_at BIGINT DEFAULT 0
);

-- 租户内角色编码唯一约束
CREATE UNIQUE INDEX idx_roles_tenant_code ON roles(tenant_id, code) WHERE deleted_at = 0;

COMMENT ON TABLE roles IS '角色表(租户隔离，继承关系由Casbin g2策略管理)';
COMMENT ON COLUMN roles.role_id IS '角色ID(18位字符串)';
COMMENT ON COLUMN roles.tenant_id IS '所属租户ID';
COMMENT ON COLUMN roles.code IS '角色编码(租户内唯一，用于Casbin策略)';
COMMENT ON COLUMN roles.name IS '角色名称';
COMMENT ON COLUMN roles.description IS '角色描述';
COMMENT ON COLUMN roles.status IS '状态(1:启用, 2:禁用)';
COMMENT ON COLUMN roles.created_at IS '创建时间戳(毫秒)';
COMMENT ON COLUMN roles.updated_at IS '更新时间戳(毫秒)';
COMMENT ON COLUMN roles.deleted_at IS '删除时间戳(毫秒,软删除)';


-- ========================================
-- 4. 菜单表 (Menus)
-- 全局菜单元数据定义表（不区分租户）
-- 菜单的租户边界通过 tenant_menus 表控制
-- ========================================

CREATE TABLE menus (
    menu_id VARCHAR(20) PRIMARY KEY,
    parent_id VARCHAR(20),                   -- 父菜单ID（用于构建树形结构）

    -- 菜单基础信息
    name VARCHAR(100) NOT NULL,               -- 菜单名称
    path VARCHAR(255),                        -- 前端路由路径
    component VARCHAR(255),                   -- 前端组件路径
    redirect VARCHAR(255),                    -- 重定向路径
    icon VARCHAR(100),                        -- 图标
    sort INT DEFAULT 0,                       -- 排序权重

    -- 状态控制（1=启用且显示, 2=禁用且隐藏）
    status SMALLINT DEFAULT 1,                -- 状态(1:启用, 2:禁用)

    description TEXT,
    created_at BIGINT NOT NULL DEFAULT 0,
    updated_at BIGINT NOT NULL DEFAULT 0,
    deleted_at BIGINT DEFAULT 0,
    KEY idx_parent(parent_id, deleted_at)
);

COMMENT ON TABLE menus IS '菜单元数据表(全局定义，租户边界由tenant_menus表控制)';
COMMENT ON COLUMN menus.menu_id IS '菜单ID(18位字符串)';
COMMENT ON COLUMN menus.parent_id IS '父菜单ID(用于构建树形结构)';
COMMENT ON COLUMN menus.name IS '菜单名称';
COMMENT ON COLUMN menus.path IS '前端路由路径';
COMMENT ON COLUMN menus.component IS '前端组件路径';
COMMENT ON COLUMN menus.redirect IS '重定向路径';
COMMENT ON COLUMN menus.icon IS '图标';
COMMENT ON COLUMN menus.sort IS '显示排序';
COMMENT ON COLUMN menus.status IS '状态(1:启用, 2:禁用)';
COMMENT ON COLUMN menus.description IS '描述信息';
COMMENT ON COLUMN menus.created_at IS '创建时间戳(毫秒)';
COMMENT ON COLUMN menus.updated_at IS '更新时间戳(毫秒)';
COMMENT ON COLUMN menus.deleted_at IS '删除时间戳(毫秒,软删除)';


-- ========================================
-- 5. 权限点定义表 (Permissions)
-- 全局权限点定义，供前端权限选择器使用
-- resource 格式: menu:xxx, btn:xxx, /api/v1/xxx
-- ========================================
CREATE TABLE permissions (
    permission_id VARCHAR(20) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,              -- 权限名称
    type VARCHAR(20) NOT NULL,               -- 类型: MENU, BUTTON, API
    resource VARCHAR(255) NOT NULL,          -- 资源标识 (menu:xxx, btn:xxx, /api/v1/xxx)
    action VARCHAR(50),                      -- 请求方法 (GET, POST, PUT, DELETE)，仅 API 类型有效
    description TEXT,
    created_at BIGINT NOT NULL DEFAULT 0,
    updated_at BIGINT NOT NULL DEFAULT 0,
    deleted_at BIGINT DEFAULT 0
);

-- 索引优化
CREATE INDEX idx_permissions_type ON permissions(type, deleted_at);
CREATE INDEX idx_permissions_resource ON permissions(resource, deleted_at);

COMMENT ON TABLE permissions IS '权限点定义表(全局，供前端权限选择器使用)';
COMMENT ON COLUMN permissions.permission_id IS '权限ID(18位字符串)';
COMMENT ON COLUMN permissions.name IS '权限名称';
COMMENT ON COLUMN permissions.type IS '类型(MENU:菜单, BUTTON:按钮, API:接口)';
COMMENT ON COLUMN permissions.resource IS '资源标识(menu:xxx, btn:xxx, /api/v1/xxx)';
COMMENT ON COLUMN permissions.action IS '请求方法(GET/POST/PUT/DELETE,仅API类型)';
COMMENT ON COLUMN permissions.description IS '描述信息';
COMMENT ON COLUMN permissions.created_at IS '创建时间戳(毫秒)';
COMMENT ON COLUMN permissions.updated_at IS '更新时间戳(毫秒)';
COMMENT ON COLUMN permissions.deleted_at IS '删除时间戳(毫秒,软删除)';


-- ========================================
-- 6. 租户菜单边界表 (Tenant Menus)
-- 超管给租户分配的可用菜单范围
-- 租户角色的菜单权限必须在此范围内
-- ========================================
CREATE TABLE tenant_menus (
    id VARCHAR(20) PRIMARY KEY,
    tenant_id VARCHAR(20) NOT NULL,          -- 租户ID
    menu_id VARCHAR(20) NOT NULL,            -- 菜单ID
    created_at BIGINT NOT NULL DEFAULT 0,
    deleted_at BIGINT DEFAULT 0,
    UNIQUE KEY uk_tenant_menu(tenant_id, menu_id, deleted_at)
);

-- 索引优化
CREATE INDEX idx_tenant_menus_tenant ON tenant_menus(tenant_id, deleted_at);

COMMENT ON TABLE tenant_menus IS '租户菜单边界表(超管分配租户可用菜单)';
COMMENT ON COLUMN tenant_menus.id IS '主键ID(18位字符串)';
COMMENT ON COLUMN tenant_menus.tenant_id IS '租户ID';
COMMENT ON COLUMN tenant_menus.menu_id IS '菜单ID';
COMMENT ON COLUMN tenant_menus.created_at IS '创建时间戳(毫秒)';
COMMENT ON COLUMN tenant_menus.deleted_at IS '删除时间戳(毫秒,软删除)';


-- ========================================
-- 7. Casbin 策略表 (Casbin Rules)
-- 由 Casbin gorm-adapter 自动创建，用于 RBAC 模型持久化
-- 支持带租户的 RBAC 模型 (RBAC with Domains)
-- ptype='p': 权限策略 p(sub, dom, obj, act)
-- ptype='g': 用户角色关联 g(user, role, domain)
-- ptype='g2': 角色继承 g2(child_role, parent_role) - 不需要 domain
-- ========================================
-- 注：该表由 Casbin 自动创建，无需手动建表


-- ========================================
-- 8. 组织结构系统 (Departments & Positions)
-- 部门按职能划分，岗位按职责定义，支持一人多岗
-- 详细设计见: docs/plan/org-system-design.md
-- ========================================

-- 8.1 部门表 (Departments)
CREATE TABLE departments (
    department_id VARCHAR(20) PRIMARY KEY,
    tenant_id VARCHAR(20) NOT NULL,
    parent_id VARCHAR(20),                    -- 父部门ID（用于构建树形结构）
    department_name VARCHAR(100) NOT NULL,     -- 部门名称
    department_code VARCHAR(50),               -- 部门编码
    description TEXT,                          -- 部门描述
    sort INT DEFAULT 0,                        -- 排序权重
    status SMALLINT DEFAULT 1,                 -- 状态 (1:启用, 2:禁用)
    created_at BIGINT NOT NULL DEFAULT 0,
    updated_at BIGINT NOT NULL DEFAULT 0,
    deleted_at BIGINT DEFAULT 0
);

-- 索引优化：支持租户内部门和父子关系查询
CREATE INDEX idx_departments_tenant_parent ON departments(tenant_id, parent_id, deleted_at);
CREATE INDEX idx_departments_tenant_code ON departments(tenant_id, department_code, deleted_at);

COMMENT ON TABLE departments IS '部门表(按职能划分，支持树形结构)';
COMMENT ON COLUMN departments.department_id IS '部门ID(18位字符串)';
COMMENT ON COLUMN departments.tenant_id IS '租户ID';
COMMENT ON COLUMN departments.parent_id IS '父部门ID(用于构建树形结构)';
COMMENT ON COLUMN departments.department_name IS '部门名称';
COMMENT ON COLUMN departments.department_code IS '部门编码';
COMMENT ON COLUMN departments.description IS '部门描述';
COMMENT ON COLUMN departments.sort IS '排序权重';
COMMENT ON COLUMN departments.status IS '状态(1:启用, 2:禁用)';
COMMENT ON COLUMN departments.created_at IS '创建时间戳(毫秒)';
COMMENT ON COLUMN departments.updated_at IS '更新时间戳(毫秒)';
COMMENT ON COLUMN departments.deleted_at IS '删除时间戳(毫秒,软删除)';


-- 8.2 岗位表 (Positions)
-- 岗位编码（如 DEPT_LEADER, EMPLOYEE）与 Casbin 角色对应
CREATE TABLE positions (
    position_id VARCHAR(20) PRIMARY KEY,
    tenant_id VARCHAR(20) NOT NULL,
    position_code VARCHAR(50) NOT NULL,        -- 岗位编码 (DEPT_LEADER, EMPLOYEE, HR 等)
    position_name VARCHAR(100) NOT NULL,       -- 岗位名称
    level INT,                                 -- 职级
    description TEXT,                          -- 岗位描述
    sort INT DEFAULT 0,                        -- 排序权重
    status SMALLINT DEFAULT 1,                 -- 状态 (1:启用, 2:禁用)
    created_at BIGINT NOT NULL DEFAULT 0,
    updated_at BIGINT NOT NULL DEFAULT 0,
    deleted_at BIGINT DEFAULT 0
);

-- 租户内岗位编码唯一约束（标准岗位编码与 Casbin 角色对应）
CREATE UNIQUE INDEX uk_positions_tenant_code ON positions(tenant_id, position_code) WHERE deleted_at = 0;

COMMENT ON TABLE positions IS '岗位表(按职责定义，岗位编码与Casbin角色对应)';
COMMENT ON COLUMN positions.position_id IS '岗位ID(18位字符串)';
COMMENT ON COLUMN positions.tenant_id IS '租户ID';
COMMENT ON COLUMN positions.position_code IS '岗位编码(租户内唯一，如DEPT_LEADER, EMPLOYEE)';
COMMENT ON COLUMN positions.position_name IS '岗位名称';
COMMENT ON COLUMN positions.level IS '职级';
COMMENT ON COLUMN positions.description IS '岗位描述';
COMMENT ON COLUMN positions.sort IS '排序权重';
COMMENT ON COLUMN positions.status IS '状态(1:启用, 2:禁用)';
COMMENT ON COLUMN positions.created_at IS '创建时间戳(毫秒)';
COMMENT ON COLUMN positions.updated_at IS '更新时间戳(毫秒)';
COMMENT ON COLUMN positions.deleted_at IS '删除时间戳(毫秒,软删除)';


-- 8.3 用户多岗关联表 (User Positions) - 可选
-- 支持用户兼任多个岗位
CREATE TABLE user_positions (
    user_id VARCHAR(20) NOT NULL,
    position_id VARCHAR(20) NOT NULL,
    is_primary BOOLEAN DEFAULT TRUE,          -- 是否为主岗位
    PRIMARY KEY (user_id, position_id)
);

COMMENT ON TABLE user_positions IS '用户多岗关联表(支持用户兼任多个岗位)';
COMMENT ON COLUMN user_positions.user_id IS '用户ID';
COMMENT ON COLUMN user_positions.position_id IS '岗位ID';
COMMENT ON COLUMN user_positions.is_primary IS '是否为主岗位';


-- ========================================
-- 9. 登录日志表 (Login Logs)
-- 记录用户登录行为，用于安全审计
-- ========================================
CREATE TABLE login_logs (
    log_id VARCHAR(20) PRIMARY KEY,
    tenant_id VARCHAR(20) NOT NULL,
    user_id VARCHAR(20),
    user_name VARCHAR(100),                  -- 登录账号
    user_display_name VARCHAR(100),          -- 昵称
    login_type VARCHAR(20),                  -- PASSWORD, SSO, OAUTH
    login_ip VARCHAR(50),
    login_location VARCHAR(100),             -- IP解析的地理位置
    user_agent VARCHAR(255),
    status SMALLINT,                         -- 1:成功 0:失败
    fail_reason VARCHAR(255),                -- 失败原因
    created_at BIGINT NOT NULL DEFAULT 0,
    INDEX idx_tenant_user (tenant_id, user_id),
    INDEX idx_tenant_time (tenant_id, created_at)
);

COMMENT ON TABLE login_logs IS '登录日志表(记录用户登录行为，用于安全审计)';
COMMENT ON COLUMN login_logs.log_id IS '日志ID(18位字符串)';
COMMENT ON COLUMN login_logs.tenant_id IS '租户ID';
COMMENT ON COLUMN login_logs.user_id IS '用户ID';
COMMENT ON COLUMN login_logs.user_name IS '登录账号';
COMMENT ON COLUMN login_logs.user_display_name IS '昵称';
COMMENT ON COLUMN login_logs.login_type IS '登录类型(PASSWORD:密码, SSO:单点登录, OAUTH:第三方登录)';
COMMENT ON COLUMN login_logs.login_ip IS '登录IP地址';
COMMENT ON COLUMN login_logs.login_location IS '登录位置(IP解析的地理位置)';
COMMENT ON COLUMN login_logs.user_agent IS '用户代理(浏览器/客户端信息)';
COMMENT ON COLUMN login_logs.status IS '状态(1:成功, 0:失败)';
COMMENT ON COLUMN login_logs.fail_reason IS '失败原因';
COMMENT ON COLUMN login_logs.created_at IS '创建时间戳(毫秒)';


-- ========================================
-- 10. 操作日志表 (Operation Logs)
-- 记录用户在系统中的操作行为，用于审计和追踪
-- ========================================
CREATE TABLE operation_logs (
    log_id VARCHAR(20) PRIMARY KEY,
    tenant_id VARCHAR(20) NOT NULL,
    user_id VARCHAR(20),
    user_name VARCHAR(100),                  -- 登录账号
    user_display_name VARCHAR(100),          -- 昵称
    module VARCHAR(50),                      -- 模块名
    operation_type VARCHAR(20),              -- CREATE, UPDATE, DELETE, QUERY
    resource_type VARCHAR(50),               -- 资源类型
    resource_id VARCHAR(255),                -- 资源ID
    resource_name VARCHAR(255),              -- 资源名称
    request_method VARCHAR(10),              -- GET, POST, PUT, DELETE
    request_path VARCHAR(500),               -- 请求路径
    request_params TEXT,                     -- 请求参数（脱敏）
    old_value TEXT,                          -- 旧值（JSON）
    new_value TEXT,                          -- 新值（JSON）
    status SMALLINT,                         -- 1:成功 2:失败
    error_message TEXT,                      -- 错误信息
    ip_address VARCHAR(50),
    user_agent TEXT,
    created_at BIGINT NOT NULL DEFAULT 0,
    INDEX idx_tenant_user (tenant_id, user_id),
    INDEX idx_tenant_time (tenant_id, created_at),
    INDEX idx_module (tenant_id, module, created_at),
    INDEX idx_resource (resource_type, resource_id)
);

COMMENT ON TABLE operation_logs IS '操作日志表(记录用户操作行为，用于审计和追踪)';
COMMENT ON COLUMN operation_logs.log_id IS '日志ID(18位字符串)';
COMMENT ON COLUMN operation_logs.tenant_id IS '租户ID';
COMMENT ON COLUMN operation_logs.user_id IS '操作人用户ID';
COMMENT ON COLUMN operation_logs.user_name IS '登录账号';
COMMENT ON COLUMN operation_logs.user_display_name IS '昵称';
COMMENT ON COLUMN operation_logs.module IS '模块名';
COMMENT ON COLUMN operation_logs.operation_type IS '操作类型(CREATE:创建, UPDATE:更新, DELETE:删除, QUERY:查询)';
COMMENT ON COLUMN operation_logs.resource_type IS '资源类型';
COMMENT ON COLUMN operation_logs.resource_id IS '资源ID';
COMMENT ON COLUMN operation_logs.resource_name IS '资源名称';
COMMENT ON COLUMN operation_logs.request_method IS '请求方法(GET, POST, PUT, DELETE)';
COMMENT ON COLUMN operation_logs.request_path IS '请求路径';
COMMENT ON COLUMN operation_logs.request_params IS '请求参数(已脱敏)';
COMMENT ON COLUMN operation_logs.old_value IS '操作前数据(JSON格式)';
COMMENT ON COLUMN operation_logs.new_value IS '操作后数据(JSON格式)';
COMMENT ON COLUMN operation_logs.status IS '操作状态(1:成功, 2:失败)';
COMMENT ON COLUMN operation_logs.error_message IS '错误信息';
COMMENT ON COLUMN operation_logs.ip_address IS '操作来源IP';
COMMENT ON COLUMN operation_logs.user_agent IS '用户代理信息';
COMMENT ON COLUMN operation_logs.created_at IS '操作时间戳(毫秒)';


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
