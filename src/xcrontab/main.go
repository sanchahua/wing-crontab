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
	"flag"
	service2 "service"
	_ "net/http/pprof"
	"net/http"
	"github.com/go-redis/redis"
)

func main() {
	// -l "0.0.0.0:38001"
	// -h
	listen := flag.String("l", "0.0.0.0:38001", "restful http server listen")
	pl     := flag.String("pl", "127.0.0.1:7772", "restful http server listen")
	help   := flag.Bool("h", false, "show help info")
	flag.Parse()

	go func() {
		//http://localhost:8880/debug/pprof/  内存性能分析工具
		//go tool pprof logDemo.exe --text a.prof
		//go tool pprof your-executable-name profile-filename
		//go tool pprof your-executable-name http://localhost:8880/debug/pprof/heap
		//go tool pprof wing-binlog-go http://localhost:8880/debug/pprof/heap
		//https://lrita.github.io/2017/05/26/golang-memory-pprof/
		//然后执行 text
		//go tool pprof -alloc_space http://127.0.0.1:8880/debug/pprof/heap
		//top20 -cum
		//下载文件 http://localhost:8880/debug/pprof/profile
		//分析 go tool pprof -web /Users/yuyi/Downloads/profile
		http.ListenAndServe(*pl, nil)
	}()

	if *help {
		fmt.Fprintf(os.Stderr, "./xcrontab -l 0.0.0.0:38001  [restful http server listen]\r\n")
		return
	}

	err := config.SeelogInit()
	if err != nil {
		log.Errorf("main config.SeelogInit fail, error=[%v]", err)
		return
	}
	defer log.Flush()

	appConfig, err := config.GetAppConfig()
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
			appConfig.User,
			appConfig.Password,
			appConfig.Host,
			appConfig.Port,
			appConfig.Database,
			appConfig.Charset,
		)
		handler, err = sql.Open("mysql", dataSource)
		if err != nil {
			log.Errorf("main sql.Open fail, source=[%v], error=[%+v]", dataSource, err)
			return
		}
		//设置最大空闲连接数
		handler.SetMaxIdleConns(8)
		//设置最大允许打开的连接
		handler.SetMaxOpenConns(32)
		defer handler.Close()
	}
	fmt.Println("start xcrontab")

	redisClient := redis.NewClient(&redis.Options{
		Addr:     appConfig.RedisAddress,
		Password: appConfig.RedisPassword, // no password set
		DB:       0,  // use default DB
	})

	_, err = redisClient.Ping().Result()
	if err != nil {
		log.Errorf("%v", err)
		panic(0);
	}

	service2.NewService(handler, *listen, appConfig.LeaderKey, redisClient, func(runTimeId int64) {

	}, func(i int64) {
		log.Warnf("#####%v is down#####", i)
	}, func(i int64) {
		log.Infof("#####%v is up#####", i)
	}, func(id int64) {
		
	})
	m := manager.NewManager(handler, *listen, appConfig.LogKeepDay)
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
