package config

import (
	"github.com/BurntSushi/toml"
	"library/path"
	"fmt"
	"library/file"
	_ "net/http/pprof"
	log "github.com/cihub/seelog"
	"errors"
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
	configFile := path.CurrentPath + "/config/canal.toml"
	if !file.Exists(configFile) {
		log.Errorf("GetMysqlConfig config file not found, file=[%v]", configFile)
		return nil, errors.New(fmt.Sprintf("config file not found, file=[%v]", configFile))
	}
	if _, err := toml.DecodeFile(configFile, &appConfig); err != nil {
		log.Errorf("GetMysqlConfig toml.DecodeFile fail, file=[%v], error=[%v]", configFile, err)
		return nil, err
	}
	return &appConfig, nil
}

// 初始化seelog日志组件
func SeelogInit() error {
	// 初始化日志组件
	logger, err := log.LoggerFromConfigAsFile(path.CurrentPath + "/logger.xml")
	if err != nil {
		log.Errorf("SeelogInit fail, config file=[%v], error=[%v]", path.CurrentPath + "/logger.xml", err)
		return err
	}
	log.ReplaceLogger(logger)
	return nil
}
