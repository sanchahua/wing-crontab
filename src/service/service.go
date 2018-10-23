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

// new service
func NewService(
	db *sql.DB,                // 数据库操作资源句柄
	Address string,            // 服务地址， 如 127.0.0.1：38001
	leaderKey string,
	redis *redis.Client,
	onRegister OnRegisterFunc, // 服务注册成功回调
	onServiceDown func(int64), // 服务下线时回调
	onServiceUp func(int64),   // 服务恢复时回调
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
		onLeader:      nil,
		lock:          new(sync.RWMutex),
	}
	// 初始化，主要检查服务是否存在，如果存在会初始化ID
	s.init()
	s.register()
	return s
}

// service start
// set onleader callback
// try to select a leader
// keep try to select a leader
// keep alive
func (s *Service) Start(onLeader func(isLeader bool, id int64)) {
	s.onLeader = onLeader
	s.selectLeader()
	go s.tryGetLeader()
	// 更新updated，此字段用于判断服务是否存活
	go s.keepAlive()
}

// panic if query database error
func (s *Service) init() {
	row := s.db.QueryRow("SELECT `id`, `updated` FROM `services` WHERE `address`=?", s.Name + "-" + s.Address)
	var id, updated int64
	err := row.Scan(&id, &updated)
	if err != nil && err != sql.ErrNoRows {
		panic(err)
	}
	if err == nil {
		s.ID = id
		s.Updated = updated
	}
}

// panic if error happened
func (s *Service) selectLeader() {
	v, err := s.redis.Incr(s.leaderKey).Result()
	if err != nil {
		log.Errorf("selectLeader s.redis.Incr fail, error=[%v]", err)
		panic(err)
		return
	}
	s.Leader = v == 1
	s.onLeader(s.Leader, s.ID)
	if err = s.redis.Expire(s.leaderKey, time.Second * 6).Err(); nil != err {
		log.Errorf("selectLeader s.redis.Expire fail, error=[%v]", err)
		panic(err)
	}
	if s.Leader {
		if err = s.updateIsLeader(); nil != err {
			panic(err)
		}
	}
}

// update service to leader
func (s *Service) updateIsLeader() error {
	sqlStr := "UPDATE `services` SET `is_leader`=0 WHERE id!=?"
	_, err := s.db.Exec(sqlStr, s.ID)
	if err != nil {
		log.Errorf("updateIsLeader s.db.Exec fail, error=[%v]", err)
		return err
	}
	sqlStr = "UPDATE `services` SET `is_leader`=1 WHERE id=?"
	_, err = s.db.Exec(sqlStr, s.ID)
	if err != nil {
		log.Errorf("updateIsLeader s.db.Exec fail, error=[%v]", err)
		return err
	}
	return nil
}

// keep try to select a new leader
// if old leader is offline
func (s *Service) tryGetLeader()  {
	if s.Leader {
		return
	}
	for {
		v, err := s.redis.Incr(s.leaderKey).Result()
		if err != nil {
			log.Errorf("tryGetLeader s.redis.Incr fail, error=[%v]", err)
			s.Status = 0
			continue
		}
		if v == 1 {
			if err = s.updateIsLeader(); nil != err {
				// try to free the current leader
				s.redis.Del(s.leaderKey)
				s.Status = 0
				continue
			}
			s.lock.Lock()
			s.Leader = true
			s.lock.Unlock()
			s.onLeader(s.Leader, s.ID)
		}
		s.Status = 1
		time.Sleep(time.Second * 6)
	}
}

// 服务注册
// panic if error happened
func (s *Service) register() (int64, error) {
	defer s.onRegister(s.ID)
	if s.ID > 0 {
		return s.ID, nil
	}
	res, err := s.db.Exec("INSERT INTO `services`(`address`, `updated`) VALUES (?,?)", s.Name + "-" + s.Address, time.Now().Unix())
	if err != nil {
		log.Errorf("Register s.db.Exec fail, error=[%v]", err)
		panic(err)
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		log.Errorf("Register res.LastInsertId fail, error=[%v]", err)
		panic(err)
		return 0, err
	}
	s.ID = id
	s.Updated = time.Now().Unix()
	return id, nil
}

// keep service alive
func (s *Service) keepAlive() (error) {
	for {
		if s.ID <= 0 {
			time.Sleep(time.Second * 1)
			continue
		}

		s.lock.RLock()
		if s.Leader {
			if err := s.redis.Expire(s.leaderKey, time.Second * 6).Err(); nil != err {
				log.Errorf("keepAlive s.redis.Expire fail, error=[%v]", err)
				s.Status = 0
				s.lock.RUnlock()
				time.Sleep(time.Second * 1)
				continue
			}
		}
		s.lock.RUnlock()

		t := time.Now().Unix()
		res, err := s.db.Exec("UPDATE `services` SET `updated`=? WHERE id=?", t, s.ID)
		if err != nil {
			s.Status = 0
			log.Errorf("keepAlive s.db.Exec fail, error=[%v]", err)
			time.Sleep(time.Second * 1)
			continue
		}
		_, err = res.RowsAffected()
		if err != nil {
			s.Status = 0
			log.Errorf("keepAlive res.RowsAffected fail, error=[%v]", err)
			time.Sleep(time.Second * 1)
			continue
		}

		s.Status = 1
		s.Updated = t
		time.Sleep(time.Second * 1)
	}
}

// 服务注销
func (s *Service) Deregister() error {
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
	if s.Leader {
		s.redis.Del(s.leaderKey)
		s.Status = 0
		s.Leader = false
	}
	return nil
}

// get all service
func (s *Service) GetServices() ([]*Service, error) {
	services := make([]*Service, 0)
	rows, err := s.db.Query("SELECT `id`, `address`, `is_leader`, `updated` FROM `services` WHERE 1")
	if err != nil {
		log.Errorf("GetServices s.db.Query fail, error=[%v]", err)
		return nil, err
	}
	for rows.Next() {
		sr := new(Service)
		var leader int
		err = rows.Scan(&sr.ID, &sr.Address, &leader, &sr.Updated)
		if err != nil {
			log.Errorf("GetServices rows.Scan fail, error=[%v]", err)
			continue
		}
		sr.Status = 1
		sr.Leader = leader == 1
		if time.Now().Unix() - sr.Updated >= 6 {
			sr.Status = 0
		}
		services = append(services, sr)
	}
	return services, nil
}