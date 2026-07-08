-- 000002 数据 down - 显式删除本次范围
DELETE FROM system_config WHERE config_key IN ('site_name', 'site_status');
