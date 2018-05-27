package main

import (
	"app"
	"library/path"
	"controllers/consul"
	"controllers/http"
	"controllers/crontab"
	log "github.com/sirupsen/logrus"
	"controllers/agent"
	"models/cron"
	"encoding/binary"
	"encoding/json"
	"time"
	"controllers/models"
	"database/sql"
	"fmt"
	mlog "models/log"
	"os"
)

func main() {
	//defer func() {
	//	if err := recover(); err != nil {
	//		log.Errorf("%+v", err)
	//	}
	//}()
	// 初始化相关环境
	// 这里传入的参数为配置文件目录
	// 后续增加命令行参数支持，可以指定配置文件目录
	app.Init(path.CurrentPath + "/config")
	defer app.Release()

	ctx := app.NewContext()
	var err error

	// init database
	// 数据库资源
	var handler *sql.DB
	{
		dataSource := fmt.Sprintf(
			"%s:%s@tcp(%s:%d)/%s?charset=%s",
			ctx.Config.MysqlUser,
			ctx.Config.MysqlPassword,
			ctx.Config.MysqlHost,
			ctx.Config.MysqlPort,
			ctx.Config.MysqlDatabase,
			ctx.Config.MysqlCharset,
		)
		handler, err = sql.Open("mysql", dataSource)
		if err != nil {
			log.Panicf("链接数据库错误：%+v", err)
		}
		//设置最大空闲连接数
		handler.SetMaxIdleConns(8)
		//设置最大允许打开的连接
		handler.SetMaxOpenConns(32)
		defer handler.Close()
	}

	// cronController负责对完成定时任务db的增删改查操作
	cronController := models.NewCronController(ctx, handler)
	defer cronController.Close()

	// logController负责记录执行日志
	logController     := models.NewLogController(ctx, handler)

	// 负责定时任务管理，主要是解析定时任务、分发和执行
	crontabController := crontab.NewCrontabController(crontab.SetOnBefore(func(id int64, dispatchServer string, runServer string, output []byte, useTime time.Duration) {
		// 定时任务开始执行
		logController.Add(id, string(output), int64(useTime.Nanoseconds()/1000000), dispatchServer, runServer, int64(time.Now().UnixNano() / 1000000), mlog.Step_2, "定时任务开始执行")
	}), crontab.SetOnAfter(func(id int64, dispatchServer string, runServer string, output []byte, useTime time.Duration) {
		// 定时任务执行完毕
		logController.Add(id, string(output), int64(useTime.Nanoseconds()/1000000), dispatchServer, runServer, int64(time.Now().UnixNano() / 1000000), mlog.Step_3, "定时任务执行完成")
	}))

	// consulControl 实现选leader
	consulControl := consul.NewConsulController(ctx)
	defer consulControl.Close()

	//agentController负责集群内部的通信
	agentController := agent.NewController(ctx, consulControl.GetLeader,  func(event int, data []byte) {
		// client端收到http请求后，转发给server（leader）端
		// leader 端解析后，把变化的的定时任务追加到定时任务列表
		var e cron.CronEntity
		err := json.Unmarshal(data, &e)
		if err != nil {
			log.Errorf("%+v", err)
			return
		}
		crontabController.Add(event, &e)
	}, crontabController.ReceiveCommand, func(cronId int64) {
		//定时任务被发送出去
		logController.Add(cronId, "", 0, "", "", int64(time.Now().UnixNano() / 1000000), mlog.Step_1, "")
	}, func(cronId int64, dispatchServer string) {
		//定时任务发送完成响应回到server端（leader）
		logController.Add(cronId, "", 0, dispatchServer, "", int64(time.Now().UnixNano() / 1000000), mlog.Step_4, "")
	})
	agentController.Start()
	defer agentController.Close()

	crontab.SetOnWillRun(agentController.Dispatch)(crontabController)
	crontab.SetPullCommand(agentController.Pull)(crontabController)

	// 这里设定选leader后的相应操作
	consul.SetOnleader(agentController.OnLeader)(consulControl)
	consul.SetOnleader(func(isLeader bool) {
		fmt.Fprintf(os.Stderr, "==============on leader %v=====================", isLeader)
		if !isLeader {
			crontabController.Stop()
		} else {
			go func() {
				list, err := cronController.GetList()
				if err != nil {
					log.Errorf("%+v", err)
					return
				}
				log.Debugf("==============init crontab list==============")
				for _, e := range list  {
					crontabController.Add(cron.EVENT_ADD, e)
				}
			}()
			crontabController.Start()
		}
	})(consulControl)
	consulControl.Start()

	// httpController负责对应提供http api服务
	httpController := http.NewHttpController(ctx, cronController, logController, http.SetCronHook(func(event int, row *cron.CronEntity) {
		var e = make([]byte, 4)
		binary.LittleEndian.PutUint32(e, uint32(event))
		data, err := json.Marshal(row)
		if err != nil {
			return
		}
		e = append(e, data...)
		agentController.SyncToLeader(e)
	}))
	httpController.Start()
	defer httpController.Close()

	select {
		case <- ctx.Done():
	}
	log.Debug("service exit")
	ctx.Cancel()
}
