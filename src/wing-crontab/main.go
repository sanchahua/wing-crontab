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
)

func main() {
	app.Init(path.CurrentPath + "/config")
	defer app.Release()

	ctx := app.NewContext()
	consulControl := consul.NewConsulController(ctx)
	defer consulControl.Close()

	crontabController := crontab.NewCrontabController()
	crontabController.Start()
	defer crontabController.Stop()

	agentController := agent.NewAgentController(ctx, consulControl.GetLeader, func(event int, data *cron.CronEntity) {
		log.Infof("===========%+v", data)
		crontabController.OnCrontabChange(event, data)
	})
	agentController.Start()
	defer agentController.Close()

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
