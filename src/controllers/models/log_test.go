package models

import (
	"testing"
	"library/path"
	"app"
	log "github.com/sirupsen/logrus"
)

func TestNewLogController(t *testing.T) {
	p := path.GetParent(path.WorkingDir)
	p = path.GetParent(p)
	p = path.GetParent(p)
	log.Debugf(p)

	app.Init(p + "/bin/config")
	log.Debugf(p + "/bin/config")

	defer app.Release()
	ctx := app.NewContext()

	con := NewLogController(ctx)
	defer con.Close()
	e, err := con.Add(96, "hello", 100, "123")
	if err != nil || e == nil || e.Id <= 0 {
		t.Errorf("Add error: %+v", err)
	}

	e, err = con.Get(e.Id)
	if err != nil {
		t.Errorf("Add error: %+v", err)
	}

	list, num, err := con.GetList(0, "", "123", 0, 0)
	log.Info(list, num, err)
	if err != nil || list == nil || num != 1 {
		t.Errorf("GetList search run_server error: %+v", err)
	}

	list, num, err = con.GetList(0, "ell", "", 0, 0)
	log.Info(list, num, err)
	if err != nil || list == nil || num != 1 {
		t.Errorf("GetList search error 1: %+v", err)
	}

	e, err = con.Delete(e.Id)
	if err != nil || e == nil || e.Id <= 0 {
		t.Errorf("Delete error: %+v", err)
	}

	// not after does not exists, if err == nil should be error
	e, err = con.Get(e.Id)
	if err == nil {
		t.Errorf("Delete error: %+v", err)
	}

	list, num, err = con.GetList(0, "", "", 0, 0)
	if list != nil || num > 0 {
		t.Errorf("GetList error: %+v", err)
	}

	e, err = con.Add(96, "hello", 100, "123")
	if err != nil || e == nil || e.Id <= 0 {
		t.Errorf("Add error: %+v", err)
	}
	list, err = con.DeleteFormCronId(96)
	if err != nil || list == nil {
		t.Errorf("DeleteFormCronId search error: %+v", err)
	}

	list, num, err = con.GetList(96, "", "", 0, 9999999)
	if list != nil || num > 0 {
		t.Errorf("GetList error: %+v", err)
	}
}
