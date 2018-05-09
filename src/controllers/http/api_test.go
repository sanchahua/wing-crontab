package http

import (
	"testing"
	"app"
	log "github.com/sirupsen/logrus"
	"library/path"
)

func TestNewHttpController(t *testing.T) {
	p := path.GetParent(path.WorkingDir)
	p = path.GetParent(p)
	p = path.GetParent(p)
	log.Debugf(p)

	app.Init(p + "/bin/config")
	log.Debugf(p + "/bin/config")

	defer app.Release()
	ctx := app.NewContext()
	con := NewHttpController(ctx)





	entity, err := con.cron.Add("*/1 * * * * *", "php -v", "", false)
	if err != nil {
		t.Errorf("%+v", err)
	}
	entity, err = con.cron.Get(entity.Id)
	if err != nil {
		t.Errorf("%+v", err)
	}
	entity, err = con.cron.Stop(entity.Id)
	if err != nil {
		t.Errorf("%+v", err)
	}
	if !entity.Stop {
		t.Errorf("%+v", "stop error")
	}
	entity, err = con.cron.Get(entity.Id)
	if err != nil {
		t.Errorf("%+v", err)
	}
	if !entity.Stop {
		t.Errorf("%+v", "stop error 2")
	}
	entity, err = con.cron.Start(entity.Id)
	if err != nil {
		t.Errorf("%+v", err)
	}
	if entity.Stop {
		t.Errorf("%+v", "start error")
	}
	entity, err = con.cron.Get(entity.Id)
	if err != nil {
		t.Errorf("%+v", err)
	}
	if entity.Stop {
		t.Errorf("%+v", "start error 2")
	}
	newRemark := "hello"
	newCronSet := "*/2 * * * * *"
	newCommand := "php -i | grep php.ini"
	entity, err = con.cron.Update(entity.Id, newCronSet, newCommand, newRemark, false)
	if err != nil {
		t.Errorf("%+v", err)
	}
	if entity.Remark != newRemark || entity.CronSet != newCronSet || entity.Command != newCommand {
		t.Errorf("%+v", "update error")
	}
	entity, err = con.cron.Get(entity.Id)
	if err != nil {
		t.Errorf("%+v", err)
	}
	if entity.Remark != newRemark || entity.CronSet != newCronSet || entity.Command != newCommand {
		t.Errorf("%+v", "update error 2")
	}

	list, err := con.cron.GetList()
	if err != nil {
		t.Errorf("%+v", err)
	}

	found := false
	for _, v := range list {
		if v.Id == entity.Id {
			found = true
		}
	}
	if !found {
		t.Errorf("get list error")
	}

	entity, err = con.cron.Delete(entity.Id)
	if err != nil {
		t.Errorf("%+v", err)
	}
	entity, err = con.cron.Get(entity.Id)
	if err == nil || entity.Id > 0 {
		t.Errorf("delete error")
	}

 }
