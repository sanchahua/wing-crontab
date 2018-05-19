wing-crontab功能说明
----------------
1、支持定时任务的实时增删改查、暂停、开始和指定运行时间范围
2、支持定时任务执行日志
3、过载保护
4、支持指定严格互斥的运行模式，即同一时间内只能用一个在运行

数据库相关
-----
所有的操作，如需立即生效，不可以直接修改数据库，请使用api
直接手动修改数据库的操作如需生效，请重启wing-crontab
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
 `is_mutex` int(11) NOT NULL DEFAULT '0' COMMENT '0可以并发执行 1严格互斥执行',
 PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=1538 DEFAULT CHARSET=utf8
````

如何安装wing-crontab
--
1、安装 consul
--------------
下载
https://www.consul.io/downloads.html
如下运行dev调试版本，单机模式下使用，非常简单，consul的监听端口一般为8500，即127.0.0.1:8500
````
./consul agent --dev
````
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
http接口
----
全局说明
````
所有的api返回字段里面，如下字段意义相同
code 错误码 200为正常
data 具体的业务数据
message 具体的错误信息
````

````
GET  /cron/list
GET  /cron/stop/{id}
GET  /cron/start/{id}
GET  /cron/delete/{id}
POST /cron/update
POST /cron/add
````
查询定时任务执行日志
````
http://localhost:9990/log/list
协议  GET
参数               类型     是否必须    说明
cron_id           int      否         指定定时任务id
search            string   否         模糊查询定时任务指定返回值
dispatch_server   string   否         指定调度服务器，如 127.0.0.1:9991
run_server        string   否         指定运行服务器，如 127.0.0.1:9991
page              int      否         指定第几页，用于分页查询支持，默认为1，总页数=总条数/limit参数 向上取整
limit             int      否         指定每页返回的条数，用于分页查询支持

demo http://localhost:9990/log/list?limit=1 返回值如下
{
    "code": 200,
    "data": {
        "list": [
            {
                "id": 6620495,
                "cron_id": 1633,
                "time": 1526772008,
                "output": "  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current\n                                 Dload  Upload   Total   Spent    Left  Speed\n\r  0     0    0     0    0     0      0      0 --:--:-- --:--:-- --:--:--     0\r100   207  100   207    0     0  21219      0 --:--:-- --:--:-- --:--:-- 23000\n<!DOCTYPE HTML PUBLIC \"-//IETF//DTD HTML 2.0//EN\">\n<html><head>\n<title>404 Not Found</title>\n</head><body>\n<h1>Not Found</h1>\n<p>The requested URL /index.php was not found on this server.</p>\n</body></html>\n",
                "use_time": 48,
                "dispatch_server": "127.0.0.1:9991",
                "run_server": "127.0.0.1:9991"
            }
        ],
        "total": 125784
    },
    "message": "ok"
}
list 数据列表，这里是一个数组，按照limit指定返回条数，数量小于等于limit
total 为系统数据条数
id 为日志id
crom_id 为定时任务id，对应cron表的id
time 为定时任务指定的时间戳
output 为定时任务指定的输出结果
use_time 为定时任务执行耗时时长，单位为毫秒
dispatch_server 为调度服务器
run_server 为最终运行定时任务的服务器
````
