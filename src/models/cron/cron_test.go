package cron

import (
	"testing"
	"library/debug"
)

// go test -v -test.run TestDbCron_Add
func TestDbCron_Add(t *testing.T) {
	handler := debug.NewLocalDb()
	defer handler.Close()
	db := NewCron(handler)
	if db == nil {
		t.Errorf("open db connect error")
		return
	}
	_, err := db.Add(" ", "curl http://www.baidu.com/", "", false, 0, 0, false)
	if err == nil {
		t.Errorf("%v", "db.Add check cronSet fail")
		return
	}
	_, err = db.Add("", "curl http://www.baidu.com/", "", false, 0, 0, false)
	if err == nil {
		t.Errorf("%v", "db.Add check cronSet fail")
		return
	}
	_, err = db.Add("*/1 * * * * *", "", "", false, 0, 0, false)
	if err == nil {
		t.Errorf("%v", "db.Add check command fail")
		return
	}
	_, err = db.Add("*/1 * * * * *", "curl http://www.baidu.com/", "", false, 1, 0, false)
	if err == nil {
		t.Errorf("%v", "db.Add check startTime/endTime fail")
		return
	}
	id, err := db.Add("   */1 * * * * *   ", "curl http://www.baidu.com/ ", "", false, 0, 0, false)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	if id <= 0 {
		t.Errorf("%v", "check id fail")
		return
	}

	db.Delete(id)
}

// go test -v -test.run TestDbCron_Get
func TestDbCron_Get(t *testing.T) {
	handler := debug.NewLocalDb()
	defer handler.Close()
	db := NewCron(handler)
	id, err := db.Add("   */1 * * * * *   ", "curl http://www.baidu.com/ ", "", false, 0, 0, false)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	if id <= 0 {
		t.Errorf("Add fail")
		return
	}
	c, err := db.Get(id)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	if c.Id <= 0 {
		t.Errorf("%v", "check id fail")
		return
	}
	if c.CronSet != "*/1 * * * * *" {
		t.Errorf("%v", "check CronSet fail")
		return
	}
	if c.Command != "curl http://www.baidu.com/" {
		t.Errorf("%v", "check Command fail")
		return
	}
	if c.Stop {
		t.Errorf("%v", "check Stop fail")
		return
	}
	if c.Remark != "" {
		t.Errorf("%v", "check Remark fail")
		return
	}
	if c.StartTime != 0 {
		t.Errorf("%v", "check StartTime fail")
		return
	}
	if c.EndTime != 0 {
		t.Errorf("%v", "check EndTime fail")
		return
	}
	if c.IsMutex {
		t.Errorf("%v", "check IsMutex fail")
		return
	}
	db.Delete(c.Id)
}

// go test -v -test.run TestDbCron_Update
func TestDbCron_Update(t *testing.T) {
	handler := debug.NewLocalDb()
	defer handler.Close()
	db := NewCron(handler)
	id, err := db.Add("   */1 * * * * *   ", "curl http://www.baidu.com/ ", "", false, 0, 0, false)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	if id <= 0 {
		t.Errorf("%v", "check id fail")
		return
	}

	err = db.Update(id, "   */2 * * * * *   ", "curl http://www.baidu.com/2 ", "2", true, 1, 2, true)
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	db.Delete(id)
}

// go test -v -test.run TestDbCron_Stop
func TestDbCron_Stop(t *testing.T) {
	handler := debug.NewLocalDb()
	defer handler.Close()
	db := NewCron(handler)
	c, err := db.Add("   */1 * * * * *   ", "curl http://www.baidu.com/ ", "", false, 0, 0, false)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	err = db.Stop(c, true)
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	err = db.Stop(c, false)
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	db.Delete(c)
}

// go test -v -test.run TestDbCron_Delete
func TestDbCron_Delete(t *testing.T) {
	handler := debug.NewLocalDb()
	defer handler.Close()
	db := NewCron(handler)
	id, err := db.Add("   */1 * * * * *   ", "curl http://www.baidu.com/ ", "", false, 0, 0, false)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	if id <= 0 {
		t.Errorf("add fail")
		return
	}
	c, err := db.Get(id)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	err = db.Delete(c.Id)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	c, err = db.Get(c.Id)
	if err == nil {
		t.Errorf("%v", "get fail")
		return
	}
}

// go test -v -test.run TestDbCron_GetList
func TestDbCron_GetList(t *testing.T) {
	handler := debug.NewLocalDb()
	defer handler.Close()
	db := NewCron(handler)
	c, err := db.Add("   */1 * * * * *   ", "curl http://www.baidu.com/ ", "", false, 0, 0, false)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	rows, err := db.GetList()
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	if rows == nil {
		t.Errorf("GetList fail")
		return
	}
	err = db.Delete(c)
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	rows, err = db.GetList()
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	found := false
	for _, r := range rows {
		if r.Id == c {
			found = true
		}
	}
	if found {
		t.Errorf("GetList fail")
	}
}
