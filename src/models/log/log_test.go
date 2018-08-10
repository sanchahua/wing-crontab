package log

import (
	"testing"
	"library/debug"
)

// go test -v -test.run TestDbLog_Add
func TestDbLog_Add(t *testing.T) {
	handler := debug.NewLocalDb()
	db := newDbLog(handler)
	_, err := db.Add(0, "", 0, "")
	if err == nil {
		t.Errorf("Add check cronId fail")
		return
	}
	row, err := db.Add(1, "123", 1000, "hello")
	if err != nil {
		t.Errorf("Add fail, error=[%v]", err)
		return
	}
	if row.Id <= 0 || row.CronId != 1 || row.Output != "123"||
		row.UseTime != 1000 || row.Remark != "hello" {
		t.Errorf("Add check rows fail")
		return
	}
	db.Delete(row.Id)
}

// go test -v -test.run TestDbLog_Delete
func TestDbLog_Delete(t *testing.T) {
	handler := debug.NewLocalDb()
	db := newDbLog(handler)
	row, err := db.Add(1, "123", 1000, "hello")
	if err != nil {
		t.Errorf("Add fail, error=[%v]", err)
		return
	}
	_, err = db.Delete(row.Id)
	if err != nil {
		t.Errorf("Delete fail, error=[%v]", err)
		return
	}
}

// go test -v -test.run TestDbLog_DeleteByCronId
func TestDbLog_DeleteByCronId(t *testing.T) {
	handler := debug.NewLocalDb()
	db := newDbLog(handler)
	row, err := db.Add(1, "123", 1000, "hello")
	if err != nil {
		t.Errorf("Add fail, error=[%v]", err)
		return
	}
	_, err = db.DeleteByCronId(row.CronId)
	if err != nil {
		t.Errorf("DeleteByCronId fail, error=[%v]", err)
		return
	}
}

// go test -v -test.run TestDbLog_Get
func TestDbLog_Get(t *testing.T) {
	handler := debug.NewLocalDb()
	db := newDbLog(handler)
	_, err := db.Add(0, "", 0, "")
	if err == nil {
		t.Errorf("Add check cronId fail")
		return
	}
	row, err := db.Add(1, "123", 1000, "hello")
	if err != nil {
		t.Errorf("Add fail, error=[%v]", err)
		return
	}
	row, err = db.Get(row.Id)
	if err != nil {
		t.Errorf("Get fail, error=[%v]", err)
		return
	}
	if row.Id <= 0 || row.CronId != 1 || row.Output != "123"||
		row.UseTime != 1000 || row.Remark != "hello" {
		t.Errorf("Add check rows fail")
		return
	}
	db.Delete(row.Id)
}

// go test -v -test.run TestDbLog_GetList
func TestDbLog_GetList(t *testing.T) {
	handler := debug.NewLocalDb()
	db := newDbLog(handler)
	_, err := db.Add(0, "", 0, "")
	if err == nil {
		t.Errorf("Add check cronId fail")
		return
	}
	row, err := db.Add(1, "123", 1000, "hello")
	if err != nil {
		t.Errorf("Add fail, error=[%v]", err)
		return
	}
	rows, num, err := db.GetList(row.CronId, 0, 0)
	if err != nil || num <= 0 {
		t.Errorf("Get GetList, error=[%v], num=[%v]", err, num)
		return
	}
	id := row.Id
	row = nil
	for _, r := range rows {
		if r.Id == id {
			row = r
		}
	}
	if row == nil {
		t.Errorf("GetList fail")
		return
	}
	if row.Id <= 0 || row.CronId != 1 || row.Output != "123"||
		row.UseTime != 1000 || row.Remark != "hello" {
		t.Errorf("Add check rows fail")
		return
	}
	db.Delete(row.Id)
}
