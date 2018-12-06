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
	db            *sql.DB                       `json:"-"`
	leaderKey     string                        `json:"-"`
	redis         *redis.Client                 `json:"-"`
	ID            int64   `json:"ID"`
	Address       string  `json:"Address"`
	Updated       int64   `json:"Updated"`
	Status        int64   `json:"Status"`    // 1在线 0离线
	Name          string  `json:"Name"`
	Leader        int64   `json:"Leader"`
	Unique        string  `json:"Unique"`
	Offline       int64   `json:"Offline"`
}


// new service
func NewService(
	db *sql.DB,                // 数据库操作资源句柄
	Address string,            // 服务地址， 如 127.0.0.1：38001
	leaderKey string,
	redis *redis.Client,
) *Service {
	name := "xcrontab"
	n, _ := os.Hostname()
	if "" != n {
		name = n
	}
 	s := &Service{
		db:            db,
		Address:       Address,
		Name:          name,
		Status:        1,
		leaderKey:     leaderKey,
		redis:         redis,
		Leader:        0,
		Unique:        name + "-" + Address,
		Offline:       0,
	}
	// 初始化，主要检查服务是否存在，如果存在会初始化ID
	s.init()
	s.register()
 	go s.keepAlive()
	return s
}

// set node to offline or online
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

// check node is offline
func (s *Service) IsOffline() bool {
	if atomic.LoadInt64(&s.Offline) == 1 {
		log.Warnf("node [%v] is offline", s.ID)
		return true
	}
	return false
}

// panic if query database error
// init service info at start
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
func (s *Service) SelectLeader() bool {
	v, err := s.redis.Incr(s.leaderKey).Result()
	if err != nil {
		log.Errorf("selectLeader s.redis.Incr fail, error=[%v]", err)
		return false
	}
	if 1 != v {
		atomic.StoreInt64(&s.Leader, 0)
		return false
	}
	if err = s.redis.Expire(s.leaderKey, time.Second * 3).Err(); nil != err {
		log.Errorf("selectLeader s.redis.Expire fail, error=[%v]", err)
		//return false
	}
	atomic.StoreInt64(&s.Leader, 1)
	return true
}

func (s *Service) FreeLeader() {
	s.redis.Del(s.leaderKey)
	atomic.StoreInt64(&s.Leader, 0)
}


// keep service alive
func (s *Service) keepAlive() (error) {
	for {
		if s.ID <= 0 {
			time.Sleep(time.Second * 1)
			continue
		}

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


// update db
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


func (s *Service) updateNotIsLeader() error {
	sqlStr := "UPDATE `services` SET `is_leader`=0 WHERE id=?"
	_, err := s.db.Exec(sqlStr, s.ID)
	if err != nil {
		log.Errorf("updateNotIsLeader s.db.Exec fail, error=[%v]", err)
		return err
	}
	return nil
}

// update db
// update service to offline or online
func (s *Service) UpdateOffline(serviceId, offline int64) error {
	sqlStr := "UPDATE `services` SET `offline`=? WHERE id=?"
	_, err := s.db.Exec(sqlStr, offline, serviceId)
	if err != nil {
		log.Errorf("UpdateOffline s.db.Exec fail, error=[%v]", err)
		return err
	}
	return nil
}

// 服务注册
// panic if error happened
// only register at start
func (s *Service) register() (int64, error) {
	//defer s.onRegister(s.ID)
	if s.ID > 0 {
		return s.ID, nil
	}
	res, err := s.db.Exec("INSERT INTO `services`(`name`, `address`, `updated`) VALUES (?,?,?)", s.Name, s.Address, time.Now().Unix())
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
		atomic.StoreInt64(&s.Leader, 0)
		//s.onLeader(false, s.ID)
	}
	return nil
}

func (s *Service) IsLeader() bool {
	return 1 == atomic.LoadInt64(&s.Leader)
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

func (s *Service) SearchService(id int64) (*Service, error) {
	//services := make([]*Service, 0)
	rows := s.db.QueryRow("SELECT `id`,`name`, `address`, `is_leader`, `updated`, `offline` " +
		"FROM `services` WHERE id=?", id)

	sr := new(Service)
	err := rows.Scan(&sr.ID, &sr.Name, &sr.Address, &sr.Leader, &sr.Updated, &sr.Offline)
	if err != nil {
		log.Errorf("SearchService rows.Scan fail, error=[%v]", err)
		return nil, err
	}
	atomic.StoreInt64(&sr.Status, 1)
	if time.Now().Unix() - sr.Updated >= 6 {
		atomic.StoreInt64(&sr.Status, 0)
	}
	sr.Unique = sr.Name + "-" + sr.Address

	return sr, nil
}