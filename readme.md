
go test 相关指令支持，将GOPATH指向项目的根目录和vendor目录
`export GOPATH=/Users/yuyi/Code/go/wing-crontab:/Users/yuyi/Code/go/wing-crontab/vendor`

数据库相关
-----
所有的增加、删除、更新操作，直接修改数据库，修改后重启软件
````
CREATE DATABASE `cron` /*!40100 DEFAULT CHARACTER SET utf8 */
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
 `is_mutex` int(11) NOT NULL DEFAULT '0' COMMENT '0可以并发执行 1严格互斥执行',
 PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=1634 DEFAULT CHARSET=utf8
````

````
CREATE TABLE `log` (
 `id` int(11) NOT NULL AUTO_INCREMENT,
 `cron_id` int(11) NOT NULL DEFAULT '0',
 `event` varchar(32) NOT NULL DEFAULT '' COMMENT '事件',
 `time` bigint(20) NOT NULL COMMENT '命令运行的时间(毫秒时间戳)',
 `output` longtext NOT NULL COMMENT '执行命令输出',
 `use_time` bigint(20) NOT NULL COMMENT '执行命令耗时，单位为毫秒',
 `dispatch_server` varchar(1024) NOT NULL DEFAULT '' COMMENT '调度server',
 `run_server` varchar(1024) NOT NULL DEFAULT '' COMMENT '该命令在那个节点上被执行（服务器）',
 `remark` varchar(1024) NOT NULL DEFAULT '' COMMENT '备注',
 PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=7730455 DEFAULT CHARSET=utf8
````

如何安装wing-crontab
编译wing-crontab，需要提前安装go环境
-----
````
./bin/build.sh
````
修改配置文件
------
````
vim ./config/app.toml

配置文件说明
##这个就是consul的服务地址了，本地调试模式使用127.0.0.1:8500即可
consul_address = "127.0.0.1:8500"
##服务名称，用于consul的服务注册
service_name = "wing-crontab-service"
##用来选leader时的竞争锁的key值
##集群内所有的节点的key要保持一致
lock_key = "wing-crontab-lock"
##tcp服务的监听地址
##不允许使用"0.0.0.0:9991"这种模式，必须要指定具体的ip和端口
##原因是这个监听的地址要用于服务发现
bind_address = "127.0.0.1:9991"
##http接口的服务监听端口，这个可以使用"0.0.0.0:9990"这种模式
http_bind_address = "127.0.0.1:9990"
##这个用于性能监控调优，如果不想打开，去掉留空就可以了
pprof_listen = "127.0.0.1:8880"
##日志级别
# 0 =	PanicLevel Level = iota
# 1 =	FatalLevel
# 2 =	ErrorLevel
# 3 =	WarnLevel
# 4 =	InfoLevel
# 5 =	DebugLevel
log_level=5
##如下几个字段是mysql相关的连接配置
mysql_user = "root"
mysql_password = "123456"
mysql_host = "127.0.0.1"
mysql_port = 3306
mysql_database = "cron"
mysql_charset = "utf8"
````
运行
----
````
./bin/wing-crontab
````