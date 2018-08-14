package db

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"sync"

	log "github.com/cihub/seelog"
)

type Database struct {
	Config  *DBConfig // 配置信息
	*sql.DB	   // 数据库连接
}


// NewDatabase 创建数据库对象.  -- 增加最大连接数的设置
func NewDatabase(dbName string, config *DBConfig, maxConn int) (*Database, error) {
	instConfig := config.Instances[dbName]

	database, err := sql.Open(instConfig.Driver, instConfig.Url)
	if err != nil {
		log.Errorf("open database error %v", err)
		return nil, err
	}

	if err := database.Ping(); err != nil {
		log.Errorf("ping database error %v", err)
		return nil, err
	}
	if maxConn > 0 {
		database.SetMaxOpenConns(maxConn)
	}
	return &Database{
		Config: config,
		DB:     database,
	}, nil
}

type Databases struct {
	config    *DBConfig
	instances map[string]*Database
	maxConn   int
	lock      sync.Mutex
}

//获取数据库连接池, map[string]*Database
func GetDatabases(config *DBConfig) (*Databases, error) {
	return &Databases{config: config, instances: make(map[string]*Database), maxConn: 0, lock: sync.Mutex{}}, nil
}

//获取数据库连接池, map[string]*Database -- 增加连接池限制
func GetDatabasesV2(config *DBConfig, maxConn int) (*Databases, error) {
	return &Databases{config: config, instances: make(map[string]*Database), maxConn: maxConn, lock: sync.Mutex{}}, nil
}


// TODO 出错需返回 error
func (ds *Databases) GetDatabase(dbName string) *Database {
	ds.lock.Lock()
	defer ds.lock.Unlock()
	if oldDb, ok := ds.instances[dbName]; !ok {
		database, err := NewDatabase(dbName, ds.config, ds.maxConn)
		if err != nil {
			log.Errorf("init database error %v", err)
			panic(err)
		}
		ds.instances[dbName] = database
		return database
	} else {
		return oldDb
	}
}

