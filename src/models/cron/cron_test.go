package cron

import (
	"testing"
	"fmt"
	"database/sql"
	log "github.com/cihub/seelog"
	_ "github.com/go-sql-driver/mysql"
	_ "database/sql/driver"
)

func newLocalDb() *sql.DB {
	dataSource := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=%s",
		"root",
		"123456",
		"127.0.0.1",
		3306,
		"cron",
		"utf8",
	)
	handler, err := sql.Open("mysql", dataSource)
	if err != nil {
		log.Errorf("newLocalDb sql.Open fail, source=[%v], error=[%+v]", dataSource, err)
		return nil
	}
	//设置最大空闲连接数
	handler.SetMaxIdleConns(4)
	//设置最大允许打开的连接
	handler.SetMaxOpenConns(4)
	return handler
}

// go test -v -test.run TestDbCron_Add
func TestDbCron_Add(t *testing.T) {
	handler := newLocalDb()
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
	c, err := db.Add("   */1 * * * * *   ", "curl http://www.baidu.com/ ", "", false, 0, 0, false)
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

// go test -v -test.run TestDbCron_Get
func TestDbCron_Get(t *testing.T) {
	handler := newLocalDb()
	db := NewCron(handler)
	c, err := db.Add("   */1 * * * * *   ", "curl http://www.baidu.com/ ", "", false, 0, 0, false)
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
	handler := newLocalDb()
	db := NewCron(handler)
	c, err := db.Add("   */1 * * * * *   ", "curl http://www.baidu.com/ ", "", false, 0, 0, false)
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

	c, err = db.Update(c.Id, "   */2 * * * * *   ", "curl http://www.baidu.com/2 ", "2", true, 1, 2, true)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	if c.Id <= 0 {
		t.Errorf("%v", "check id fail")
		return
	}
	if c.CronSet != "*/2 * * * * *" {
		t.Errorf("%v", "check CronSet fail")
		return
	}
	if c.Command != "curl http://www.baidu.com/2" {
		t.Errorf("%v", "check Command fail")
		return
	}
	if !c.Stop {
		t.Errorf("%v", "check Stop fail")
		return
	}
	if c.Remark != "2" {
		t.Errorf("%v", "check Remark fail")
		return
	}
	if c.StartTime != 1 {
		t.Errorf("%v", "check StartTime fail")
		return
	}
	if c.EndTime != 2 {
		t.Errorf("%v", "check EndTime fail")
		return
	}
	if !c.IsMutex {
		t.Errorf("%v", "check IsMutex fail")
		return
	}
	db.Delete(c.Id)
}
