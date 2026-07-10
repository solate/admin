-- 000001 结构 - 初始化 schema
-- 建一张基础设施级配置表，用于验证 DAL 管线（时间戳毫秒 bigint + soft_delete）。
-- 业务域表（tenants/users/roles...）由后续 step-04 定义，不在此处。

CREATE TABLE IF NOT EXISTS system_config (
    config_id    uuid         PRIMARY KEY DEFAULT uuidv7(),  -- UUIDv7 主键，PG18 库端自动生成(原生 uuid 16字节)
    config_key   VARCHAR(128) NOT NULL,             -- 配置键
    config_value TEXT         NOT NULL DEFAULT '',   -- 配置值
    remark       VARCHAR(255) NOT NULL DEFAULT '',   -- 备注
    created_at   BIGINT       NOT NULL DEFAULT 0,    -- 创建时间(毫秒)
    updated_at   BIGINT       NOT NULL DEFAULT 0,    -- 更新时间(毫秒)
    deleted_at   BIGINT       NOT NULL DEFAULT 0,    -- 软删除时间(毫秒,0=未删)
    UNIQUE (config_key, deleted_at)                  -- 键唯一(软删维度)
);

CREATE INDEX IF NOT EXISTS idx_system_config_key ON system_config (config_key);
