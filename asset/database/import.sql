INSERT INTO user_info (id ,created_on , name ,password ,salt,state ,username ,created_by)VALUES (1, NOW(), '管理员', 'd3c59d25033dbf980d29554025c23a75', '8d78869f470951332959580424d4bf4f', '0', 'admin', 1);
INSERT INTO user_info (id ,created_on ,name ,password ,salt,state ,username ,created_by)VALUES (2, NOW(), 'jack li', '123456', null, '0', 'jack', 1);
INSERT INTO user_info (id ,created_on ,name ,password ,salt,state ,username ,created_by)VALUES(3, NOW(), 'jack lily', '123456', null, '0', 'jackli', 1);
INSERT INTO user_info (id ,created_on ,name ,password ,salt,state ,username ,created_by)VALUES (4, '2018-12-16 11:17:03', 'jack hony', '123456', null, '0', 'jackho', 1);
go

INSERT INTO smarterp.sys_role (id, name, remark, created_by, created_on) VALUES (1, 'admin', null, 1, '2020-05-17 16:56:31');
INSERT INTO smarterp.sys_role (id, name, remark, created_by, created_on) VALUES (2, 'user', null, 1, '2020-05-17 16:56:52');

go
INSERT INTO `sys_menu`(`id`, `parent_id`, `name`, `url`, `perms`, `type`, `icon`, `order_num`, `created_by`, `created_on`) VALUES (1, 0, '系统管理', NULL, NULL, 0, 'icon_user_nav', 4, NULL, NULL);
INSERT INTO `sys_menu`(`id`, `parent_id`, `name`, `url`, `perms`, `type`, `icon`, `order_num`, `created_by`, `created_on`) VALUES (2, 1, '用户管理', 'sys/user', NULL, 1, 'icon_power_user_nav', 1, NULL, NULL);
INSERT INTO `sys_menu`(`id`, `parent_id`, `name`, `url`, `perms`, `type`, `icon`, `order_num`, `created_by`, `created_on`) VALUES (3, 1, '角色管理', 'sys/role', NULL, 1, 'icon_power_role_nav', 2, NULL, NULL);
INSERT INTO `sys_menu`(`id`, `parent_id`, `name`, `url`, `perms`, `type`, `icon`, `order_num`, `created_by`, `created_on`) VALUES (4, 1, '菜单管理', 'sys/menu', NULL, 1, 'nested', 4, NULL, NULL);
INSERT INTO `sys_menu`(`id`, `parent_id`, `name`, `url`, `perms`, `type`, `icon`, `order_num`, `created_by`, `created_on`) VALUES (5, 1, 'SQL监控', 'http://localhost:8080/druid/sql.html', NULL, 1, 'sql', 5, NULL, NULL);
INSERT INTO `sys_menu`(`id`, `parent_id`, `name`, `url`, `perms`, `type`, `icon`, `order_num`, `created_by`, `created_on`) VALUES (6, 1, '定时任务', 'sys/schedule', NULL, 1, 'job', 5, NULL, NULL);
INSERT INTO `sys_menu`(`id`, `parent_id`, `name`, `url`, `perms`, `type`, `icon`, `order_num`, `created_by`, `created_on`) VALUES (7, 6, '查看', NULL, 'sys:schedule:list,sys:schedule:info', 2, NULL, 0, NULL, NULL);
INSERT INTO `sys_menu`(`id`, `parent_id`, `name`, `url`, `perms`, `type`, `icon`, `order_num`, `created_by`, `created_on`) VALUES (8, 6, '新增', NULL, 'sys:schedule:save', 2, NULL, 0, NULL, NULL);
INSERT INTO `sys_menu`(`id`, `parent_id`, `name`, `url`, `perms`, `type`, `icon`, `order_num`, `created_by`, `created_on`) VALUES (9, 6, '修改', NULL, 'sys:schedule:update', 2, NULL, 0, NULL, NULL);
INSERT INTO `sys_menu`(`id`, `parent_id`, `name`, `url`, `perms`, `type`, `icon`, `order_num`, `created_by`, `created_on`) VALUES (10, 6, '删除', NULL, 'sys:schedule:delete', 2, NULL, 0, NULL, NULL);
INSERT INTO `sys_menu`(`id`, `parent_id`, `name`, `url`, `perms`, `type`, `icon`, `order_num`, `created_by`, `created_on`) VALUES (11, 6, '暂停', NULL, 'sys:schedule:pause', 2, NULL, 0, NULL, NULL);
INSERT INTO `sys_menu`(`id`, `parent_id`, `name`, `url`, `perms`, `type`, `icon`, `order_num`, `created_by`, `created_on`) VALUES (12, 6, '恢复', NULL, 'sys:schedule:resume', 2, NULL, 0, NULL, NULL);
INSERT INTO `sys_menu`(`id`, `parent_id`, `name`, `url`, `perms`, `type`, `icon`, `order_num`, `created_by`, `created_on`) VALUES (13, 6, '立即执行', NULL, 'sys:schedule:run', 2, NULL, 0, NULL, NULL);
INSERT INTO `sys_menu`(`id`, `parent_id`, `name`, `url`, `perms`, `type`, `icon`, `order_num`, `created_by`, `created_on`) VALUES (14, 6, '日志列表', NULL, 'sys:schedule:log', 2, NULL, 0, NULL, NULL);
INSERT INTO `sys_menu`(`id`, `parent_id`, `name`, `url`, `perms`, `type`, `icon`, `order_num`, `created_by`, `created_on`) VALUES (15, 2, '查看', NULL, 'sys:user:list,sys:user:info', 2, NULL, 0, NULL, NULL);
INSERT INTO `sys_menu`(`id`, `parent_id`, `name`, `url`, `perms`, `type`, `icon`, `order_num`, `created_by`, `created_on`) VALUES (16, 2, '新增', NULL, 'sys:user:save,sys:role:select', 2, NULL, 0, NULL, NULL);
INSERT INTO `sys_menu`(`id`, `parent_id`, `name`, `url`, `perms`, `type`, `icon`, `order_num`, `created_by`, `created_on`) VALUES (17, 2, '修改', NULL, 'sys:user:update,sys:role:select', 2, NULL, 0, NULL, NULL);
INSERT INTO `sys_menu`(`id`, `parent_id`, `name`, `url`, `perms`, `type`, `icon`, `order_num`, `created_by`, `created_on`) VALUES (18, 2, '删除', NULL, 'sys:user:delete', 2, NULL, 0, NULL, NULL);
INSERT INTO `sys_menu`(`id`, `parent_id`, `name`, `url`, `perms`, `type`, `icon`, `order_num`, `created_by`, `created_on`) VALUES (19, 3, '查看', NULL, 'sys:role:list,sys:role:info', 2, NULL, 0, NULL, NULL);
INSERT INTO `sys_menu`(`id`, `parent_id`, `name`, `url`, `perms`, `type`, `icon`, `order_num`, `created_by`, `created_on`) VALUES (20, 3, '新增', NULL, 'sys:role:save,sys:menu:list', 2, NULL, 0, NULL, NULL);
INSERT INTO `sys_menu`(`id`, `parent_id`, `name`, `url`, `perms`, `type`, `icon`, `order_num`, `created_by`, `created_on`) VALUES (21, 3, '修改', NULL, 'sys:role:update,sys:menu:list', 2, NULL, 0, NULL, NULL);
INSERT INTO `sys_menu`(`id`, `parent_id`, `name`, `url`, `perms`, `type`, `icon`, `order_num`, `created_by`, `created_on`) VALUES (22, 3, '删除', NULL, 'sys:role:delete', 2, NULL, 0, NULL, NULL);
INSERT INTO `sys_menu`(`id`, `parent_id`, `name`, `url`, `perms`, `type`, `icon`, `order_num`, `created_by`, `created_on`) VALUES (23, 4, '查看', NULL, 'sys:menu:list,sys:menu:info', 2, NULL, 0, NULL, NULL);
INSERT INTO `sys_menu`(`id`, `parent_id`, `name`, `url`, `perms`, `type`, `icon`, `order_num`, `created_by`, `created_on`) VALUES (24, 4, '新增', NULL, 'sys:menu:save,sys:menu:select', 2, NULL, 0, NULL, NULL);
INSERT INTO `sys_menu`(`id`, `parent_id`, `name`, `url`, `perms`, `type`, `icon`, `order_num`, `created_by`, `created_on`) VALUES (25, 4, '修改', NULL, 'sys:menu:update,sys:menu:select', 2, NULL, 0, NULL, NULL);
INSERT INTO `sys_menu`(`id`, `parent_id`, `name`, `url`, `perms`, `type`, `icon`, `order_num`, `created_by`, `created_on`) VALUES (26, 4, '删除', NULL, 'sys:menu:delete', 2, NULL, 0, NULL, NULL);