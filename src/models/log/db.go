package log

import (
	"database/sql"
	"gitlab.xunlei.cn/xllive/common/log"
	"strings"
	"fmt"
	"errors"
)

type DbLog struct {
	handler *sql.DB
}

const (
	FIELDS = "`id`, `dispatch_server`, `run_server`, `cron_id`, `process_id`, `state`, `start_time`, `output`, `use_time`, `remark`"
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
func (db *DbLog) GetList(cronId, useTime int64, searchFail bool, page int64, limit int64, searchResult bool, startTime, endTime, searchOutput, sort string) ([]*LogEntity, int64, int64, int64, error) {
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

	if useTime > 0 {
		params  = append(params, useTime)
		params2 = append(params2, useTime)
		sqlStr  += " and `use_time`>=?"
		sqlStr2 += " and `use_time`>=?"
	}

	if searchFail {
		sqlStr  += " and `state`=\"fail\" "
		sqlStr2 += " and `state`=\"fail\" "
	}
	searchOutput = strings.Trim(searchOutput, " ")
	if searchOutput != "" {
		sqlStr  += " and `output` like ?"
		sqlStr2 += " and `output` like ?"
		params  = append(params, "%"+searchOutput+"%")
		params2 = append(params2, "%"+searchOutput+"%")
	}

	if searchResult {
		sqlStr += " and use_time>0"
		sqlStr2 += " and use_time>0"
	}

	if startTime != "" {
		sqlStr += " and start_time>=?"
		sqlStr2 += " and start_time>=?"
		params  = append(params, startTime)
		params2 = append(params2, startTime)
	}

	if endTime != "" {
		sqlStr += " and start_time<=?"
		sqlStr2 += " and start_time<=?"
		params  = append(params, endTime)
		params2 = append(params2, endTime)
	}

	sqlStr += " order by "+sort+" limit ?,?"
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

	//log.Infof("GetList info, sql2=[%v]", debugSql2)

	//stmtOut, err := db.handler.Prepare(sqlStr)
	//if err != nil {
	//	log.Errorf("GetList db.handler.Prepare fail, sql=[%v], error=[%v]", debugSql, err)
	//	return nil, 0, page, limit, err
	//}
	//defer stmtOut.Close()
	rows, err  := db.handler.Query(sqlStr, params...)
	if nil != err {
		log.Errorf("GetList db.handler.Query fail, sql=[%v], error=[%v]", debugSql, err)
		return nil, 0, page, limit, err
	}
	defer rows.Close()
	var records []*LogEntity
	var (
		id int64
		dispatch_server, run_server int64
		cron_id int64
		process_id int
		start_time string
		output string
		use_time int64
		remark string
		state string
	)
	for rows.Next() {
		//`id`, `cron_id`, `start_time`, `output`, `use_time`, `remark`
		err = rows.Scan(&id, &dispatch_server, &run_server, &cron_id, &process_id, &state, &start_time, &output, &use_time, &remark)
		if err != nil {
			log.Errorf("GetList rows.Scan fail, sql=[%v], error=[%v]", debugSql, err)
			continue
		}
		row := &LogEntity{
			Id:             id,
			DispatchServer: dispatch_server,
			RunServer:      run_server,
			CronId:         cron_id,
			ProcessId:      process_id,
			StartTime:      start_time,
			Output:         output,
			UseTime:        use_time,
			Remark:         remark,
			State:          state,
		}
		records = append(records, row)
	}
	//stmtOut2, err := db.handler.Prepare(sqlStr2)
	//if err != nil {
	//	log.Errorf("GetList db.handler.Prepare fail, sql=[%v], error=[%v]", debugSql2, err)
	//	return nil, 0, page, limit, err
	//}
	//defer stmtOut2.Close()
	rows2 := db.handler.QueryRow(sqlStr2, params2...)
	//if nil != err {
	//	log.Errorf("GetList stmtOut2.Query fail, sql=[%v], error=[%v]", debugSql2, err)
	//	return nil, 0, page, limit, err
	//}
	//defer rows2.Close()

	var num int64 = 0
	//for rows2.Next() {
		err = rows2.Scan(&num)
		if err != nil {
			log.Errorf("GetList rows2.Scan fail, sql=[%v], error=[%v]", debugSql2, err)
			return nil, 0, page, limit, err
		}
	//	break
	//}
	//log.Tracef("GetList success, sql=[%v], sql2=[%v], records=[%+v], num=[%v]", debugSql, debugSql2, records, num)
	return records, num, page, limit, nil
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
	//FIELDS = "`id`, `cron_id`, `process_id`, `state`, `start_time`, `output`, `use_time`, `remark`"
	err := data.Scan(&row.Id, &row.DispatchServer, &row.RunServer, &row.CronId, &row.ProcessId, &row.State, &row.StartTime, &row.Output, &row.UseTime, &row.Remark)
	if err != nil {
		log.Errorf("Get data.Scan fail, sql=[%v], id=[%v], error=[%v]", sqlStr, rid, err)
		return &row, err
	}
	log.Infof("Get success, sql=[%v], id=[%v], return=[%v]", sqlStr, rid, row)
	return &row, nil
}

func (db *DbLog) Add(dispatchServer, runServer int64, cronId int64, processId int, state string, output string, useTime int64, remark, startTime string) (int64, error) {
	if cronId <= 0 {
		log.Errorf("Add fail, error=[cron_id invalid], cronId=[%v]", cronId)
		return 0, errors.New("cron_id invalid")
	}
	sqlStr := "INSERT INTO `log`(`dispatch_server`, `run_server`,`cron_id`, `process_id`, `state`, `start_time`, `output`, `use_time`, `remark`) VALUES (?,?,?,?,?,?,?,?,?)"
	debugSql := fmt.Sprintf(strings.Replace(sqlStr, "?", "\"%v\"", -1), dispatchServer, runServer, cronId, processId, state, startTime, output, useTime, remark)
	res, err := db.handler.Exec(sqlStr, dispatchServer, runServer, cronId, processId, state, startTime, output, useTime, remark)
	if err != nil {
		log.Errorf("Add db.handler.Exec fail, sql=[%v], error=[%v]", debugSql, err)
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		log.Errorf("Add res.LastInsertId fail, sql=[%v], error=[%v]", debugSql, err)
		return 0, err
	}
	//log.Tracef("Add success, sql=[%v], id=[%+v]", debugSql, id)
	return id, nil
}

func (db *DbLog) Delete(id int64) error {
	if id <= 0 {
		log.Errorf("Delete fail, error=[id invalid]")
		return ErrIdInvalid
	}
	sqlStr := "DELETE FROM `log` WHERE id=?"
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
		log.Errorf("Delete res.RowsAffected is 0, sql=[%v], id=[%v], num=[%v], error=[%+v]", sqlStr, id, err)
		return ErrNoRowsAffected
	}
	log.Tracef("Delete success, sql=[%v], id=[%v], num=[%v]", sqlStr, id, num)
	return nil
}

func (db *DbLog) DeleteByCronId(cronId int64) error {
	if cronId <= 0 {
		log.Errorf("DeleteByCronId fail, cronId=[%v], error=[cronId invalid]", cronId)
		return ErrIdInvalid
	}
	sqlStr := "DELETE FROM `log` WHERE cron_id=?"
	res, err := db.handler.Exec(sqlStr, cronId)
	if err != nil {
		log.Errorf("DeleteByCronId db.handler.Exec fail, sql=[%v], cronId=[%v], error=[%+v]", sqlStr, cronId, err)
		return err
	}
	num, err := res.RowsAffected()
	if err != nil {
		log.Errorf("DeleteByCronId res.RowsAffected fail, sql=[%v], cronId=[%v], error=[%+v]", sqlStr, cronId, err)
		return  err
	}
	if num <= 0 {
		log.Errorf("DeleteByCronId res.RowsAffected is 0, sql=[%v], cronId=[%v], error=[%+v]", sqlStr, cronId, err)
		return ErrNoRowsAffected
	}
	log.Tracef("DeleteByCronId success, sql=[%v], cronId=[%v], num=[%v]", sqlStr, cronId, num)
	return nil
}

func (db *DbLog) DeleteByStartTime(startTime string) error {
	//log.Tracef("DeleteByStartTime start: %v", startTime)
	if startTime == "" {
		return ErrorStartTimeEmpty
	}
	sqlStr := "DELETE FROM `log` WHERE `start_time`<=?"
	res, err := db.handler.Exec(sqlStr, startTime)
	if err != nil {
		log.Errorf("DeleteByStartTime db.handler.Exec fail, sql=[%v], startTime=[%v], error=[%+v]", sqlStr, startTime, err)
		return err
	}
	_, err = res.RowsAffected()
	if err != nil {
		log.Errorf("DeleteByStartTime res.RowsAffected fail, sql=[%v], startTime=[%v], error=[%+v]", sqlStr, startTime, err)
		return  err
	}
	//if num <= 0 {
		//log.Errorf("DeleteByStartTime res.RowsAffected is 0, sql=[%v], startTime=[%v], error=[%+v]", sqlStr, startTime, err)
		//return nil//ErrNoRowsAffected
	//}
	//log.Tracef("DeleteByStartTime success, sql=[%v], startTime=[%v], num=[%v]", sqlStr, startTime, num)
	return nil
}

// 获取平均运行时长
func (db *DbLog) GetAvgRunTime() (map[int64]int64, error) {
	sqlStr := "SELECT `cron_id`, sum(`use_time`)/count(*) as avg_time FROM `log` WHERE `state`=\"success\" group by `cron_id`"
	rows, err := db.handler.Query(sqlStr)
	if err != nil {
		log.Errorf("GetAvgRunTime db.handler.Query fail, error=[%v]", err)
		return nil, err
	}
	var data = make(map[int64]int64)
	for rows.Next() {
		var id int64
		var useTime float64
		err = rows.Scan(&id, &useTime)
		if err != nil {
			log.Errorf("GetAvgRunTime rows.Scan fail, error=[%v]", err)
			continue
		}
		data[id] = int64(useTime)
	}
	return data, nil
}

// 获取指定定时任务的最大运行时间
func (db *DbLog) GetMaxRunTime(cronId int64) (int64, error) {
	sqlStr := "SELECT ifnull(max(`use_time`),0) as use_time FROM `log` WHERE  `cron_id`=?"
	row := db.handler.QueryRow(sqlStr, cronId)
	var useTime int64//sql.NullInt64
	err := row.Scan(&useTime)
	if err != nil {
		log.Errorf("GetMaxRunTime rows.Scan fail, cronId=[%v], error=[%v]", cronId, err)
		return 0, err
	}
	return useTime, nil
}
