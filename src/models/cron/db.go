package cron

import (
	"database/sql"
	log "github.com/sirupsen/logrus"
)

type DbCron struct {
	handler *sql.DB
}

func newDbCron(handler *sql.DB) ICron {
	db := &DbCron{
		handler : handler,
	}
	return db
}

// 获取所有的定时任务列表
func (db *DbCron) GetList() ([]*CronEntity, error) {
	sqlStr := "select `id`, `cron_set`, `command`, `stop`, `remark`, `start_time`, `end_time` from cron"
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
	)
	for rows.Next() {
		err = rows.Scan(&id, &cronSet, &command, &stop, &remark, &startTime, &endTime)
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
		}
		records = append(records, row)
	}
	return records, nil
}

// 根据指定id查询行
func (db *DbCron) Get(rid int64) (*CronEntity, error) {
	sqlStr := "select `id`, `cron_set`, `command`, `stop`, `remark`, `start_time`, `end_time` from cron where id=?"
	data := db.handler.QueryRow(sqlStr, rid)
	var (
		row CronEntity
		stop int
	)
	err := data.Scan(&row.Id, &row.CronSet, &row.Command, &stop, &row.Remark, &row.StartTime, &row.EndTime)
	if err != nil {
		log.Errorf("查询sql发生错误：%s, %+v", sqlStr, err)
		return &row, err
	}
	row.Stop      = stop == 1
	return &row, nil
}

func (db *DbCron) Add(cronSet, command string, remark string, stop bool, startTime, endTime int64) (*CronEntity, error) {
	iStop := 0
	if stop {
		iStop = 1
	}
	sqlStr := "INSERT INTO `cron`(`cron_set`, `command`, `stop`, `remark`, `start_time`, `end_time`) VALUES (?,?,?,?,?,?)"
	res, err := db.handler.Exec(sqlStr, cronSet, command, iStop, remark, startTime, endTime)
	if err != nil {
		log.Errorf("新增定时任务错误：%+v", err)
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		log.Errorf("新增定时任务错误：%+v", err)
		return nil, err
	}
	return &CronEntity{
		Id:id,
		CronSet:cronSet,
		Command:command,
		Remark:remark,
		Stop: stop,
		StartTime:startTime,
		EndTime:endTime,
	}, nil
}

func (db *DbCron) Update(id int64, cronSet, command string, remark string, stop bool, startTime, endTime int64) (*CronEntity, error) {
	iStop := 0
	if stop {
		iStop = 1
	}
	sqlStr := "UPDATE `cron` SET `cron_set`=?,`command`=?,`remark`=?, `stop`=?, `start_time`=?, `end_time`=? WHERE `id`=?"
	res, err := db.handler.Exec(sqlStr, cronSet, command, remark, iStop, id, startTime, endTime)
	if err != nil {
		log.Errorf("更新定时任务错误：%+v", err)
		return nil, err
	}
	num, err := res.RowsAffected()
	if err != nil {
		log.Errorf("更新定时任务错误：%+v", err)
		return nil, err
	}
	if num <= 0 {
		return nil, updateFailError
	}
	return &CronEntity{
		Id:id,
		CronSet:cronSet,
		Command:command,
		Remark:remark,
		Stop: stop,
		StartTime:startTime,
		EndTime:endTime,
	}, nil
}

func (db *DbCron) Stop(id int64) (*CronEntity, error) {
	row, err := db.Get(id)
	if err != nil || row == nil {
		log.Errorf("停止定时任务错误：%v", err)
		return row, err
	}
	return db.Update(id, row.CronSet, row.Command, row.Remark, true, row.StartTime, row.EndTime)
}

func (db *DbCron) Start(id int64) (*CronEntity, error) {
	row, err := db.Get(id)
	if err != nil || row == nil {
		log.Errorf("开始定时任务错误：%v", err)
		return row, err
	}
	return db.Update(id, row.CronSet, row.Command, row.Remark, false, row.StartTime, row.EndTime)
}

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
