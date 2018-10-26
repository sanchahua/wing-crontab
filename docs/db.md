CREATE TABLE `statistics` (
 `id` int(11) NOT NULL AUTO_INCREMENT,
 `cron_id` int(11) NOT NULL DEFAULT '0' COMMENT '定时任务id',
 `day` date NOT NULL COMMENT '日期 如2018-01-01',
 `success` int(11) NOT NULL COMMENT '成功的次数',
 `fail` int(11) NOT NULL COMMENT '失败的次数',
 `avg_use_time` int(11) NOT NULL DEFAULT '0' COMMENT '当天平均运行时长，单位毫秒',
 `max_use_time` int(11) NOT NULL DEFAULT '0' COMMENT '当天最大运行时长，单位毫秒',
 PRIMARY KEY (`id`),
 UNIQUE KEY `cron_id` (`cron_id`,`day`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=2497 DEFAULT CHARSET=utf8 COMMENT='统计信息';

CREATE TABLE `log` (
 `id` int(11) NOT NULL AUTO_INCREMENT,
 `dispatch_server` int(11) NOT NULL DEFAULT '0' COMMENT '分发服务器，即当时的leader服务器',
 `run_server` int(11) NOT NULL DEFAULT '0' COMMENT '定时任务最终在该服务器上运行',
 `cron_id` int(11) NOT NULL DEFAULT '0' COMMENT '定时任务id',
 `process_id` int(11) NOT NULL DEFAULT '0' COMMENT '进程id',
 `state` varchar(32) NOT NULL DEFAULT '' COMMENT '执行结果',
 `start_time` datetime NOT NULL COMMENT '命令开始执行的时间',
 `output` longtext NOT NULL COMMENT '执行命令输出',
 `use_time` bigint(20) NOT NULL COMMENT '执行命令耗时，单位为毫秒',
 `remark` varchar(1024) NOT NULL DEFAULT '' COMMENT '备注',
 PRIMARY KEY (`id`),
 KEY `state` (`state`),
 KEY `cron_id` (`cron_id`),
 KEY `use_time` (`use_time`),
 KEY `start_time` (`start_time`)
) ENGINE=InnoDB AUTO_INCREMENT=5029965 DEFAULT CHARSET=utf8;

CREATE TABLE `cron` (
 `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '自增id',
 `cron_set` varchar(128) NOT NULL DEFAULT '' COMMENT '定时任务配置，如：* * * * * *，这里精确到秒，前面的意思是每秒执行一次，分别对应，秒分时日月周',
 `start_time` datetime NOT NULL DEFAULT '1970-01-01 08:00:00' COMMENT '大于等于此时间才执行',
 `end_time` datetime NOT NULL DEFAULT '2099-01-01 08:00:00' COMMENT '小于此时间才执行',
 `command` varchar(2048) NOT NULL DEFAULT '' COMMENT '定时任务执行的命令',
 `stop` tinyint(4) NOT NULL DEFAULT '0' COMMENT '1停止执行，0非，0为默认值',
 `remark` varchar(1024) NOT NULL DEFAULT '' COMMENT '定时任务的备注信息',
 `is_mutex` int(11) NOT NULL DEFAULT '0' COMMENT '0可以并发执行 1严格互斥执行',
 `userid` int(11) NOT NULL DEFAULT '0' COMMENT '该定时任务由该用户添加',
 `blame` int(11) NOT NULL DEFAULT '0' COMMENT '责任人用户id',
 PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=1895 DEFAULT CHARSET=utf8;

CREATE TABLE `services` (
 `id` int(11) NOT NULL AUTO_INCREMENT,
 `name` varchar(256) NOT NULL DEFAULT '',
 `address` varchar(64) NOT NULL DEFAULT '' COMMENT 'ip地址',
 `is_leader` tinyint(4) NOT NULL DEFAULT '0' COMMENT '1为leader，默认值为0不是leader',
 `updated` int(11) NOT NULL DEFAULT '0',
 PRIMARY KEY (`id`),
 UNIQUE KEY `address` (`address`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=31 DEFAULT CHARSET=utf8;

CREATE TABLE `users` (
 `id` int(11) NOT NULL AUTO_INCREMENT,
 `user_name` varchar(64) CHARACTER SET utf8mb4 NOT NULL DEFAULT '',
 `password` varchar(128) NOT NULL DEFAULT '',
 `real_name` varchar(64) CHARACTER SET utf8mb4 NOT NULL DEFAULT '',
 `phone` varchar(12) NOT NULL DEFAULT '',
 `created` datetime NOT NULL,
 `updated` datetime NOT NULL,
 `enable` int(11) NOT NULL DEFAULT '1' COMMENT '1 启用 0禁用',
 `admin` smallint(6) NOT NULL DEFAULT '0' COMMENT '1管理员，0普通用户',
 `powers` bigint(20) NOT NULL DEFAULT '0' COMMENT '用户权限，使用方式 “与或非”',
 PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=13 DEFAULT CHARSET=utf8