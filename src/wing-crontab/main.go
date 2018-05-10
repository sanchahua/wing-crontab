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
)

func main() {
	app.Init(path.CurrentPath + "/config")
	defer app.Release()

	ctx := app.NewContext()
	consulControl := consul.NewConsulController(ctx)
	defer consulControl.Close()

	crontabController := crontab.NewCrontabController()

	agentController := agent.NewAgentController(ctx, consulControl.GetLeader, func(event int, data []byte) {
		log.Infof("===========%+v", data)
		var e cron.CronEntity
		err := json.Unmarshal(data, &e)
		if err != nil {
			log.Errorf("%+v", err)
			return
		}
		crontabController.OnCrontabChange(event, &e)
	}, crontabController.RunCommand)
	agentController.Start()
	defer agentController.Close()

	logController := models.NewLogController(ctx)
	defer logController.Close()

	crontab.SetOnWillRun(agentController.Dispatch)(crontabController)
	crontab.SetOnRun(func(id int64, runServer string, output []byte, useTime time.Duration) {
		log.Infof("run %v in server(%v), use time:%v, output: %+v", id, runServer, useTime, string(output))
		logController.Add(id, string(output), int64(useTime.Nanoseconds()/1000000), runServer)
	})(crontabController)
	crontabController.Start()
	defer crontabController.Stop()

	consul.SetOnleader(agentController.OnLeader)(consulControl)
	consulControl.Start()

	httpController := http.NewHttpController(ctx, http.SetHook(func(event int, row *cron.CronEntity) {
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
