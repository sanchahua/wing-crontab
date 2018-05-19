install consul
--------------
download
https://www.consul.io/downloads.html
run
````
./consul agent --dev
````
build
-----
````
./bin/build.sh
````
config
------
````
vim ./config/app.toml
````
run
----
````
./bin/wing-crontab
````
http api
----
全局说明
````
所有的api返回字段里面，如下字段意义相同
code 错误码 200为正常
data 具体的业务数据
message 具体的错误信息
````
1、查询定时任务执行日志
````
http://localhost:9990/log/list
协议  GET
参数                      类型     是否必须    说明
cron_id                 int        否         指定定时任务id
search                  string   否         模糊查询定时任务指定返回值
dispatch_server   string   否         指定调度服务器，如 127.0.0.1:9991
run_server           string   否         指定运行服务器，如 127.0.0.1:9991
page                     int        否         指定第几页，用于分页查询支持，默认为1，总页数=总条数/limit参数 向上取整
limit                      int        否         指定每页返回的条数，用于分页查询支持

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
