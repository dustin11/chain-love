
INSERT INTO sys_user (id, created_on, addr, nickname, email, mobile, state, avatar, country, province, city, account_part, created_by) VALUES
(1001, NOW(), '0xAbC1234567890abcdef', '管理员', 'admin@example.com', '13800000000', 0, '', 'China', 'Beijing', 'Beijing', 7, 1),
(1002, NOW(), '0xDeF9876543210fedcba', 'alice', 'alice@example.com', '13800000001', 0, 'avatar_alice.png', 'China', 'Shanghai', 'Shanghai', 7, 1),
(1003, NOW(), '0x11111111111111111111', 'bob', '', '', 0, '', '', '', '', 0, 1),
(1004, NOW(), '0x22222222222222222222', 'charlie', 'charlie@example.com', '13800000002', 0, '', 'China', 'Guangdong', 'Shenzhen', 5, 1);
