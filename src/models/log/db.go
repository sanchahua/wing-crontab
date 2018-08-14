package log

import (
	"database/sql"
	//log "github.com/cihub/seelog"
	log "gitlab.xunlei.cn/xllive/common/log"
	"strings"
	"fmt"
	"errors"
	ltime "library/time"
)

type DbLog struct {
	handler *sql.DB
}

const (
	FIELDS = "`id`, `cron_id`, `start_time`, `output`, `use_time`, `remark`"
	MaxQueryRows = 10000
)
func newDbLog(handler *sql.DB) *DbLog {
	db := &DbLog{
		handler : handler,
	}
	return db
}

// 查询定时任务执行记录
// 所有的参数都是可选参数
// int类型的值写0表示默认值
// 字符串类型的写为空表示默认值
// 返回值为查询结果集合、总数量、发生的错误
func (db *DbLog) GetList(cronId int64, page int64, limit int64) ([]*LogEntity, int64, error) {
	sqlStr  := "SELECT " + FIELDS + " FROM `log` where 1"
	sqlStr2 := "select count(*) as num  from log where 1 "
	var params  []interface{}
	var params2 []interface{}
	if cronId > 0 {
		params  = append(params, cronId)
		params2 = append(params2, cronId)
		sqlStr  += " and `cron_id`=?"
		sqlStr2 += " and `cron_id`=?"
	}
	sqlStr += " order by id desc limit ?,?"
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > MaxQueryRows {
		limit = 50
	}
	params = append(params, (page - 1) * limit)
	params = append(params, limit)
	debugSql  := fmt.Sprintf(strings.Replace(sqlStr, "?", "%v", -1), params...)
	debugSql2 := fmt.Sprintf(strings.Replace(sqlStr2, "?", "%v", -1), params2...)

	log.Infof("GetList info, sql2=[%v]", debugSql2)

	stmtOut, err := db.handler.Prepare(sqlStr)
	if err != nil {
		log.Errorf("GetList db.handler.Prepare fail, sql=[%v], error=[%v]", debugSql, err)
		return nil, 0, err
	}
	defer stmtOut.Close()
	rows, err  := stmtOut.Query(params...)
	if nil != err {
		log.Errorf("GetList stmtOut.Query fail, sql=[%v], error=[%v]", debugSql, err)
		return nil, 0, err
	}
	defer rows.Close()
	var records []*LogEntity
	var (
		id int64
		cron_id int64
		start_time string
		output string
		use_time int64
		remark string
	)
	for rows.Next() {
		//`id`, `cron_id`, `start_time`, `output`, `use_time`, `remark`
		err = rows.Scan(&id, &cron_id, &start_time, &output, &use_time, &remark)
		if err != nil {
			log.Errorf("GetList rows.Scan fail, sql=[%v], error=[%v]", debugSql, err)
			continue
		}
		row := &LogEntity{
			Id:        id,
			CronId:    cron_id,
			StartTime: start_time,
			Output:    output,
			UseTime:   use_time,
			Remark:    remark,
		}
		records = append(records, row)
	}
	stmtOut2, err := db.handler.Prepare(sqlStr2)
	if err != nil {
		log.Errorf("GetList db.handler.Prepare fail, sql=[%v], error=[%v]", debugSql2, err)
		return nil, 0, err
	}
	defer stmtOut2.Close()
	rows2, err := stmtOut2.Query(params2...)
	if nil != err {
		log.Errorf("GetList stmtOut2.Query fail, sql=[%v], error=[%v]", debugSql2, err)
		return nil, 0, err
	}
	defer rows2.Close()

	var num int64 = 0
	for rows2.Next() {
		err = rows2.Scan(&num)
		if err != nil {
			log.Errorf("GetList rows2.Scan fail, sql=[%v], error=[%v]", debugSql2, err)
			return nil, 0, err
		}
		break
	}
	log.Tracef("GetList success, sql=[%v], sql2=[%v], records=[%+v], num=[%v]", debugSql, debugSql2, records, num)
	return records, num, nil
}

// 根据指定id查询行
func (db *DbLog) Get(rid int64) (*LogEntity, error) {
	if rid <= 0 {
		log.Errorf("Get fail, error=[id invalid]")
		return nil, errors.New("id invalid")
	}
	sqlStr := "select " + FIELDS + " from log where id=?"
	data := db.handler.QueryRow(sqlStr, rid)
	var (
		row LogEntity
	)
	//`id`, `cron_id`, `start_time`, `output`, `use_time`, `remark`
	err := data.Scan(&row.Id, &row.CronId, &row.StartTime, &row.Output, &row.UseTime, &row.Remark)
	if err != nil {
		log.Errorf("Get data.Scan fail, sql=[%v], id=[%v], error=[%v]", sqlStr, rid, err)
		return &row, err
	}
	log.Infof("Get success, sql=[%v], id=[%v], return=[%v]", sqlStr, rid, row)
	return &row, nil
}

func (db *DbLog) Add(cronId int64, output string, useTime int64, remark string) (*LogEntity, error) {
	if cronId <= 0 {
		log.Errorf("Add fail, error=[cron_id invalid], cronId=[%v]", cronId)
		return nil, errors.New("cron_id invalid")
	}
	startTime := ltime.GetDayTime()
	sqlStr := "INSERT INTO `log`(`cron_id`, `start_time`, `output`, `use_time`, `remark`) VALUES (?,?,?,?,?)"
	debugSql := fmt.Sprintf(strings.Replace(sqlStr, "?", "\"%v\"", -1), cronId, startTime, output, useTime, remark)
	res, err := db.handler.Exec(sqlStr, cronId, startTime, output, useTime, remark)
	if err != nil {
		log.Errorf("Add db.handler.Exec fail, sql=[%v], error=[%v]", debugSql, err)
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		log.Errorf("Add res.LastInsertId fail, sql=[%v], error=[%v]", debugSql, err)
		return nil, err
	}
	rsp := &LogEntity{
		Id:         id,
		CronId:     cronId,
		StartTime:  startTime,
		Output:     output,
		UseTime:    useTime,
		Remark:     remark,
	}
	log.Tracef("Add success, sql=[%v], res=[%+v]", debugSql, rsp)
	return rsp, nil
}

func (db *DbLog) Delete(id int64) (*LogEntity, error) {
	if id <= 0 {
		log.Errorf("Delete fail, error=[id invalid]")
		return nil, errors.New("id invalid")
	}
	row, err := db.Get(id)
	if err != nil {
		log.Errorf("Delete db.Get fail, id=[%v], error=[%v]", id, err)
		return row, err
	}
	sqlStr := "DELETE FROM `log` WHERE id=?"
	res, err := db.handler.Exec(sqlStr, id)
	if err != nil {
		log.Errorf("Delete db.handler.Exec fail, sql=[%v], id=[%v], error=[%+v]", sqlStr, id, err)
		return nil, err
	}
	num, err := res.RowsAffected()
	if err != nil {
		log.Errorf("Delete res.RowsAffected fail, sql=[%v], id=[%v], error=[%+v]", sqlStr, id, err)
		return nil, err
	}
	if num <= 0 {
		log.Errorf("Delete res.RowsAffected is 0, sql=[%v], id=[%v], num=[%v], error=[%+v]", sqlStr, id, err)
		return nil, err
	}
	log.Tracef("Delete success, sql=[%v], id=[%v], num=[%v]", sqlStr, id, num)
	return row, nil
}

func (db *DbLog) DeleteByCronId(cronId int64) ([]*LogEntity, error) {
	if cronId <= 0 {
		log.Errorf("Delete fail, cronId=[%v], error=[cronId invalid]", cronId)
		return nil, errors.New("cronId invalid")
	}
	rows, _, err := db.GetList(cronId, 0, MaxQueryRows)
	if err != nil {
		log.Errorf("Delete db.GetList fail, cronId=[%v], error=[%v]", cronId, err)
		return nil, err
	}
	sqlStr := "DELETE FROM `log` WHERE cron_id=?"
	res, err := db.handler.Exec(sqlStr, cronId)
	if err != nil {
		log.Errorf("Delete db.handler.Exec fail, sql=[%v], cronId=[%v], error=[%+v]", sqlStr, cronId, err)
		return nil, err
	}
	num, err := res.RowsAffected()
	if err != nil {
		log.Errorf("Delete res.RowsAffected fail, sql=[%v], cronId=[%v], error=[%+v]", sqlStr, cronId, err)
		return nil, err
	}
	if num <= 0 {
		log.Errorf("Delete res.RowsAffected is 0, sql=[%v], cronId=[%v], error=[%+v]", sqlStr, cronId, err)
		return nil, err
	}
	log.Tracef("Delete success, sql=[%v], cronId=[%v], num=[%v], rows=[%+v]", sqlStr, cronId, num, rows)
	return rows, nil
}
