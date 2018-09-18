package service

import (
	"database/sql"
	"time"
	"gitlab.xunlei.cn/xllive/common/log"
	"os"
	"github.com/go-redis/redis"
	"sync"
)

type Service struct {
	ID            int64
	Address       string
	Updated       int64
	db            *sql.DB
	onRegister    OnRegisterFunc
	onServiceDown func(int64)
	onServiceUp   func(int64)
	onLeader      func(isLeader bool, id int64)
	Status        int       // 1在线 0离线
	Name          string
	Leader        bool
	leaderKey     string
	redis         *redis.Client
	lock          *sync.RWMutex
}

type OnRegisterFunc func(runTimeId int64)

func NewService(
	db *sql.DB,                // 数据库操作资源句柄
	Address string,            // 服务地址， 如 127.0.0.1：38001
	leaderKey string,
	redis *redis.Client,
	onRegister OnRegisterFunc, // 服务注册成功回调
	onServiceDown func(int64), // 服务下线时回调
	onServiceUp func(int64),   // 服务恢复时回调
	onLeader func(isLeader bool, id int64),
) *Service {
	name := "xcrontab"
	n, _ := os.Hostname()
	if "" != n {
		name = n
	}
 	s := &Service{
		db:            db,
		onRegister:    onRegister,
		Address:       Address,
		Name:          name,
		onServiceDown: onServiceDown,
		onServiceUp:   onServiceUp,
		Status:        1,
		leaderKey:     leaderKey,
		redis:         redis,
		Leader:        false,
		onLeader:      onLeader,
		lock:          new(sync.RWMutex),
	}
	// 初始化，主要检查服务是否存在，如果存在会初始化ID
	s.init()
 	s.register()
 	s.selectLeader()
 	go s.tryGetLeader()
 	// 更新updated，此字段用于判断服务是否存活
	go s.keepAlive()
	return s
}

func (s *Service) init() {
	row := s.db.QueryRow("SELECT `id`, `updated` FROM `services` WHERE `address`=?",
		s.Name + "-" + s.Address)
	var id int64
	var updated int64
	err := row.Scan(&id, &updated)
	if err == nil {
		s.ID = id
		s.Updated = updated
	}
}

func (s *Service) selectLeader() {
	v, err := s.redis.Incr(s.leaderKey).Result()
	if err != nil {
		return
	}
	s.Leader = v == 1
	s.onLeader(s.Leader, s.ID)
	s.redis.Expire(s.leaderKey, time.Second * 6)
	if s.Leader {
		s.updateIsLeader()
	}
}

func (s *Service) tryGetLeader()  {
	if s.Leader {
		return
	}
	for {
		v, err := s.redis.Incr(s.leaderKey).Result()
		if err != nil {
			continue
		}
		if v == 1 {
			s.lock.Lock()
			s.Leader = true
			s.lock.Unlock()
			s.onLeader(s.Leader, s.ID)
			s.updateIsLeader()
		}
		time.Sleep(time.Second * 6)
	}
}

// 服务注册
func (s *Service) register() (int64, error) {
	defer s.onRegister(s.ID)
	if s.ID > 0 {
		return s.ID, nil
	}
	res, err := s.db.Exec("INSERT INTO `services`(`address`, `updated`) VALUES (?,?)",
		s.Name + "-" + s.Address, time.Now().Unix())
	if err != nil {
		log.Errorf("Register s.db.Exec fail, error=[%v]", err)
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		log.Errorf("Register res.LastInsertId fail, error=[%v]", err)
		return 0, err
	}
	s.ID = id
	s.Updated = time.Now().Unix()
	return id, nil
}

func (s *Service) keepAlive() (error) {
	for {
		if s.ID <= 0 {
			time.Sleep(time.Second * 1)
			continue
		}

		s.lock.RLock()
		if s.Leader {
			s.redis.Expire(s.leaderKey, time.Second * 6)
		}
		s.lock.RUnlock()

		t := time.Now().Unix()
		res, err := s.db.Exec("UPDATE `services` SET `updated`=? WHERE id=?", t, s.ID)
		if err != nil {
			log.Errorf("keepAlive s.db.Exec fail, error=[%v]", err)
			time.Sleep(time.Second * 1)
			continue
		}
		_, err = res.RowsAffected()
		if err != nil {
			log.Errorf("keepAlive res.RowsAffected fail, error=[%v]", err)
			time.Sleep(time.Second * 1)
			continue
		}
		s.Updated = t
		time.Sleep(time.Second * 1)
	}
}

func (s *Service) updateIsLeader() {
	sqlStr := "UPDATE `services` SET `is_leader`=0 WHERE id!=?"
	s.db.Exec(sqlStr, s.ID)
	sqlStr = "UPDATE `services` SET `is_leader`=1 WHERE id=?"
	s.db.Exec(sqlStr, s.ID)
}

// 服务注销
func (s *Service) Deregister() (error) {
	res, err := s.db.Exec("DELETE FROM `services` WHERE id=?", s.ID)
	if err != nil {
		log.Errorf("Deregister s.db.Exec fail, error=[%v]", err)
		return err
	}
	_, err = res.RowsAffected()
	if err != nil {
		log.Errorf("Deregister res.RowsAffected fail, error=[%v]", err)
		return err
	}
	return nil
}

func (s *Service) GetServices() ([]*Service, error) {
	services := make([]*Service, 0)
	rows, err := s.db.Query("SELECT `id`, `address`, `updated` FROM `services` WHERE 1")
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		sr := new(Service)
		err = rows.Scan(&sr.ID, &sr.Address, &sr.Updated)
		if err != nil {
			continue
		}
		sr.Status = 1
		if time.Now().Unix() - sr.Updated >= 6 {
			sr.Status = 0
		}
		services = append(services, sr)
	}
	return services, nil
}