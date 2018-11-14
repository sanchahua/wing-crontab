package debug

import (
	"database/sql"
	"fmt"
	"gitlab.xunlei.cn/xllive/common/log"
	_ "github.com/go-sql-driver/mysql"
	_ "database/sql/driver"
)

func NewLocalDb() *sql.DB {
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

