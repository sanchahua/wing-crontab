CREATE TABLE `statistics` (
 `id` int(11) NOT NULL AUTO_INCREMENT,
 `cron_id` int(11) NOT NULL DEFAULT '0' COMMENT '定时任务id',
 `day` date NOT NULL COMMENT '日期 如2018-01-01',
 `success` int(11) NOT NULL COMMENT '成功的次数',
 `fail` int(11) NOT NULL COMMENT '失败的次数',
 PRIMARY KEY (`id`),
 UNIQUE KEY `cron_id` (`cron_id`,`day`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=2318 DEFAULT CHARSET=utf8 COMMENT='统计信息';


CREATE TABLE `log` (
 `id` int(11) NOT NULL AUTO_INCREMENT,
 `cron_id` int(11) NOT NULL DEFAULT '0' COMMENT '定时任务id',
 `process_id` int(11) NOT NULL DEFAULT '0' COMMENT '进程id',
 `state` varchar(32) NOT NULL DEFAULT '' COMMENT '执行结果',
 `start_time` datetime NOT NULL COMMENT '命令开始执行的时间',
 `output` longtext NOT NULL COMMENT '执行命令输出',
 `use_time` bigint(20) NOT NULL COMMENT '执行命令耗时，单位为毫秒',
 `remark` varchar(1024) NOT NULL DEFAULT '' COMMENT '备注',
 PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=1696480 DEFAULT CHARSET=utf8;


CREATE TABLE `cron` (
 `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '自增id',
 `cron_set` varchar(128) NOT NULL DEFAULT '' COMMENT '定时任务配置，如：* * * * * *，这里精确到秒，前面的意思是每秒执行一次，分别对应，秒分时日月周',
 `start_time` datetime NOT NULL DEFAULT '1970-01-01 08:00:00' COMMENT '大于等于此时间才执行',
 `end_time` datetime NOT NULL DEFAULT '2099-01-01 08:00:00' COMMENT '小于此时间才执行',
 `command` varchar(2048) NOT NULL DEFAULT '' COMMENT '定时任务执行的命令',
 `stop` tinyint(4) NOT NULL DEFAULT '0' COMMENT '1停止执行，0非，0为默认值',
 `remark` varchar(1024) NOT NULL DEFAULT '' COMMENT '定时任务的备注信息',
 `is_mutex` int(11) NOT NULL DEFAULT '0' COMMENT '0可以并发执行 1严格互斥执行',
 PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=1885 DEFAULT CHARSET=utf8;