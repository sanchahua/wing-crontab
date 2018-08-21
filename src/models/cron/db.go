package cron

import (
	"database/sql"
	//log "github.com/cihub/seelog"
	log "gitlab.xunlei.cn/xllive/common/log"
	"strings"
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
	if nil != err {
		log.Errorf("GetList fail, error=[%+v]", err)
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
			log.Errorf("GetList rows.Scan fail，sql=[%s]，error=[%+v]", sqlStr, err)
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
		log.Errorf("Get fail, error=[%v]", ErrIdInvalid)
		return nil, ErrIdInvalid//errors.New("rid invalid")
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
	row.Stop    = stop == 1
	row.IsMutex = isMutex == 1
	log.Infof("Get success, sql=[%v], id=[%v]", sqlStr, rid)
	return &row, nil
}

func (db *DbCron) Add(cronSet, command string, remark string, stop bool, startTime, endTime int64, isMutex bool) (int64, error) {
	cronSet = strings.Trim(cronSet, " ")
	if cronSet == "" {
		log.Errorf("Add fail, cronSet=[%v], error=[%v]", cronSet, ErrCronSetInvalid)
		return 0, ErrCronSetInvalid//errors.New("cronSet is empty")
	}
	command = strings.Trim(command, " ")
	if command == "" {
		log.Errorf("Add fail, command=[%v], error=[%v]", command, ErrCommandInvalid)
		return 0, ErrCommandInvalid//errors.New("command invalid")
	}
	if endTime < startTime && (endTime > 0 || startTime > 0) {
		log.Errorf("Add fail, [endTime=[%v]<startTime=[%v]], endTime=[%v], startTime=[%v], error=[%v]", endTime, startTime, endTime, startTime, ErrEndTimeInvalid)
		return 0, ErrEndTimeInvalid//errors.New("endTime invalid")
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
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		log.Errorf("Add res.LastInsertId fail, sql=[%v], error=[%+v]", debugSql, err)
		return 0, err
	}
	log.Infof("Add success, sql=[%v]", debugSql)
	return id, nil
}

func (db *DbCron) Update(id int64, cronSet, command string, remark string, stop bool, startTime, endTime int64, isMutex bool) error {
	if id <= 0 {
		log.Errorf("Update fail, id=[%v], error=[%v]", id, ErrIdInvalid)
		return ErrIdInvalid
	}
	cronSet = strings.Trim(cronSet, " ")
	if cronSet == "" {
		log.Errorf("Update fail, cronSet=[%v], error=[%v]", cronSet, ErrCronSetInvalid)
		return ErrCronSetInvalid//nil, errors.New("cronSet is empty")
	}
	command = strings.Trim(command, " ")
	if command == "" {
		log.Errorf("Update fail, command=[%v], error=[%v]", command, ErrCommandInvalid)
		return ErrCommandInvalid//nil, errors.New("command is empty")
	}
	if endTime < startTime && (endTime > 0 || startTime > 0) {
		log.Errorf("Update [endTime=[%v]<startTime=[%v]], endTime=[%v], startTime=[%v], error=[%v]", endTime, startTime, endTime, startTime, ErrEndTimeInvalid)
		return ErrEndTimeInvalid//nil, errors.New("endTime invalid")
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
		return err
	}
	num, err := res.RowsAffected()
	if err != nil {
		log.Errorf("Update res.RowsAffected fail, sql=[%v], error=[%+v]", debugSql, err)
		return err
	}
	if num <= 0 {
		log.Errorf("Update fail, sql=[%v], error=[%+v]", debugSql, ErrNoRowsChange)
		return ErrNoRowsChange
	}
	log.Infof("Update success, sql=[%v]", debugSql)
	return nil
}

// 开始、停止定时任务，取决于第二个参数
// true为停止之意、false为开始的意思
func (db *DbCron) Stop(id int64, stop bool) error {
	if id <= 0 {
		log.Errorf("Stop fail, error=[%v]", ErrIdInvalid)
		return ErrIdInvalid//nil, errors.New("id invalid")
	}
	//row, err := db.Get(id)
	//if err != nil {
	//	log.Errorf("Stop db.Get fail, id=[%v], stop=[%v], error=[%v]", id, stop, err)
	//	return err
	//}
	//err = db.Update(id, row.CronSet, row.Command, row.Remark, stop, row.StartTime, row.EndTime, row.IsMutex)
	//if err != nil {
	//	log.Errorf("Stop db.Update fail, id=[%v], stop=[%v], error=[%v]", id, stop, err)
	//	return nil, err
	//}
	//log.Infof("Stop success, id=[%v], stop=[%v]", id, stop)
	//return row, nil

	iStop := 0
	if stop {
		iStop = 1
	}
	sqlStr := "UPDATE `cron` SET `stop`=? WHERE `id`=?"
	debugSql := fmt.Sprintf(strings.Replace(sqlStr, "?", "\"%v\"", -1), iStop, id)
	res, err := db.handler.Exec(sqlStr, iStop, id)
	if err != nil {
		log.Errorf("Update db.handler.Exec fail, sql=[%v], error=[%+v]", debugSql, err)
		return err
	}
	num, err := res.RowsAffected()
	if err != nil {
		log.Errorf("Update res.RowsAffected fail, sql=[%v], error=[%+v]", debugSql, err)
		return err
	}
	if num <= 0 {
		log.Errorf("Update fail, sql=[%v], error=[%+v]", debugSql, ErrNoRowsChange)
		return ErrNoRowsChange
	}
	return nil
}

func (db *DbCron) Delete(id int64) error {
	if id <= 0 {
		log.Errorf("Delete fail, id invalid, error=[id==0]")
		return ErrIdInvalid//nil, errors.New("id invalid")
	}
	//row, err := db.Get(id)
	//if err != nil {
	//	log.Errorf("Delete db.Get fail, error=[%v]", err)
	//	return row, err
	//}
	sqlStr := "DELETE FROM `cron` WHERE id=?"
	res, err := db.handler.Exec(sqlStr, id)
	if err != nil {
		log.Errorf("Delete db.handler.Exec fail, sql=[%v], id=[%v], error=[%+v]", sqlStr, id, err)
		return err
	}
	num, err := res.RowsAffected()
	if err != nil {
		log.Errorf("Delete res.RowsAffected fail, sql=[%v], id=[%v], error=[%+v]", sqlStr, id, err)
		return err
	}
	if num <= 0 {
		log.Errorf("Delete res.RowsAffected is 0, sql=[%v], id=[%v]", sqlStr, id)
		return ErrNoRowsAffected//nil, err
	}
	log.Infof("Delete success, sql=[%v], id=[%v]", sqlStr, id)
	return nil//row, nil
}
