-- 创建用户表
CREATE TABLE IF NOT EXISTS users (
    user_id VARCHAR(255) PRIMARY KEY,
    user_name VARCHAR(255),
    password VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'uploader',
    
    name VARCHAR(255) NOT NULL DEFAULT '用户',
    avatar VARCHAR(255),
    phone VARCHAR(20),
    email VARCHAR(255),
    status INTEGER NOT NULL DEFAULT 1,
    last_login_time BIGINT,
    remark VARCHAR(500),
    created_at BIGINT NOT NULL DEFAULT 0,
    updated_at BIGINT NOT NULL DEFAULT 0,
    deleted_at BIGINT
);

-- 创建索引
CREATE INDEX idx_users_user_id ON users(user_id);
CREATE INDEX idx_users_user_name ON users(user_name);
CREATE INDEX idx_users_role ON users(role);

-- 插入预置用户 (密码都是: admin123)
-- bcrypt hash of "admin123" (cost=10)
INSERT INTO users (user_id, user_name, password, role, name, created_at, updated_at) VALUES 
    ('153547313510393194', 'admin', '$2a$10$oZjrI9lUNaF3Vn1jTYKqPupglNiQb8SIh8ApThsUcVmY.JlgaBY3G', 'admin', '管理员', EXTRACT(EPOCH FROM NOW())::BIGINT, EXTRACT(EPOCH FROM NOW())::BIGINT),
    ('153547313510458730', 'auditor', '$2a$10$oZjrI9lUNaF3Vn1jTYKqPupglNiQb8SIh8ApThsUcVmY.JlgaBY3G', 'auditor', '审核员', EXTRACT(EPOCH FROM NOW())::BIGINT, EXTRACT(EPOCH FROM NOW())::BIGINT),
    ('153547313510524266', 'uploader', '$2a$10$oZjrI9lUNaF3Vn1jTYKqPupglNiQb8SIh8ApThsUcVmY.JlgaBY3G', 'uploader', '上传员', EXTRACT(EPOCH FROM NOW())::BIGINT, EXTRACT(EPOCH FROM NOW())::BIGINT);

-- 添加备注
COMMENT ON TABLE users IS '用户表';
COMMENT ON COLUMN users.created_at IS '创建时间';
COMMENT ON COLUMN users.updated_at IS '修改时间';
COMMENT ON COLUMN users.deleted_at IS '删除时间（软删除）';
COMMENT ON COLUMN users.user_id IS '会员ID（唯一）';
COMMENT ON COLUMN users.user_name IS '用户名';
COMMENT ON COLUMN users.password IS '密码（bcrypt加密）';
COMMENT ON COLUMN users.role IS '角色：admin, auditor, uploader';
COMMENT ON COLUMN users.name IS '昵称';
COMMENT ON COLUMN users.avatar IS '头像';
COMMENT ON COLUMN users.phone IS '手机号';
COMMENT ON COLUMN users.email IS '邮箱';
COMMENT ON COLUMN users.status IS '状态: 1-启用 2-禁用';
COMMENT ON COLUMN users.last_login_time IS '最后登录时间';
COMMENT ON COLUMN users.remark IS '备注';

