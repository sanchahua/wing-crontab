package main

import (
	"gitlab.xunlei.cn/xllive/common/log"
	"database/sql"
	"fmt"
	"config"
	"manager"
	"os"
	"os/signal"
	"syscall"
	_ "github.com/go-sql-driver/mysql"
	_ "database/sql/driver"
)

func main() {
	err := config.SeelogInit()
	if err != nil {
		log.Errorf("main config.SeelogInit fail, error=[%v]", err)
		return
	}
	defer log.Flush()

	mysqlConfig, err := config.GetMysqlConfig()
	if err != nil {
		log.Errorf("main config.GetMysqlConfig fail, error=[%v]", err)
		return
	}
	config.WtitePid()
	// init database
	// 数据库资源
	var handler *sql.DB
	{
		dataSource := fmt.Sprintf(
			"%s:%s@tcp(%s:%d)/%s?charset=%s",
			mysqlConfig.User,
			mysqlConfig.Password,
			mysqlConfig.Host,
			mysqlConfig.Port,
			mysqlConfig.Database,
			mysqlConfig.Charset,
		)
		handler, err = sql.Open("mysql", dataSource)
		if err != nil {
			log.Errorf("main sql.Open fail, source=[%v], error=[%+v]", dataSource, err)
			return
		}
		//设置最大空闲连接数
		handler.SetMaxIdleConns(4)
		//设置最大允许打开的连接
		handler.SetMaxOpenConns(4)
		defer handler.Close()
	}
	fmt.Println("start xcrontab")
	m := manager.NewManager(handler)
	m.Start()
	defer m.Stop()
	sc := make(chan os.Signal, 1)
	signal.Notify(sc,
		os.Kill,
		os.Interrupt,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	<-sc
}
