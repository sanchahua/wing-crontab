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
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("%+v", err)
		}
	}()
	app.Init(path.CurrentPath + "/config")
	defer app.Release()

	ctx := app.NewContext()
	var err error

	// init database
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
		handler.SetMaxOpenConns(8)
		defer handler.Close()
	}

	consulControl := consul.NewConsulController(ctx)
	defer consulControl.Close()

	cronController := models.NewCronController(ctx, handler)
	defer cronController.Close()

	crontabController := crontab.NewCrontabController()

	agentController := agent.NewAgentController(ctx, consulControl.GetLeader, func(event int, data []byte) {
		log.Infof("===========%+v", data)
		var e cron.CronEntity
		err := json.Unmarshal(data, &e)
		if err != nil {
			log.Errorf("%+v", err)
			return
		}
		crontabController.Add(event, &e)
	}, crontabController.RunCommand)
	agentController.Start()
	defer agentController.Close()

	logController := models.NewLogController(ctx, handler)

	crontab.SetOnWillRun(agentController.Dispatch)(crontabController)
	crontab.SetOnRun(func(id int64, dispatchTime int64, dispatchServer string, runServer string, output []byte, useTime time.Duration) {
		log.Infof("run %v in server(%v), use time:%v, output: %+v", id, runServer, useTime, string(output))
		start := time.Now()
		logController.AsyncAdd(id, string(output), int64(useTime.Nanoseconds()/1000000), dispatchTime, dispatchServer, runServer, time.Now().Unix())
		log.Debugf("onrun use time %+v", time.Since(start))
	})(crontabController)
	defer crontabController.Stop()

	consul.SetOnleader(agentController.OnLeader)(consulControl)
	consul.SetOnleader(func(isLeader bool) {
		if !isLeader {
			return
		}
		list, err := cronController.GetList()
		if err != nil {
			log.Errorf("%+v", err)
			return
		}
		log.Debugf("==============init crontab list==============")
		for _, e := range list  {
			crontabController.Add(cron.EVENT_ADD, e)
		}
	})(consulControl)
	consul.SetOnleader(func(isLeader bool) {
		if !isLeader {
			crontabController.Stop()
		} else {
			crontabController.Start()
		}
	})(consulControl)
	consulControl.Start()

	httpController := http.NewHttpController(ctx, cronController, logController, http.SetCronHook(func(event int, row *cron.CronEntity) {
		var e = make([]byte, 4)
		binary.LittleEndian.PutUint32(e, uint32(event))
		data, err := json.Marshal(row)
		if err != nil {
			return
		}
		e = append(e, data...)
		agentController.SendToLeader(e)
	}))
	httpController.Start()
	defer httpController.Close()

	select {
		case <- ctx.Done():
	}
	log.Debug("service exit")
	ctx.Cancel()
}
