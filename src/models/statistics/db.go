package statistics

import (
	"database/sql"
	"gitlab.xunlei.cn/xllive/common/log"
	"fmt"
	"time"
)

/**
CREATE TABLE `statistics` (
 `id` int(11) NOT NULL AUTO_INCREMENT,
 `cron_id` int(11) NOT NULL DEFAULT '0' COMMENT '定时任务id',
 `day` date NOT NULL COMMENT '日期 如2018-01-01',
 `success` int(11) NOT NULL COMMENT '成功的次数',
 `fail` int(11) NOT NULL COMMENT '失败的次数',
 PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='统计信息'
*/
type Entity struct {
	//Id int64 `json:"id"`
	//CronId int64 `json:"cron_id"`
	Day string `json:"day"`
	Success int64 `json:"success"`
	Fail int64 `json:"fail"`
}

type Statistics struct {
	handler *sql.DB
}

func NewStatistics(handler *sql.DB) *Statistics {
	db := &Statistics{
		handler : handler,
	}
	return db
}

// 查询历史执行次数
func (db *Statistics) GetCount() (int64, error) {
	sqlStr := "SELECT sum(`success` + `fail`) as num FROM `statistics` WHERE 1"
	row := db.handler.QueryRow(sqlStr)
	var (
		num int64
	)
	err := row.Scan(&num)
	if err != nil {
		log.Errorf("GetCount row.Scan fail，sql=[%s]，error=[%+v]", sqlStr, err)
		return 0, err
	}
	return num, nil
}

// 今日执行次数、失败次数
func (db *Statistics) GetDayCount(day string) (int64, int64, error) {
	sqlStr := "SELECT sum(`success` + `fail`) as daynum, sum(fail) as fail FROM `statistics` WHERE `day`=?"
	row := db.handler.QueryRow(sqlStr, day)
	var (
		dayNum int64
		failNum int64
	)
	err := row.Scan(&dayNum, &failNum)
	if err != nil {
		log.Errorf("GetCount row.Scan fail，sql=[%s]，day=[%s], error=[%+v]", sqlStr, day, err)
		return 0, 0, err
	}
	return dayNum, failNum, nil
}

func (db *Statistics) Add(cron_id int64, day string, addSuccessNum, addFailNum int64) error {
	//day := time.Now().Format("2006-01-02")
	// 先查询cron_id, day是否已存在记录
	// 如果有，update
	sqlStr := "select `id` from `statistics` where `day`=? and `cron_id`=?"
	row := db.handler.QueryRow(sqlStr, day, cron_id)
	var (
		id int64
	)
	err := row.Scan(&id)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Errorf("Add row.Scan fail，sql=[%s]，error=[%+v]", sqlStr, err)
		} else {
			//return 0, err
			// 如果没有然后insert
			sqlStr = "INSERT INTO `statistics`(`cron_id`, `day`, `success`, `fail`) VALUES (?,?,?,?)"
			_, err := db.handler.Exec(sqlStr, cron_id, day, addSuccessNum, addFailNum)
			if err != nil {
				log.Errorf("Add db.handler.Exec fail，sql=[%s]，error=[%+v]", sqlStr, err)
				return err
			}
			return nil
		}
	}

	sqlStr = fmt.Sprintf("UPDATE `statistics` SET `success`=(`success`+%d),`fail`=(`fail`+%d) WHERE id=?", addSuccessNum, addFailNum)
	_, err = db.handler.Exec(sqlStr, id)
	if err != nil {
		log.Errorf("Add db.handler.Exec fail，sql=[%s]，error=[%+v]", sqlStr, err)
		return err
	}
	return nil
}

func (db *Statistics) GetAvgTime(day string) (map[int64]int64, error) {
	sqlStr := "SELECT `cron_id`, `avg_use_time` FROM `statistics` WHERE `day`=?"
	rows, err := db.handler.Query(sqlStr, day)
	if err != nil {
		return nil, err
	}
	var (
		avg, cronId int64
	)
	var data = make(map[int64]int64)
	for rows.Next() {
		err := rows.Scan(&cronId, &avg)
		if err != nil {
			continue
		}
		data[cronId] = avg
	}
	return data, nil
}

func (db *Statistics) GetCharts(days int) ([]*Entity, error) {
	if days < 1 {
		days = 7
	}
	t30 := time.Unix(time.Now().Unix()- int64(days) * 86400, 0).Format("2006-01-02")
	sqlStr := "SELECT day, sum(`success`+fail) as success, sum(fail) as fail FROM `statistics` WHERE day>=\""+t30+"\" group by `day`"
	rows, err := db.handler.Query(sqlStr)
	if nil != err {
		log.Errorf("GetCharts fail, error=[%+v]", err)
		return nil, err
	}
	defer rows.Close()
	var records []*Entity
	var (
		day string
		success, fail int64
	)
	for rows.Next() {
		err = rows.Scan(&day, &success, &fail)
		if err != nil {
			log.Errorf("GetList rows.Scan fail，sql=[%s]，error=[%+v]", sqlStr, err)
			continue
		}
		row := &Entity{
			Day:day,
			Success: success,
			Fail: fail,
		}
		records = append(records, row)
	}
	return records, nil
}

func (db *Statistics) SetAvgMAxUseTime(avgUseTime, maxUseTime, cronId int64) error {
	sqlStr := "UPDATE `statistics` SET `avg_use_time`=?,`max_use_time`=? WHERE `cron_id`=? and `day`=?"
	_, err := db.handler.Exec(sqlStr, avgUseTime, maxUseTime, cronId, time.Now().Format("2006-01-02"))
	if err != nil {
		log.Errorf("SetAvgMAxUseTime db.handler.Exec fail，sql=[%s]，error=[%+v]", sqlStr, err)
		return err
	}
	return nil
}