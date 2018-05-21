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
	logController     := models.NewLogController(ctx, handler)

	agentController := agent.NewController(ctx, consulControl.GetLeader,  func(event int, data []byte) {
		//log.Infof("===========%+v", data)
		var e cron.CronEntity
		err := json.Unmarshal(data, &e)
		if err != nil {
			log.Errorf("%+v", err)
			return
		}
		crontabController.Add(event, &e)
	}, crontabController.ReceiveCommand, logController.AsyncAdd)
	agentController.Start()
	defer agentController.Close()


	crontab.SetOnWillRun(func(id int64, command string, isMutex bool) {
		logEntity, err := logController.Add(id, "", 0, ctx.Config.BindAddress, "", int64(time.Now().UnixNano() / 1000000), mlog.EVENT_CRON_GEGIN, "定时任务开始 - 1")
		if err != nil {
			log.Errorf(" add log with error %v", err)
			return
		}
		agentController.Dispatch(id, command, isMutex, logEntity.Id)
	})(crontabController)
	crontab.SetPullCommand(agentController.Pull)(crontabController)

	crontab.SetOnRun(func(id int64, dispatchTime int64, dispatchServer string, runServer string, output []byte, useTime time.Duration) {
		//log.Infof("run %v in server(%v), use time:%v, output: %+v", id, runServer, useTime, string(output))
		//start := time.Now()
		logController.AsyncAdd(id, string(output), int64(useTime.Nanoseconds()/1000000), dispatchServer, runServer, int64(time.Now().UnixNano() / 1000000), mlog.EVENT_CRON_RUN_END, "定时任务运行结束 - 4")
		//log.Debugf("onrun use time %+v", time.Since(start))
	})(crontabController)
	//crontabController.Start()
	//defer crontabController.Stop()

	consul.SetOnleader(agentController.OnLeader)(consulControl)
	consul.SetOnleader(func(isLeader bool) {
		if !isLeader {
			return
		}
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
