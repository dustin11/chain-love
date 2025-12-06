/*
 Navicat Premium Data Transfer

 Source Server         : localhost
 Source Server Type    : MySQL
 Source Server Version : 50729
 Source Host           : localhost:3307
 Source Schema         : smarterp

 Target Server Type    : MySQL
 Target Server Version : 50729
 File Encoding         : 65001

 Date: 23/08/2020 00:05:56
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for ds_goods
-- ----------------------------
DROP TABLE IF EXISTS `ds_goods`;
CREATE TABLE `ds_goods` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `ccid` bigint(20) DEFAULT NULL,
  `code` varchar(20) DEFAULT NULL COMMENT '编码',
  `name` varchar(100) DEFAULT NULL COMMENT '名称',
  `unit` varchar(30) DEFAULT NULL COMMENT '计量单位',
  `spec` varchar(30) DEFAULT NULL COMMENT '规格',
  `class` int(11) DEFAULT NULL COMMENT '分类id',
  `price` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '单价',
  `amount` bigint(20) NOT NULL DEFAULT '0' COMMENT '总数量',
  `init_amount` bigint(20) NOT NULL DEFAULT '0' COMMENT '初始数量',
  `remark` varchar(100) DEFAULT NULL COMMENT '备注',
  `created_by` bigint(20) DEFAULT NULL COMMENT '创建人',
  `created_on` datetime DEFAULT NULL COMMENT '创建时间',
  `updated_on` datetime DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=7 DEFAULT CHARSET=utf8mb4 COMMENT='货品库存';

-- ----------------------------
-- Records of ds_goods
-- ----------------------------
BEGIN;
INSERT INTO `ds_goods` VALUES (1, 10, 'YB', '雨布', '1', '1', 0, 21.12, 99, 10, '载需要', 1, '2020-08-05 16:50:52', '2020-08-16 20:50:54');
INSERT INTO `ds_goods` VALUES (2, 10, 'YI', '雨衣', '1', '2', 0, 43.00, 16, 3, '', 1, '2020-08-15 15:53:33', '2020-08-16 21:00:28');
COMMIT;

-- ----------------------------
-- Table structure for ds_goods_class
-- ----------------------------
DROP TABLE IF EXISTS `ds_goods_class`;
CREATE TABLE `ds_goods_class` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `ccid` bigint(20) DEFAULT NULL,
  `name` varchar(50) DEFAULT NULL COMMENT '名称',
  `remark` varchar(100) DEFAULT NULL COMMENT '备注',
  `created_by` bigint(20) DEFAULT NULL COMMENT '创建人',
  `created_on` datetime DEFAULT NULL COMMENT '创建时间',
  `updated_on` datetime DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='货品分类';

-- ----------------------------
-- Table structure for ds_instore
-- ----------------------------
DROP TABLE IF EXISTS `ds_instore`;
CREATE TABLE `ds_instore` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `ccid` bigint(20) DEFAULT NULL,
  `code` varchar(20) DEFAULT NULL COMMENT '编码',
  `goods_id` bigint(20) DEFAULT NULL COMMENT '货品id',
  `goods_name` varchar(100) DEFAULT NULL COMMENT '名称',
  `goods_code` varchar(100) DEFAULT NULL COMMENT '货品编码',
  `intime` datetime DEFAULT NULL COMMENT '入库时间',
  `amount` int(11) NOT NULL DEFAULT '0' COMMENT '入库数量',
  `remark` varchar(100) DEFAULT NULL COMMENT '备注',
  `created_by` bigint(20) DEFAULT NULL COMMENT '创建人',
  `created_on` datetime DEFAULT NULL COMMENT '创建时间',
  `updated_on` datetime DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=6 DEFAULT CHARSET=utf8mb4 COMMENT='入库表';

-- ----------------------------
-- Records of ds_instore
-- ----------------------------
BEGIN;
INSERT INTO `ds_instore` VALUES (1, 10, NULL, 1, '雨布', 'YB', '2020-08-15 17:31:05', 1, '', 1, '2020-08-15 17:31:11', '2020-08-15 17:31:11');
INSERT INTO `ds_instore` VALUES (2, 10, NULL, 2, '雨衣', 'YI', '2020-08-15 17:43:28', 2, '买了两件', 1, '2020-08-15 17:43:28', '2020-08-15 17:43:28');
INSERT INTO `ds_instore` VALUES (3, 10, NULL, 2, '雨衣', 'YI', '2020-08-15 17:45:40', 2, '买了两件1', 1, '2020-08-15 17:45:40', '2020-08-15 17:50:11');
INSERT INTO `ds_instore` VALUES (4, 10, NULL, 2, '雨衣', 'YI', '2020-08-16 08:33:26', 2, '222', 1, '2020-08-16 08:33:26', '2020-08-16 08:33:26');
INSERT INTO `ds_instore` VALUES (5, 10, NULL, 2, '雨衣', 'YI', '2020-08-16 14:55:04', 2, '', 1, '2020-08-16 14:55:04', '2020-08-16 14:55:04');
COMMIT;

-- ----------------------------
-- Table structure for ds_outstore
-- ----------------------------
DROP TABLE IF EXISTS `ds_outstore`;
CREATE TABLE `ds_outstore` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `ccid` bigint(20) DEFAULT NULL,
  `code` varchar(20) DEFAULT NULL COMMENT '编码',
  `goods_id` bigint(20) DEFAULT NULL COMMENT '货品id',
  `goods_name` varchar(100) DEFAULT NULL COMMENT '名称',
  `goods_code` varchar(100) DEFAULT NULL COMMENT '货品编码',
  `ontime` datetime DEFAULT NULL COMMENT '出库时间',
  `amount` int(11) NOT NULL DEFAULT '0' COMMENT '出库数量',
  `remark` varchar(100) DEFAULT NULL COMMENT '备注',
  `created_by` bigint(20) DEFAULT NULL COMMENT '创建人',
  `created_on` datetime DEFAULT NULL COMMENT '创建时间',
  `updated_on` datetime DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4 COMMENT='出库表';

-- ----------------------------
-- Records of ds_outstore
-- ----------------------------
BEGIN;
INSERT INTO `ds_outstore` VALUES (1, 10, NULL, 1, '雨布', 'YB', '2020-08-16 20:50:53', 1, '11', 1, '2020-08-16 20:50:53', '2020-08-16 20:50:53');
INSERT INTO `ds_outstore` VALUES (2, 10, NULL, 2, '雨衣', 'YI', '2020-08-16 21:00:28', 2, '111', 1, '2020-08-16 21:00:28', '2020-08-16 21:00:28');
COMMIT;

-- ----------------------------
-- Table structure for ds_spec
-- ----------------------------
DROP TABLE IF EXISTS `ds_spec`;
CREATE TABLE `ds_spec` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `ccid` bigint(20) DEFAULT NULL,
  `code` varchar(30) DEFAULT NULL COMMENT '编码',
  `name` varchar(50) DEFAULT NULL COMMENT '名称',
  `remark` varchar(100) DEFAULT NULL COMMENT '备注',
  `created_by` bigint(20) DEFAULT NULL COMMENT '创建人',
  `created_on` datetime DEFAULT NULL COMMENT '创建时间',
  `updated_on` datetime DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4 COMMENT='规格';

-- ----------------------------
-- Records of ds_spec
-- ----------------------------
BEGIN;
INSERT INTO `ds_spec` VALUES (1, 10, NULL, '大件', NULL, 1, '2020-08-05 15:54:03', '2020-08-05 15:54:10');
INSERT INTO `ds_spec` VALUES (2, 10, NULL, '小件', 'ss', 1, '2020-08-05 15:54:36', '2020-08-16 14:53:15');
COMMIT;

-- ----------------------------
-- Table structure for ds_unit
-- ----------------------------
DROP TABLE IF EXISTS `ds_unit`;
CREATE TABLE `ds_unit` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `ccid` bigint(20) DEFAULT NULL,
  `code` varchar(30) DEFAULT NULL COMMENT '编码',
  `name` varchar(50) DEFAULT NULL COMMENT '名称',
  `remark` varchar(100) DEFAULT NULL COMMENT '备注',
  `created_by` bigint(20) DEFAULT NULL COMMENT '创建人',
  `created_on` datetime DEFAULT NULL COMMENT '创建时间',
  `updated_on` datetime DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=23 DEFAULT CHARSET=utf8mb4 COMMENT='计量单位';

-- ----------------------------
-- Records of ds_unit
-- ----------------------------
BEGIN;
INSERT INTO `ds_unit` VALUES (1, 10, 'c1', '米', NULL, 1, '2020-08-02 22:29:25', '2020-08-02 22:29:32');
INSERT INTO `ds_unit` VALUES (2, 10, 'g', '斤', NULL, 1, '2020-08-03 18:19:33', '2020-08-03 18:19:39');
INSERT INTO `ds_unit` VALUES (4, 10, 'K', '颗', 'kk', 1, '2020-08-16 13:01:20', '2020-08-16 12:08:50');
COMMIT;

-- ----------------------------
-- Table structure for sys_menu
-- ----------------------------
DROP TABLE IF EXISTS `sys_menu`;
CREATE TABLE `sys_menu` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `parent_id` bigint(20) DEFAULT NULL COMMENT '父菜单ID，一级菜单为0',
  `name` varchar(50) DEFAULT NULL COMMENT '菜单名称',
  `url` varchar(200) DEFAULT NULL COMMENT '菜单URL',
  `perms` varchar(500) DEFAULT NULL COMMENT '授权(多个用逗号分隔，如：user:list,user:create)',
  `type` int(11) DEFAULT NULL COMMENT '类型   0：目录   1：菜单   2：按钮',
  `icon` varchar(50) DEFAULT NULL COMMENT '菜单图标',
  `order_num` int(11) DEFAULT NULL COMMENT '排序',
  `created_by` bigint(20) DEFAULT NULL COMMENT '创建人',
  `created_on` datetime DEFAULT NULL COMMENT '创建时间',
  `updated_on` datetime DEFAULT NULL COMMENT '修改时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=48 DEFAULT CHARSET=utf8mb4 COMMENT='菜单管理';

-- ----------------------------
-- Records of sys_menu
-- ----------------------------
BEGIN;
INSERT INTO `sys_menu` VALUES (1, 0, '系统管理', '', '', 0, 'shezhi', 0, 0, '0001-01-01 00:00:00', '2020-08-16 14:33:36');
INSERT INTO `sys_menu` VALUES (2, 1, '用户管理', 'sys/user', NULL, 1, 'admin', 1, NULL, NULL, NULL);
INSERT INTO `sys_menu` VALUES (3, 1, '角色管理', 'sys/role', '', 1, 'role', 2, 0, '0001-01-01 00:00:00', '2020-08-22 00:04:21');
INSERT INTO `sys_menu` VALUES (4, 1, '菜单管理', 'sys/menu', '', 1, 'menu', 4, 0, '0001-01-01 00:00:00', '2020-08-16 20:47:45');
INSERT INTO `sys_menu` VALUES (15, 2, '查看', NULL, 'sys:user:list,sys:user:info', 2, NULL, 0, NULL, NULL, NULL);
INSERT INTO `sys_menu` VALUES (16, 2, '新增', NULL, 'sys:user:save,sys:role:select', 2, NULL, 0, NULL, NULL, NULL);
INSERT INTO `sys_menu` VALUES (17, 2, '修改', NULL, 'sys:user:update,sys:role:select', 2, NULL, 0, NULL, NULL, NULL);
INSERT INTO `sys_menu` VALUES (18, 2, '删除', NULL, 'sys:user:delete', 2, NULL, 0, NULL, NULL, NULL);
INSERT INTO `sys_menu` VALUES (19, 3, '查看', NULL, 'sys:role:list,sys:role:info', 2, NULL, 0, NULL, NULL, NULL);
INSERT INTO `sys_menu` VALUES (20, 3, '新增', NULL, 'sys:role:save,sys:menu:list', 2, NULL, 0, NULL, NULL, NULL);
INSERT INTO `sys_menu` VALUES (21, 3, '修改', NULL, 'sys:role:update,sys:menu:list', 2, NULL, 0, NULL, NULL, NULL);
INSERT INTO `sys_menu` VALUES (22, 3, '删除', NULL, 'sys:role:delete', 2, NULL, 0, NULL, NULL, NULL);
INSERT INTO `sys_menu` VALUES (23, 4, '查看', NULL, 'sys:menu:list,sys:menu:info', 2, NULL, 0, NULL, NULL, NULL);
INSERT INTO `sys_menu` VALUES (24, 4, '新增', NULL, 'sys:menu:save,sys:menu:select', 2, NULL, 0, NULL, NULL, NULL);
INSERT INTO `sys_menu` VALUES (25, 4, '修改', NULL, 'sys:menu:update,sys:menu:select', 2, NULL, 0, NULL, NULL, NULL);
INSERT INTO `sys_menu` VALUES (26, 4, '删除', NULL, 'sys:menu:delete', 2, NULL, 0, NULL, NULL, NULL);
INSERT INTO `sys_menu` VALUES (27, 0, '物资管理', '', '', 0, 'log', 2, 0, '0001-01-01 00:00:00', '2020-08-16 14:35:34');
INSERT INTO `sys_menu` VALUES (28, 27, '物资管理', '/ds/goods', '', 1, 'zonghe', 2, 0, '0001-01-01 00:00:00', '2020-08-16 14:34:12');
INSERT INTO `sys_menu` VALUES (29, 28, '添加', '', 'ds:goods:add', 2, '', 0, 1, '2020-08-02 19:29:11', '2020-08-02 19:29:11');
INSERT INTO `sys_menu` VALUES (30, 28, '修改', '', 'ds:goods:update', 2, '', 0, 1, '2020-08-02 19:29:42', '2020-08-02 19:29:42');
INSERT INTO `sys_menu` VALUES (31, 28, '删除', '', 'ds:goods:del', 2, '', 0, 1, '2020-08-02 19:30:07', '2020-08-02 19:30:07');
INSERT INTO `sys_menu` VALUES (32, 43, '入库', 'ds/instore', '', 1, 'zhedie', 3, 0, '0001-01-01 00:00:00', '2020-08-16 20:16:26');
INSERT INTO `sys_menu` VALUES (33, 27, '计量单位', '/ds/unit', 'ds:unit', 1, 'daohang', 1, 0, NULL, NULL);
INSERT INTO `sys_menu` VALUES (34, 27, '规格', '/ds/spec', 'ds:spec', 1, 'daohang', 0, 1, NULL, NULL);
INSERT INTO `sys_menu` VALUES (35, 33, '添加', '', 'ds:unit:add', 2, '', 0, 1, NULL, NULL);
INSERT INTO `sys_menu` VALUES (36, 33, '修改', '', 'ds:unit:update', 2, '', 0, 1, NULL, NULL);
INSERT INTO `sys_menu` VALUES (37, 33, '删除', '', 'ds:unit:del', 2, '', 0, 1, NULL, NULL);
INSERT INTO `sys_menu` VALUES (38, 34, '添加', '', 'ds:spec:add', 2, '', 0, 1, NULL, NULL);
INSERT INTO `sys_menu` VALUES (39, 34, '修改', '', 'ds:spec:update', 2, '', 0, 1, NULL, NULL);
INSERT INTO `sys_menu` VALUES (40, 34, '删除', '', 'ds:spec:del', 2, '', 0, 1, NULL, NULL);
INSERT INTO `sys_menu` VALUES (41, 32, '添加', '', 'ds:instore:add', 2, '', 0, 1, '2020-08-16 14:44:24', '2020-08-16 14:44:24');
INSERT INTO `sys_menu` VALUES (42, 32, '修改', '', 'ds:instore:update', 2, '', 0, 1, '2020-08-16 14:44:58', '2020-08-16 14:44:58');
INSERT INTO `sys_menu` VALUES (43, 0, '仓库管理', '', '', 0, 'shouye', 3, 1, '2020-08-16 20:09:47', '2020-08-16 20:09:47');
INSERT INTO `sys_menu` VALUES (44, 43, '出库', 'ds/outstore', '', 1, 'menu', 4, 1, '2020-08-16 20:13:52', '2020-08-16 20:13:52');
INSERT INTO `sys_menu` VALUES (45, 44, '添加', '', 'ds:outstore:add', 2, '', 0, 0, '0001-01-01 00:00:00', '2020-08-16 20:16:01');
INSERT INTO `sys_menu` VALUES (46, 44, '修改', '', 'ds:outstore:update', 2, '', 0, 1, '2020-08-16 20:15:45', '2020-08-16 20:15:45');
COMMIT;

-- ----------------------------
-- Table structure for sys_role
-- ----------------------------
DROP TABLE IF EXISTS `sys_role`;
CREATE TABLE `sys_role` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `ccid` bigint(20) DEFAULT NULL,
  `name` varchar(100) DEFAULT NULL COMMENT '角色名称',
  `remark` varchar(100) DEFAULT NULL COMMENT '备注',
  `created_by` bigint(20) DEFAULT NULL COMMENT '创建人',
  `created_on` datetime DEFAULT NULL COMMENT '创建时间',
  `updated_on` datetime DEFAULT NULL COMMENT '修改时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=6 DEFAULT CHARSET=utf8mb4 COMMENT='角色';

-- ----------------------------
-- Records of sys_role
-- ----------------------------
BEGIN;
INSERT INTO `sys_role` VALUES (1, 10, '管理员', NULL, 1, '2020-05-17 16:56:31', '2020-08-21 23:54:25');
INSERT INTO `sys_role` VALUES (2, 10, '仓库管理员', NULL, 1, '2020-05-17 16:56:52', '2020-08-21 23:53:59');
INSERT INTO `sys_role` VALUES (3, 10, '人事', '', 1, '0000-00-00 00:00:00', '2020-08-22 17:42:41');
COMMIT;

-- ----------------------------
-- Table structure for sys_role_menu
-- ----------------------------
DROP TABLE IF EXISTS `sys_role_menu`;
CREATE TABLE `sys_role_menu` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `role_id` bigint(20) DEFAULT NULL COMMENT '角色ID',
  `menu_id` bigint(20) DEFAULT NULL COMMENT '菜单ID',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=340 DEFAULT CHARSET=utf8mb4 COMMENT='角色与菜单对应关系';

-- ----------------------------
-- Records of sys_role_menu
-- ----------------------------
BEGIN;
INSERT INTO `sys_role_menu` VALUES (68, 0, 15);
INSERT INTO `sys_role_menu` VALUES (69, 0, 3);
INSERT INTO `sys_role_menu` VALUES (70, 0, 19);
INSERT INTO `sys_role_menu` VALUES (71, 0, -666666);
INSERT INTO `sys_role_menu` VALUES (72, 0, 1);
INSERT INTO `sys_role_menu` VALUES (73, 0, 2);
INSERT INTO `sys_role_menu` VALUES (259, 2, 27);
INSERT INTO `sys_role_menu` VALUES (260, 2, 28);
INSERT INTO `sys_role_menu` VALUES (261, 2, 29);
INSERT INTO `sys_role_menu` VALUES (262, 2, 30);
INSERT INTO `sys_role_menu` VALUES (263, 2, 31);
INSERT INTO `sys_role_menu` VALUES (264, 2, 33);
INSERT INTO `sys_role_menu` VALUES (265, 2, 35);
INSERT INTO `sys_role_menu` VALUES (266, 2, 36);
INSERT INTO `sys_role_menu` VALUES (267, 2, 37);
INSERT INTO `sys_role_menu` VALUES (268, 2, 34);
INSERT INTO `sys_role_menu` VALUES (269, 2, 38);
INSERT INTO `sys_role_menu` VALUES (270, 2, 39);
INSERT INTO `sys_role_menu` VALUES (271, 2, 40);
INSERT INTO `sys_role_menu` VALUES (272, 2, 43);
INSERT INTO `sys_role_menu` VALUES (273, 2, 32);
INSERT INTO `sys_role_menu` VALUES (274, 2, 41);
INSERT INTO `sys_role_menu` VALUES (275, 2, 42);
INSERT INTO `sys_role_menu` VALUES (276, 2, 44);
INSERT INTO `sys_role_menu` VALUES (277, 2, 45);
INSERT INTO `sys_role_menu` VALUES (278, 2, 46);
INSERT INTO `sys_role_menu` VALUES (279, 2, -666666);
INSERT INTO `sys_role_menu` VALUES (280, 1, 1);
INSERT INTO `sys_role_menu` VALUES (281, 1, 2);
INSERT INTO `sys_role_menu` VALUES (282, 1, 15);
INSERT INTO `sys_role_menu` VALUES (283, 1, 16);
INSERT INTO `sys_role_menu` VALUES (284, 1, 17);
INSERT INTO `sys_role_menu` VALUES (285, 1, 18);
INSERT INTO `sys_role_menu` VALUES (286, 1, 3);
INSERT INTO `sys_role_menu` VALUES (287, 1, 19);
INSERT INTO `sys_role_menu` VALUES (288, 1, 20);
INSERT INTO `sys_role_menu` VALUES (289, 1, 21);
INSERT INTO `sys_role_menu` VALUES (290, 1, 22);
INSERT INTO `sys_role_menu` VALUES (291, 1, 27);
INSERT INTO `sys_role_menu` VALUES (292, 1, 28);
INSERT INTO `sys_role_menu` VALUES (293, 1, 29);
INSERT INTO `sys_role_menu` VALUES (294, 1, 30);
INSERT INTO `sys_role_menu` VALUES (295, 1, 31);
INSERT INTO `sys_role_menu` VALUES (296, 1, 33);
INSERT INTO `sys_role_menu` VALUES (297, 1, 35);
INSERT INTO `sys_role_menu` VALUES (298, 1, 36);
INSERT INTO `sys_role_menu` VALUES (299, 1, 37);
INSERT INTO `sys_role_menu` VALUES (300, 1, 34);
INSERT INTO `sys_role_menu` VALUES (301, 1, 38);
INSERT INTO `sys_role_menu` VALUES (302, 1, 39);
INSERT INTO `sys_role_menu` VALUES (303, 1, 40);
INSERT INTO `sys_role_menu` VALUES (304, 1, 43);
INSERT INTO `sys_role_menu` VALUES (305, 1, 32);
INSERT INTO `sys_role_menu` VALUES (306, 1, 41);
INSERT INTO `sys_role_menu` VALUES (307, 1, 42);
INSERT INTO `sys_role_menu` VALUES (308, 1, 44);
INSERT INTO `sys_role_menu` VALUES (309, 1, 45);
INSERT INTO `sys_role_menu` VALUES (310, 1, 46);
INSERT INTO `sys_role_menu` VALUES (311, 1, -666666);
INSERT INTO `sys_role_menu` VALUES (329, 3, 15);
INSERT INTO `sys_role_menu` VALUES (330, 3, 16);
INSERT INTO `sys_role_menu` VALUES (331, 3, 17);
INSERT INTO `sys_role_menu` VALUES (332, 3, 19);
INSERT INTO `sys_role_menu` VALUES (333, 3, 20);
INSERT INTO `sys_role_menu` VALUES (334, 3, 21);
INSERT INTO `sys_role_menu` VALUES (335, 3, -666666);
INSERT INTO `sys_role_menu` VALUES (336, 3, 1);
INSERT INTO `sys_role_menu` VALUES (337, 3, 2);
INSERT INTO `sys_role_menu` VALUES (338, 3, 3);
INSERT INTO `sys_role_menu` VALUES (339, 0, -666666);
COMMIT;

-- ----------------------------
-- Table structure for sys_user_role
-- ----------------------------
DROP TABLE IF EXISTS `sys_user_role`;
CREATE TABLE `sys_user_role` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `user_id` bigint(20) DEFAULT NULL COMMENT '用户ID',
  `role_id` bigint(20) DEFAULT NULL COMMENT '角色ID',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=11 DEFAULT CHARSET=utf8mb4 COMMENT='用户与角色对应关系';

-- ----------------------------
-- Records of sys_user_role
-- ----------------------------
BEGIN;
INSERT INTO `sys_user_role` VALUES (5, 4, 2);
INSERT INTO `sys_user_role` VALUES (7, 2, 1);
INSERT INTO `sys_user_role` VALUES (8, 3, 2);
INSERT INTO `sys_user_role` VALUES (9, 1, 1);
COMMIT;

-- ----------------------------
-- Table structure for user_info
-- ----------------------------
DROP TABLE IF EXISTS `user_info`;
CREATE TABLE `user_info` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `ccid` bigint(20) DEFAULT NULL,
  `name` varchar(255) DEFAULT NULL,
  `password` varchar(255) DEFAULT NULL,
  `salt` varchar(255) DEFAULT NULL,
  `state` tinyint(4) NOT NULL,
  `username` varchar(255) DEFAULT NULL,
  `created_by` bigint(20) DEFAULT NULL COMMENT '创建人',
  `created_on` datetime DEFAULT NULL COMMENT '创建时间',
  `updated_on` datetime DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `UK_f2ksd6h8hsjtd57ipfq9myr64` (`username`)
) ENGINE=InnoDB AUTO_INCREMENT=12 DEFAULT CHARSET=utf8mb4 COMMENT='用户';

-- ----------------------------
-- Records of user_info
-- ----------------------------
BEGIN;
INSERT INTO `user_info` VALUES (1, 1, '管理员', 'admin', '8d78869f470951332959580424d4bf4f', 0, 'admin', 1, '2020-07-13 19:34:24', '2020-08-22 23:02:29');
INSERT INTO `user_info` VALUES (2, 10, 'jack li', 'qwe123', NULL, 0, 'jack', 1, '2020-07-13 19:34:24', '2020-08-21 22:55:06');
INSERT INTO `user_info` VALUES (3, 10, 'jack lily', '123456', NULL, 0, 'jim', 1, '2020-07-13 19:34:24', '2020-08-21 23:56:46');
INSERT INTO `user_info` VALUES (4, 10, 'jack hony', '123456', NULL, 0, 'jackho', 1, '2018-12-16 11:17:03', NULL);
COMMIT;

SET FOREIGN_KEY_CHECKS = 1;
