-- 000002 数据 - 基础配置初始化（幂等）
-- 演示 data migration 规范：INSERT ... ON CONFLICT DO UPDATE，ID 硬编码，时间戳毫秒。
-- 后续新增配置项：新建 000NNN_data_xxx.up.sql 只加新行，不改此文件。

INSERT INTO system_config (config_id, config_key, config_value, remark, created_at, updated_at, deleted_at) VALUES
  ('153547313510393201', 'site_name', 'Admin', '站点名称',
   EXTRACT(EPOCH FROM NOW())::BIGINT * 1000, EXTRACT(EPOCH FROM NOW())::BIGINT * 1000, 0),
  ('153547313510393202', 'site_status', 'active', '站点状态',
   EXTRACT(EPOCH FROM NOW())::BIGINT * 1000, EXTRACT(EPOCH FROM NOW())::BIGINT * 1000, 0)
ON CONFLICT (config_key, deleted_at) DO UPDATE SET
  config_value = EXCLUDED.config_value,
  remark = EXCLUDED.remark,
  updated_at = EXCLUDED.updated_at;
