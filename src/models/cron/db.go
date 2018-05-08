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
	sqlStr := "select * from cron"
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
		isMutex int
		remark string
		stop int
		lockLimit int64
	)
	for rows.Next() {
		err = rows.Scan(&id, &cronSet, &command, &isMutex, &stop, &remark, &lockLimit)
		if err != nil {
			log.Errorf("查询错误，sql=%s，error=%+v", sqlStr, err)
			continue
		}
		row := &CronEntity{
			Id:id,
			CronSet:cronSet,
			Command:command,
			IsMutex:isMutex == 1,
			Remark:remark,
			Stop:stop == 1,
			LockLimit:lockLimit,
		}
		records = append(records, row)
	}
	return records, nil
}

// 根据指定id查询行
func (db *DbCron) Get(rid int64) (*CronEntity, error) {
	sqlStr := "select * from cron where id=?"
	data := db.handler.QueryRow(sqlStr, rid)
	var (
		row CronEntity
		stop int
		isMutex int
	)
	err := data.Scan(&row.Id, &row.CronSet, &row.Command, &isMutex, &stop, &row.Remark, &row.LockLimit)
	if err != nil {
		log.Errorf("查询sql发生错误：%s, %+v", sqlStr, err)
		return &row, err
	}
	row.IsMutex   = isMutex == 1
	row.Stop      = stop == 1
	return &row, nil
}

func (db *DbCron) Add(cronSet, command string, isMutex bool, remark string, lockLimit int64, stop bool) (*CronEntity, error) {
	iIsMutex := 0
	if isMutex {
		iIsMutex = 1
	}
	iStop := 0
	if stop {
		iStop = 1
	}
	sqlStr := "INSERT INTO `cron`(`cron_set`, `command`, `is_mutex`, `stop`, `remark`, `lock_limit`) " +
		"VALUES (?,?,?,?,?,?)"
	res, err := db.handler.Exec(sqlStr,cronSet, command, iIsMutex, iStop, remark, lockLimit)
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
		IsMutex:isMutex,
		Remark:remark,
		Stop: stop,
		LockLimit:lockLimit,
	}, nil
}

func (db *DbCron) Update(id int64, cronSet, command string, isMutex bool, remark string, lockLimit int64, stop bool) (*CronEntity, error) {
	iIsMutex := 0
	if isMutex {
		iIsMutex = 1
	}
	iStop := 0
	if stop {
		iStop = 1
	}
	sqlStr := "UPDATE `cron` SET `cron_set`=?,`command`=?,`is_mutex`=?,`remark`=?, `stop`=?, `lock_limit`=? WHERE `id`=?"
	res, err := db.handler.Exec(sqlStr, cronSet, command, iIsMutex, remark, iStop, lockLimit, id)
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
		IsMutex:isMutex,
		Remark:remark,
		Stop: stop,
		LockLimit:lockLimit,
	}, nil
}

func (db *DbCron) Stop(id int64) (*CronEntity, error) {
	row, err := db.Get(id)
	if err != nil || row == nil {
		log.Errorf("停止定时任务错误：%v", err)
		return row, err
	}
	return db.Update(id, row.CronSet, row.Command,row.IsMutex, row.Remark, row.LockLimit, true)
}

func (db *DbCron) Start(id int64) (*CronEntity, error) {
	row, err := db.Get(id)
	if err != nil || row == nil {
		log.Errorf("开始定时任务错误：%v", err)
		return row, err
	}
	return db.Update(id, row.CronSet, row.Command,row.IsMutex, row.Remark, row.LockLimit, false)
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
