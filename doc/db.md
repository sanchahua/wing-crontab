##创建数据库
````
CREATE DATABASE `cron` /*!40100 DEFAULT CHARACTER SET utf8 */
````
##创建表
````
CREATE TABLE `cron` (
 `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '自增id',
 `cron_set` varchar(128) NOT NULL DEFAULT '' COMMENT '定时任务配置，如：* * * * * *，这里精确到秒，前面的意思是每秒执行一次，分别对应，秒分时日月周',
 `start_time` int(11) NOT NULL DEFAULT '0' COMMENT '大于等于此时间才执行，默认0',
 `end_time` int(11) NOT NULL DEFAULT '0' COMMENT '小于此时间才执行，默认0不限',
 `command` varchar(2048) NOT NULL DEFAULT '' COMMENT '定时任务执行的命令',
 `stop` tinyint(4) NOT NULL DEFAULT '0' COMMENT '1停止执行，0非，0为默认值',
 `remark` varchar(1024) NOT NULL DEFAULT '' COMMENT '定时任务的备注信息',
 PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=1184 DEFAULT CHARSET=utf8
````

````
CREATE TABLE `cron` (
 `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '自增id',
 `cron_set` varchar(128) NOT NULL DEFAULT '' COMMENT '定时任务配置，如：* * * * * *，这里精确到秒，前面的意思是每秒执行一次，分别对应，秒分时日月周',
 `start_time` int(11) NOT NULL DEFAULT '0' COMMENT '大于等于此时间才执行，默认0',
 `end_time` int(11) NOT NULL DEFAULT '0' COMMENT '小于此时间才执行，默认0不限',
 `command` varchar(2048) NOT NULL DEFAULT '' COMMENT '定时任务执行的命令',
 `stop` tinyint(4) NOT NULL DEFAULT '0' COMMENT '1停止执行，0非，0为默认值',
 `remark` varchar(1024) NOT NULL DEFAULT '' COMMENT '定时任务的备注信息',
 `is_metux` int(11) NOT NULL DEFAULT '0' COMMENT '0可以并发执行 1严格互斥执行',
 PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=1538 DEFAULT CHARSET=utf8
````