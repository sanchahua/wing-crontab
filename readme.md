
main函数：src/xcrontab/main.go

go test 相关指令支持，将GOPATH指向项目的根目录和vendor目录，注意一下目录换成自己的真实项目路径
`export GOPATH={project path}:{project path}/vendor`

数据库相关
-----
所有的增加、删除、更新操作，直接修改数据库，修改后重启软件
````
CREATE DATABASE `cron` /*!40100 DEFAULT CHARACTER SET utf8 */;
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
) ENGINE=InnoDB AUTO_INCREMENT=1634 DEFAULT CHARSET=utf8mb4;
````

````
CREATE TABLE `log` (
 `id` int(11) NOT NULL AUTO_INCREMENT,
 `cron_id` int(11) NOT NULL DEFAULT '0' COMMENT '定时任务id',
 `start_time` datetime NOT NULL COMMENT '命令开始执行的时间',
 `output` longtext NOT NULL COMMENT '执行命令输出',
 `use_time` bigint(20) NOT NULL COMMENT '执行命令耗时，单位为毫秒',
 `remark` varchar(1024) NOT NULL DEFAULT '' COMMENT '备注',
 PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=706922 DEFAULT CHARSET=utf8mb4;
````

如何安装xcrontab
编译xcrontab，需要提前安装go环境
-----
````
./bin/build.sh debug ##发布debug版本
./bin/build.sh ## 默认无参数发布release版本
````
修改配置文件
------
````
vim ./config/mysql.toml

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
debug模式
./bin/xcrontab
````
````
后台运行
./bin/start.sh
````
````
停止运行
./bin/stop.sh
````
````
重新加载配置(修改数据库后执行此命令重新加载定时任务)
./bin/reload.sh
````