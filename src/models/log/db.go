package log

import (
	"database/sql"
	log "github.com/sirupsen/logrus"
	"strings"
	"time"
)

const (
	EVENT_CRON_GEGIN    = "1"//cron_start"     //定时任务到点开始执行事件
	EVENT_CRON_DISPATCH = "2"//cron_dispatch"  //定时任务开始分发事件
	EVENT_CRON_RUN      = "3"//cron_run_start" //定时任务开始运行事件
	EVENT_CRON_RUN_END  = "4"//cron_run_end"   //定时任务运行结束事件
	EVENT_CRON_END      = "5"//cron_end"       //定时任务结束
)

type DbLog struct {
	handler *sql.DB
}

func newDbLog(handler *sql.DB) ILog {
	db := &DbLog{
		handler : handler,
	}
	return db
}

// 获取所有的定时任务列表
func (db *DbLog) GetList(cronId int64, search string, dispatchServer, runServer string, page int64, limit int64) ([]*LogEntity, int64, error) {
	sqlStr  := "select `id`, `cron_id`, `time`, `output`, `use_time`, `dispatch_server`, `run_server`  from log where 1 "
	sqlStr2 := "select count(*) as num  from log where 1 "
	var params []interface{}
	var params2 []interface{}
	if cronId > 0 {
		params = append(params, cronId)
		params2 = append(params2, cronId)

		sqlStr  += " and `cron_id`=?"
		sqlStr2 += " and `cron_id`=?"
	}
	search = strings.Trim(search, " ")
	if search != "" {
		params = append(params, "%"+search+"%")
		params2 = append(params2, "%"+search+"%")

		sqlStr  += " and output like ?"
		sqlStr2 += " and output like ?"
	}
	runServer = strings.Trim(runServer, " ")
	if runServer != "" {
		params = append(params, runServer)
		params2 = append(params2, runServer)

		sqlStr  += " and run_server=?"
		sqlStr2 += " and run_server=?"
	}

	dispatchServer = strings.Trim(dispatchServer, " ")
	if dispatchServer != "" {
		params = append(params, dispatchServer)
		params2 = append(params2, dispatchServer)

		sqlStr  += " and dispatch_server=?"
		sqlStr2 += " and dispatch_server=?"
	}

	sqlStr += " order by id desc limit ?,?"


	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 10000 {
		limit = 50
	}
	params = append(params, (page - 1) * limit)
	params = append(params, limit)

	log.Debugf("\n%+v\n%v\n%+v\n%+v", sqlStr, sqlStr2, params, params2)

	stmtOut, err := db.handler.Prepare(sqlStr)
	if err != nil {
		log.Error(err)
		return nil, 0, err
	}
	defer stmtOut.Close()

	rows, err  := stmtOut.Query(params...)
	if nil != err || rows == nil {
		log.Errorf("查询数据库错误：%+v", err)
		return nil, 0, err
	}
	defer rows.Close()
	var records []*LogEntity
	var (
		id int64
		cron_id int64
		Time int64
		output string
		use_time int64
		run_server string
		dispatch_server string
	)
	for rows.Next() {
		//id`, `cron_id`, `time`, `output`, `use_time`, `run_server`
		err = rows.Scan(&id, &cron_id, &Time, &output, &use_time, &dispatch_server, &run_server)
		if err != nil {
			log.Errorf("查询错误，sql=%s，error=%+v", sqlStr, err)
			continue
		}
		row := &LogEntity{
			Id:        id,
			CronId:    cron_id,
			Time:      Time,
			Output:    output,
			UseTime:   use_time,
			RunServer: run_server,
			DispatchServer: dispatch_server,
		}
		log.Infof("%+v", *row)
		records = append(records, row)
	}

	stmtOut2, err := db.handler.Prepare(sqlStr2)
	if err != nil {
		log.Error(err)
		return nil, 0, err
	}
	defer stmtOut2.Close()
	rows2, err := stmtOut2.Query(params2...)
	if nil != err || rows2 == nil {
		log.Errorf("查询数据库错误：%+v", err)
		return nil, 0, err
	}
	defer rows2.Close()

	var num int64 = 0
	for rows2.Next() {
		err = rows2.Scan(&num)
		if err != nil {
			log.Errorf("查询错误，sql=%s，error=%+v", sqlStr2, err)
			continue
		}
		break
	}
	return records, num, nil
}

// 根据指定id查询行
func (db *DbLog) Get(rid int64) (*LogEntity, error) {
	sqlStr := "select `id`, `cron_id`, `time`, `output`, `use_time`, `dispatch_server`, `run_server` from log where id=?"
	data := db.handler.QueryRow(sqlStr, rid)
	var (
		row LogEntity
	)
	err := data.Scan(&row.Id, &row.CronId, &row.Time, &row.Output, &row.UseTime, &row.RunServer)
	if err != nil {
		log.Errorf("查询sql发生错误：%s, %+v", sqlStr, err)
		return &row, err
	}
	return &row, nil
}

func (db *DbLog) Add(cronId int64, output string, useTime int64, dispatchServer, runServer string, rtime int64, event string, remark string) (*LogEntity, error) {
	sqlStr := "INSERT INTO `log`(`cron_id`, `time`, `output`, `use_time`, `dispatch_server`, `run_server`, `event`, `remark`) VALUES (?,?,?,?,?,?,?,?)"
	res, err := db.handler.Exec(sqlStr, cronId, rtime, output, useTime, dispatchServer, runServer, event, remark)
	if err != nil {
		log.Errorf("新增log错误：%+v", err)
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		log.Errorf("新增log错误：%+v", err)
		return nil, err
	}
	return &LogEntity{
		Id:             id,
		CronId:         cronId,
		Time:           time.Now().Unix(),
		Output:         output,
		UseTime:        useTime,
		RunServer:      runServer,
		DispatchServer: dispatchServer,
		Event:          event,
		Remark:         remark,
	}, nil
}


func (db *DbLog) Delete(id int64) (*LogEntity, error) {
	row, err := db.Get(id)
	if err != nil || row == nil {
		log.Errorf("delete error, id does not exists：%v", err)
		return row, err
	}
	sqlStr := "DELETE FROM `log` WHERE id=?"
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

func (db *DbLog) DeleteFormCronId(cronId int64) ([]*LogEntity, error) {
	rows, num, err := db.GetList(cronId, "", "", "", 1, 10000)
	if err != nil || rows == nil {
		log.Errorf("delete error, cronId does not exists：%v", err)
		return rows, err
	}
	sqlStr := "DELETE FROM `log` WHERE cron_id=?"
	log.Debugf("%s", sqlStr)
	res, err := db.handler.Exec(sqlStr, cronId)
	if err != nil {
		log.Errorf("删除定时任务错误：%+v", err)
		return nil, err
	}
	num, err = res.RowsAffected()
	if err != nil || num <= 0{
		log.Errorf("删除定时任务错误：%+v", err)
		return nil, err
	}
	return rows, nil
}
