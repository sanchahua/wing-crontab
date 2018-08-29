package config

import (
	"github.com/BurntSushi/toml"
	"library/path"
	"fmt"
	"library/file"
	_ "net/http/pprof"
	//log "github.com/cihub/seelog"
	log "gitlab.xunlei.cn/xllive/common/log"
	"errors"
	"os"
	"io/ioutil"
)

// 配置文件管理相关实现
// 这里主要使用到了mysql的配置
type MysqlConfig struct {
	User string       `toml:"user"`
	Password string   `toml:"password"`
	Host string       `toml:"host"`
	Port int          `toml:"port"`
	Database string   `toml:"database"`
	Charset string    `toml:"charset"`
}

// 读取mysql配置
func GetMysqlConfig() (*MysqlConfig, error) {
	var appConfig MysqlConfig
	//configFile := path.CurrentPath + "/config/mysql.toml"
	configFile := "/Users/yuyi/Code/go/xcrontab/bin/config/mysql.toml"
	if !file.Exists(configFile) {
		log.Errorf("GetMysqlConfig config file not found, file=[%v]", configFile)
		return nil, errors.New(fmt.Sprintf("config file not found, file=[%v]", configFile))
	}
	if _, err := toml.DecodeFile(configFile, &appConfig); err != nil {
		log.Errorf("GetMysqlConfig toml.DecodeFile fail, file=[%v], error=[%v]", configFile, err)
		return nil, err
	}
	log.Infof("GetMysqlConfig [%+v]", appConfig)
	return &appConfig, nil
}

// 初始化seelog日志组件
func SeelogInit() error {
	// 初始化日志组件
	//logger, err := log.LoggerFromConfigAsFile(path.CurrentPath + "/config/logger.xml")
	//if err != nil {
	//	log.Errorf("SeelogInit fail, config file=[%v], error=[%v]", path.CurrentPath + "/logger.xml", err)
	//	return err
	//}
	//log.ReplaceLogger(logger)
	configFile:= "/Users/yuyi/Code/go/xcrontab/bin/config/logger.yml"
	//configFile := path.CurrentPath + "/config/logger.yml"
	ilog, err := log.NewLogMgr(configFile)
	if err != nil {
		fmt.Println("log init error:", err)
		return err
	}
	log.SetDefaultLogMgr(ilog)
	return nil
}

func WtitePid() {
	data := []byte(fmt.Sprintf("%d", os.Getpid()))
	ioutil.WriteFile(path.CurrentPath + "/xcrontab.pid", data, 0644)
}
