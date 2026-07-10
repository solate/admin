-- 000002 数据 - 基础配置初始化（幂等）
-- 演示 data migration 规范：INSERT ... ON CONFLICT DO UPDATE，时间戳毫秒。
-- config_id 是原生 uuid 列 + DEFAULT uuidv7()，且无任何外键引用它，故不硬编码 ID，
-- 由库端 uuidv7() 自动生成（对 migration-data-convention 规则 4「ID 硬编码」的合理豁免：
-- 该规则前提是跨表外键需硬编码，此表无外键）。
-- 后续新增配置项：新建 000NNN_data_xxx.up.sql 只加新行，不改此文件。

INSERT INTO system_config (config_key, config_value, remark, created_at, updated_at, deleted_at) VALUES
  ('site_name', 'Admin', '站点名称',
   EXTRACT(EPOCH FROM NOW())::BIGINT * 1000, EXTRACT(EPOCH FROM NOW())::BIGINT * 1000, 0),
  ('site_status', 'active', '站点状态',
   EXTRACT(EPOCH FROM NOW())::BIGINT * 1000, EXTRACT(EPOCH FROM NOW())::BIGINT * 1000, 0)
ON CONFLICT (config_key, deleted_at) DO UPDATE SET
  config_value = EXCLUDED.config_value,
  remark = EXCLUDED.remark,
  updated_at = EXCLUDED.updated_at;
