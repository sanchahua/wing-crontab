package service

import (
	"database/sql"
	"time"
	"gitlab.xunlei.cn/xllive/common/log"
	"os"
	"github.com/go-redis/redis"
	"sync/atomic"
)

type Service struct {
	ID            int64 `json:"ID"`
	Address       string
	Updated       int64
	db            *sql.DB `json:"-"`
	onRegister    OnRegisterFunc `json:"-"`
	onServiceDown func(int64) `json:"-"`
	onServiceUp   func(int64) `json:"-"`
	onLeader      func(isLeader bool, id int64) `json:"-"`
	Status        int64       // 1在线 0离线
	Name          string
	Leader        int64
	leaderKey     string `json:"-"`
	redis         *redis.Client `json:"-"`
	Unique        string
	Offline       int64
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
		Leader:        0,
		onLeader:      nil,
		Unique:        name + "-" + Address,
		Offline:       0,
	}
	// 初始化，主要检查服务是否存在，如果存在会初始化ID
	s.init()
	s.register()
	return s
}

func (s *Service) SetOffline(serviceId int64, offline bool) {
	if serviceId != s.ID {
		return
	}
	if offline {
		atomic.StoreInt64(&s.Offline, 1)
	} else {
		atomic.StoreInt64(&s.Offline, 0)
	}
}

func (s *Service) IsOffline() bool {
	if atomic.LoadInt64(&s.Offline) == 1 {
		log.Warnf("node [%v] is offline", s.ID)
		return true
	}
	return false//atomic.LoadInt64(&s.Offline) == 1
}

// service start
// set onleader callback
// try to select a leader
// keep try to select a leader
// keep alive
func (s *Service) Start(onLeader func(isLeader bool, id int64)) {
	s.onLeader = onLeader
	if 1 != atomic.LoadInt64(&s.Offline) {
		s.selectLeader()
	}
	go s.tryGetLeader()
	// 更新updated，此字段用于判断服务是否存活
	go s.keepAlive()
}

// panic if query database error
func (s *Service) init() {
	row := s.db.QueryRow("SELECT `id`, `updated`, `offline` FROM `services` WHERE `name`=? and `address`=?", s.Name, s.Address)
	var id, updated, offline int64
	err := row.Scan(&id, &updated, &offline)
	if err != nil && err != sql.ErrNoRows {
		panic(err)
	}
	if err == nil {
		s.ID = id
		s.Updated = updated
		atomic.StoreInt64(&s.Offline, offline)
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

	if 1 == v {
		atomic.StoreInt64(&s.Leader, 1)
		s.onLeader(true, s.ID)
	} else {
		atomic.StoreInt64(&s.Leader, 0)
		s.onLeader(false, s.ID)
	}

	if err = s.redis.Expire(s.leaderKey, time.Second * 6).Err(); nil != err {
		log.Errorf("selectLeader s.redis.Expire fail, error=[%v]", err)
		panic(err)
	}
	if 1 == atomic.LoadInt64(&s.Leader) {
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

func (s *Service) UpdateOffline(serviceId, offline int64) error {
	sqlStr := "UPDATE `services` SET `offline`=? WHERE id=?"
	_, err := s.db.Exec(sqlStr, offline, serviceId)
	if err != nil {
		log.Errorf("UpdateOffline s.db.Exec fail, error=[%v]", err)
		return err
	}
	return nil
}


// keep try to select a new leader
// if old leader is offline
func (s *Service) tryGetLeader()  {
	for {
		if 1 == atomic.LoadInt64(&s.Leader) {
			continue
		}
		// if offline, do not try again
		if 1 == atomic.LoadInt64(&s.Offline) {
			time.Sleep(time.Second)
			continue
		}
		v, err := s.redis.Incr(s.leaderKey).Result()
		if err != nil {
			log.Errorf("tryGetLeader s.redis.Incr fail, error=[%v]", err)
			//s.Status = 0
			atomic.StoreInt64(&s.Status, 0)
			time.Sleep(time.Second)
			continue
		}
		if v == 1 {
			if err = s.updateIsLeader(); nil != err {
				// try to free the current leader
				s.redis.Del(s.leaderKey)
				//s.Status = 0
				atomic.StoreInt64(&s.Status, 0)
				atomic.StoreInt64(&s.Leader, 0)
				time.Sleep(time.Second)
				continue
			}
			atomic.StoreInt64(&s.Leader, 1)
			s.onLeader(true, s.ID)
		} else {
			atomic.StoreInt64(&s.Leader, 0)
		}
		//s.Status = 1
		atomic.StoreInt64(&s.Status, 1)
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
	res, err := s.db.Exec("INSERT INTO `services`(`name`, `address`, `updated`) VALUES (?, ?,?)", s.Name, s.Address, time.Now().Unix())
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

		if 1 == atomic.LoadInt64(&s.Leader) {
			if 1 == atomic.LoadInt64(&s.Offline) {
				log.Warnf("node offline, try to free leader")
				atomic.StoreInt64(&s.Leader, 0)
				s.redis.Del(s.leaderKey)
				s.onLeader(false, s.ID)
			}
			if err := s.redis.Expire(s.leaderKey, time.Second * 6).Err(); nil != err {
				log.Errorf("keepAlive s.redis.Expire fail, error=[%v]", err)
				//s.Status = 0
				atomic.StoreInt64(&s.Status, 0)
				time.Sleep(time.Second * 1)
				continue
			}
		}

		t := time.Now().Unix()
		res, err := s.db.Exec("UPDATE `services` SET `updated`=? WHERE id=?", t, s.ID)
		if err != nil {
			//s.Status = 0
			atomic.StoreInt64(&s.Status, 0)
			log.Errorf("keepAlive s.db.Exec fail, error=[%v]", err)
			time.Sleep(time.Second * 1)
			continue
		}
		_, err = res.RowsAffected()
		if err != nil {
			//s.Status = 0
			atomic.StoreInt64(&s.Status, 0)
			log.Errorf("keepAlive res.RowsAffected fail, error=[%v]", err)
			time.Sleep(time.Second * 1)
			continue
		}

		//s.Status = 1
		atomic.StoreInt64(&s.Status, 1)
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
	if 1 == atomic.LoadInt64(&s.Leader) {
		s.redis.Del(s.leaderKey)
		atomic.StoreInt64(&s.Status, 1)
		atomic.StoreInt64(&s.Leader, 0)
	}
	return nil
}

// get all service
func (s *Service) GetServices() ([]*Service, error) {
	services := make([]*Service, 0)
	rows, err := s.db.Query("SELECT `id`,`name`, `address`, `is_leader`, `updated`, `offline` FROM `services` WHERE 1")
	if err != nil {
		log.Errorf("GetServices s.db.Query fail, error=[%v]", err)
		return nil, err
	}
	for rows.Next() {
		sr := new(Service)
		err = rows.Scan(&sr.ID, &sr.Name, &sr.Address, &sr.Leader, &sr.Updated, &sr.Offline)
		if err != nil {
			log.Errorf("GetServices rows.Scan fail, error=[%v]", err)
			continue
		}
		atomic.StoreInt64(&sr.Status, 1)
		if time.Now().Unix() - sr.Updated >= 6 {
			atomic.StoreInt64(&sr.Status, 0)
		}
		sr.Unique = sr.Name + "-" + sr.Address
		services = append(services, sr)
	}
	return services, nil
}