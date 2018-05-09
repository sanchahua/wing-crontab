package main

import (
	"app"
	"library/path"
	"controllers/consul"
	"controllers/http"
	log "github.com/sirupsen/logrus"
	"controllers/agent"
)

func main() {
	app.Init(path.CurrentPath + "/config")
	defer app.Release()

	ctx := app.NewContext()
	consulControl := consul.NewConsulController(ctx)
	defer consulControl.Close()
	//
	agentController := agent.NewAgentController(ctx, consulControl.GetLeader)
	agentController.Start()
	defer agentController.Close()

	consul.SetOnleader(agentController.OnLeader)(consulControl)
	consulControl.Start()

	httpController := http.NewHttpController(ctx)
	httpController.Start()
	defer httpController.Close()

	select {
		case <- ctx.Done():
	}
	log.Debug("service exit")
	ctx.Cancel()
}
