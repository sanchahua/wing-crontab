package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	_ "database/sql/driver"
	log "github.com/cihub/seelog"
)

func main() {
	for i := 0; i < 3; i++ {
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
			log.Errorf("main sql.Open fail, source=[%v], error=[%+v]", dataSource, err)
			return
		}
		//设置最大空闲连接数
		handler.SetMaxIdleConns(4)
		//设置最大允许打开的连接
		handler.SetMaxOpenConns(4)
		defer handler.Close()


		var a int64 = 0
		r := handler.QueryRow("select * from log where 1 limit 1")
		r.Scan(&a)
		fmt.Println(a)
		//time.Sleep(time.Second)
	}
}
