-- 数字工厂测试数据清理。
DELETE FROM fact_upgrade_record WHERE plugin_id IN ('990000000000001001', '990000000000001002');
DELETE FROM fact_user_ownership WHERE plugin_id IN ('990000000000001001', '990000000000001002');
DELETE FROM fact_mint_record WHERE plugin_id IN ('990000000000001001', '990000000000001002');
DELETE FROM fact_release_status_history WHERE plugin_id IN ('990000000000001001', '990000000000001002');
DELETE FROM fact_release_price_history WHERE plugin_id IN ('990000000000001001', '990000000000001002');
DELETE FROM fact_release WHERE plugin_id IN ('990000000000001001', '990000000000001002');
DELETE FROM dev_plugin WHERE id IN (990000000000001001, 990000000000001002);

-- 数字工厂测试插件。
INSERT INTO dev_plugin (id, name, description, version, author, created_by, created_on)
VALUES
  (990000000000001001, 'factory-test-publish-plugin', '数字工厂发布测试插件', '1.0.0', 'factory-tester', 10001, NOW()),
  (990000000000001002, 'factory-test-upgrade-plugin', '数字工厂升级测试插件', '1.0.0', 'factory-tester', 10001, NOW());
