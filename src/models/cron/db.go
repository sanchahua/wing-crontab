package cron

import (
	"database/sql"
	log "github.com/cihub/seelog"
	"strings"
	"github.com/pkg/errors"
	"fmt"
)

type DbCron struct {
	handler *sql.DB
}

func newDbCron(handler *sql.DB) *DbCron {
	db := &DbCron{
		handler : handler,
	}
	return db
}

// 获取所有的定时任务列表
func (db *DbCron) GetList() ([]*CronEntity, error) {
	sqlStr := "select `id`, `cron_set`, `command`, `stop`, `remark`, `start_time`, `end_time`, `is_mutex` from cron"
	rows, err := db.handler.Query(sqlStr)
	if nil != err || rows == nil {
		log.Errorf("查询数据库错误：%+v", err)
		return nil, err
	}
	defer rows.Close()
	var records []*CronEntity
	var (
		id int64
		cronSet string
		command string
		remark string
		stop int
		startTime int64
		endTime int64
		isMutex int
	)
	for rows.Next() {
		err = rows.Scan(&id, &cronSet, &command, &stop, &remark, &startTime, &endTime, &isMutex)
		if err != nil {
			log.Errorf("查询错误，sql=%s，error=%+v", sqlStr, err)
			continue
		}
		row := &CronEntity{
			Id:id,
			CronSet:cronSet,
			Command:command,
			Remark:remark,
			Stop:stop == 1,
			StartTime:startTime,
			EndTime:endTime,
			IsMutex:isMutex == 1,
		}
		records = append(records, row)
	}
	return records, nil
}

// 根据指定id查询行
func (db *DbCron) Get(rid int64) (*CronEntity, error) {
	if rid <= 0 {
		log.Errorf("Get fail, rid invalid, error=[rid<=0]")
		return nil, errors.New("rid invalid")
	}
	sqlStr := "select `id`, `cron_set`, `command`, `stop`, `remark`, `start_time`, `end_time`, `is_mutex` from cron where id=?"
	data := db.handler.QueryRow(sqlStr, rid)
	var (
		row CronEntity
		stop int
		isMutex int
	)
	err := data.Scan(&row.Id, &row.CronSet, &row.Command, &stop, &row.Remark, &row.StartTime, &row.EndTime, &isMutex)
	if err != nil {
		log.Errorf("Get data.Scan fail, sql=[%s], id=[%v], error=[%+v]", sqlStr, rid, err)
		return nil, err
	}
	row.Stop      = stop == 1
	row.IsMutex   = isMutex == 1
	log.Infof("Get success, sql=[%v], id=[%v]", sqlStr, rid)
	return &row, nil
}

func (db *DbCron) Add(cronSet, command string, remark string, stop bool, startTime, endTime int64, isMutex bool) (*CronEntity, error) {
	cronSet = strings.Trim(cronSet, " ")
	if cronSet == "" {
		log.Errorf("Add [cronSet invalid], cronSet=[%v]", cronSet)
		return nil, errors.New("cronSet is empty")
	}
	command = strings.Trim(command, " ")
	if command == "" {
		log.Errorf("Add [command invalid], cronSet=[%v]", command)
		return nil, errors.New("command is empty")
	}
	if endTime < startTime && (endTime > 0 || startTime > 0) {
		log.Errorf("Add [endTime invalid, endTime=[%v]<startTime=[%v]], endTime=[%v], startTime=[%v]", endTime, startTime, endTime, startTime)
		return nil, errors.New("endTime invalid")
	}
	iStop := 0
	if stop {
		iStop = 1
	}
	iIsMutex := 0
	if isMutex {
		iIsMutex = 1
	}
	sqlStr := "INSERT INTO `cron`(`cron_set`, `command`, `stop`, `remark`, `start_time`, `end_time`, `is_mutex`) VALUES (?,?,?,?,?,?,?)"
	debugSql := fmt.Sprintf(strings.Replace(sqlStr, "?", "\"%v\"", -1), cronSet, command, iStop, remark, startTime, endTime, iIsMutex)
	res, err := db.handler.Exec(sqlStr, cronSet, command, iStop, remark, startTime, endTime, iIsMutex)
	if err != nil {
		log.Errorf("Add db.handler.Exec fail, sql=[%v], error=[%+v]", debugSql, err)
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		log.Errorf("Add res.LastInsertId fail, sql=[%v], error=[%+v]", debugSql, err)
		return nil, err
	}
	log.Infof("Add success, sql=[%v]", debugSql)
	return &CronEntity{
		Id:id,
		CronSet:cronSet,
		Command:command,
		Remark:remark,
		Stop: stop,
		StartTime:startTime,
		EndTime:endTime,
		IsMutex:isMutex,// == 1,
	}, nil
}

func (db *DbCron) Update(id int64, cronSet, command string, remark string, stop bool, startTime, endTime int64, isMutex bool) (*CronEntity, error) {
	if id <= 0 {
		log.Errorf("Update [id invalid], id=[%v]", id)
		return nil, errors.New("id can not be 0")
	}
	cronSet = strings.Trim(cronSet, " ")
	if cronSet == "" {
		log.Errorf("Update [cronSet invalid], cronSet=[%v]", cronSet)
		return nil, errors.New("cronSet is empty")
	}
	command = strings.Trim(command, " ")
	if command == "" {
		log.Errorf("Update [command invalid], cronSet=[%v]", command)
		return nil, errors.New("command is empty")
	}
	if endTime < startTime && (endTime > 0 || startTime > 0) {
		log.Errorf("Update [endTime invalid, endTime=[%v]<startTime=[%v]], endTime=[%v], startTime=[%v]", endTime, startTime, endTime, startTime)
		return nil, errors.New("endTime invalid")
	}
	iStop := 0
	if stop {
		iStop = 1
	}
	iIsMutex := 0
	if isMutex {
		iIsMutex = 1
	}
	sqlStr := "UPDATE `cron` SET `cron_set`=?,`command`=?,`remark`=?, `stop`=?, `start_time`=?, `end_time`=?, `is_mutex`=? WHERE `id`=?"
	debugSql := fmt.Sprintf(strings.Replace(sqlStr, "?", "\"%v\"", -1), cronSet, command, remark, iStop, startTime, endTime, iIsMutex, id)
	res, err := db.handler.Exec(sqlStr, cronSet, command, remark, iStop, startTime, endTime, iIsMutex, id)
	if err != nil {
		log.Errorf("Update db.handler.Exec fail, sql=[%v], error=[%+v]", debugSql, err)
		return nil, err
	}
	num, err := res.RowsAffected()
	if err != nil {
		log.Errorf("Update res.RowsAffected fail, sql=[%v], error=[%+v]", debugSql, err)
		return nil, err
	}
	if num <= 0 {
		log.Errorf("Update fail, sql=[%v], error=[%+v]", debugSql, updateFailError)
		return nil, updateFailError
	}
	log.Infof("Update success, sql=[%v]", debugSql)
	return &CronEntity{
		Id:id,
		CronSet:cronSet,
		Command:command,
		Remark:remark,
		Stop: stop,
		StartTime:startTime,
		EndTime:endTime,
		IsMutex:isMutex,
	}, nil
}

func (db *DbCron) Stop(id int64, stop bool) (*CronEntity, error) {
	if id <= 0 {
		log.Errorf("Stop fail, id invalid, error=[id==0]")
		return nil, errors.New("id invalid")
	}
	row, err := db.Get(id)
	if err != nil {
		log.Errorf("Stop db.Get fail, id=[%v], stop=[%v], error=[%v]", id, stop, err)
		return nil, err
	}
	row, err = db.Update(id, row.CronSet, row.Command, row.Remark, stop, row.StartTime, row.EndTime, row.IsMutex)
	if err != nil {
		log.Errorf("Stop db.Update fail, id=[%v], stop=[%v], error=[%v]", id, stop, err)
		return nil, err
	}
	log.Infof("Stop success, id=[%v], stop=[%v]", id, stop)
	return row, nil
}

//func (db *DbCron) Start(id int64) (*CronEntity, error) {
//	row, err := db.Get(id)
//	if err != nil || row == nil {
//		log.Errorf("开始定时任务错误：%v", err)
//		return row, err
//	}
//	return db.Update(id, row.CronSet, row.Command, row.Remark, false, row.StartTime, row.EndTime, row.IsMutex)
//}

func (db *DbCron) Delete(id int64) (*CronEntity, error) {
	row, err := db.Get(id)
	if err != nil || row == nil {
		log.Errorf("delete error, id does not exists：%v", err)
		return row, err
	}
	sqlStr := "DELETE FROM `cron` WHERE id=?"
	log.Debugf("%s", sqlStr)
	res, err := db.handler.Exec(sqlStr, row.Id)
	if err != nil {
		log.Errorf("删除定时任务错误：%+v", err)
		return nil, err
	}
	num, err := res.RowsAffected()
	if err != nil || num <= 0{
		log.Errorf("删除定时任务错误：%+v", err)
		return nil, err
	}
	return row, nil
}
