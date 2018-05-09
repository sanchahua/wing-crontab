http://localhost:9990/cron/list    
--获取数据库定时任务列表    
--返回结果    
````
{
    "code": 200,
    "data": [
        {
            "Id": 1,
            "CronSet": "* * * * * *",
            "Command": "ls /",
            "IsMutex": true,
            "Remark": "123",
            "Stop": false,
            "CronId": 0,
            "RunTimes": 0,
            "FailureTimes": 0,
            "LastRunTime": 0,
            "LastLockSuccessTime": 0,
            "LastUnLockSuccessTime": 0,
            "RunTimeTotal": 0
        },
        {
            "Id": 4,
            "CronSet": "* * * * * *",
            "Command": "ls /",
            "IsMutex": true,
            "Remark": "12345",
            "Stop": false,
            "CronId": 0,
            "RunTimes": 0,
            "FailureTimes": 0,
            "LastRunTime": 0,
            "LastLockSuccessTime": 0,
            "LastUnLockSuccessTime": 0,
            "RunTimeTotal": 0
        }
    ],
    "message": "ok"
}
````
--返回值说明
这里的如下字段为运行时字段，在此接口内没有意义，忽略
````
"CronId": 0,
"RunTimes": 0,
"FailureTimes": 0,
"LastRunTime": 0,
"LastLockSuccessTime": 0,
"LastUnLockSuccessTime": 0,
"RunTimeTotal": 0
````
curl http://localhost:9990/cron/add -X POST -d "cronSet=*/1 * * * * *&command=curl http://test.com/index.php&isMutex=1&remark=&lockLimit=10&stop=0" 
--添加定时任务 -- 支持 post 和 get    
--参数    
cronSet 定时设置，比如 0 */1 * * * *    
command 执行命令，如 php -v    
isMutex 是否互斥运行，1是，0否    
remark 备注，可以为空    
--返回值    
````
{
    "code": 200,
    "data": 12,   ##数据库id
    "message": "ok"
}
````

curl http://localhost:9990/cron/add -X POST -d "cronSet=*/1 * * * * *&command=curl http://test.com/index.php&remark=&stop=0"
curl http://localhost:9992/cron/add -X POST -d "cronSet=*/5 * * * * *&command=php -v&remark=&stop=0"
http://localhost:9990/cron/stop/19
curl http://localhost:9990/cron/delete/19


