create table user_info
(
    id         bigint auto_increment
        primary key,
    ccid       bigint null,
    name       varchar(255) null,
    password   varchar(255) null,
    salt       varchar(255) null,
    state      tinyint      not null,
    username   varchar(255) null,
    created_by bigint       null comment '创建人',
    created_on    datetime     null comment '创建时间',
    updated_on      datetime     null comment '更新时间',
    constraint UK_f2ksd6h8hsjtd57ipfq9myr64
        unique (username)
)
comment '用户' charset = utf8mb4;


create table sys_menu
(
    id             bigint auto_increment
        primary key,
    parent_id      bigint       null comment '父菜单ID，一级菜单为0',
    name           varchar(50)  null comment '菜单名称',
    url            varchar(200) null comment '菜单URL',
    perms          varchar(500) null comment '授权(多个用逗号分隔，如：user:list,user:create)',
    type           int          null comment '类型   0：目录   1：菜单   2：按钮',
    icon           varchar(50)  null comment '菜单图标',
    order_num      int          null comment '排序',
    created_by bigint       null comment '创建人',
    created_on    datetime     null comment '创建时间',
    updated_on      datetime     null comment '更新时间'
)
    comment '菜单' charset = utf8mb4;
go

create table sys_role
(
    id             bigint auto_increment
        primary key,
    ccid       bigint null,
    name      varchar(100) null comment '角色名称',
    remark         varchar(100) null comment '备注',
    created_by bigint       null comment '创建人',
    created_on    datetime     null comment '创建时间',
    updated_on      datetime     null comment '更新时间'
)
    comment '角色' charset = utf8mb4;


create table sys_user_role
(
    id      bigint auto_increment
        primary key,
    ccid       bigint null,
    user_id bigint null comment '用户ID',
    role_id bigint null comment '角色ID'
)
    comment '用户与角色对应关系' charset = utf8mb4;


create table sys_role_menu
(
    id      bigint auto_increment
        primary key,
    ccid       bigint null,
    role_id bigint null comment '角色ID',
    menu_id bigint null comment '菜单ID'
)
    comment '角色与菜单对应关系' charset = utf8mb4;


//=======================================业务表
create table ds_unit
(
	id		int	auto_increment	primary key,
	ccid       bigint null,
	code	varchar(30)	null comment	'编码',
  name	varchar(50)	null comment	'名称',
  remark	varchar(100)	null	comment	'备注',
  created_by	bigint	null	comment	'创建人',
    created_on	datetime	null	comment	'创建时间',
	updated_on	datetime	null	comment	'更新时间'
)
comment '计量单位' charset = utf8mb4;
go

go
CREATE	TABLE	ds_spec
(
	id		int	auto_increment	primary	key,
	ccid       bigint null,
	code	varchar(30)	null	comment	'编码',
	name	varchar(50)	null	comment	'名称',
	remark	varchar(100)	null	comment '备注',
	created_by	bigint	null	comment	'创建人',
	created_on	datetime	null	comment	'创建时间',
	updated_on	datetime	null	comment	'更新时间'
)
comment	'规格'	charset	=	utf8mb4;
go
CREATE	TABLE	ds_goods_class
(
    id	int	auto_increment	primary	key,
    ccid       bigint null,
    name	varchar(50)	null	comment	'名称',
    remark  varchar(100) null	comment	'备注',
    created_by	bigint	null	comment	'创建人',
    created_on	datetime	null	comment	'创建时间',
    updated_on	datetime	null	comment	'更新时间'
)
comment '货品分类' charset = utf8mb4;

create	table	ds_goods
(
	id		bigint	auto_increment	primary	key,
	ccid       bigint null,
	code	varchar(20)	null	comment	'编码',
	name	varchar(100)	null	comment '名称',
	unit	int	null	comment	'计量单位',
	spec	int	null	comment	'规格',
	class	int	null	comment	'分类id',
	price		decimal(10,2)	not	null	default	0	comment	'单价',
	amount	bigint	not	null	default	0	comment	'总数量',
	init_amount	bigint	not	null	default	0	comment	'初始数量',
	remark	varchar(100)	null	comment	'备注',
	created_by	bigint	null	comment	'创建人',
	created_on	datetime	null	comment	'创建时间',
	updated_on	datetime	null	comment	'更新时间'
)
comment	'货品库存'	charset	=	utf8mb4;

create	table	ds_instore
(
	id		bigint	auto_increment	primary	key,
	ccid       bigint null,
	code	varchar(20)	null	comment	'编码',
	goods_id	bigint	null	comment	'货品id',
	goods_name	varchar(100)	null	comment '名称',
	goods_code	varchar(100)	null	comment '货品编码',
	intime	datetime	null	comment	'入库时间',
	amount	int	not	null	default	0	comment	'入库数量',
	remark	varchar(100)	null	comment	'备注',
	created_by	bigint	null	comment	'创建人',
	created_on	datetime	null	comment	'创建时间',
	updated_on	datetime	null	comment	'更新时间'
)
comment	'入库表'	charset	=	utf8mb4;

create	table	ds_outstore
(
	id		bigint	auto_increment	primary	key,
	ccid       bigint null,
	code	varchar(20)	null	comment	'编码',
	goods_id	bigint	null	comment	'货品id',
	goods_name	varchar(100)	null	comment '名称',
	goods_code	varchar(100)	null	comment '货品编码',
	ontime	datetime	null	comment	'出库时间',
	amount	int	not	null	default	0	comment	'出库数量',
	remark	varchar(100)	null	comment	'备注',
	created_by	bigint	null	comment	'创建人',
	created_on	datetime	null	comment	'创建时间',
	updated_on	datetime	null	comment	'更新时间'
)
comment	'出库表'	charset	=	utf8mb4;