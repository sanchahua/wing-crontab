package service

import (
	"database/sql"
	"encoding/json"
	"time"
	"gitlab.xunlei.cn/xllive/common/log"
	"os"
)

type Service struct {
	ID int64
	Address string
	Tags []string
	Updated int64
	db *sql.DB
	onRegister OnRegisterFunc
	onServiceDown func(int64)
	onServiceUp   func(int64)
	status int // 1 ok 0 oofline
	services map[int64]*Service
	Name string
}

type OnRegisterFunc func(runTimeId int64)

func NewService(
	db *sql.DB,                // 数据库操作资源句柄
	Address string,            // 服务地址， 如 127.0.0.1：38001
	Tags []string,             // 标签
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
		Tags:          Tags,
		onServiceDown: onServiceDown,
		onServiceUp:   onServiceUp,
		services:      make(map[int64]*Service),
		status:        1,
	}
	s.init()
	go s.keepAlive()
	go s.checkServiceDown()
	return s
}

func (s *Service) init() {
	row := s.db.QueryRow("SELECT `id`, `updated` FROM `services` WHERE `address`=?",
		s.Name+"-"+s.Address)
	var id int64
	var updated int64
	err := row.Scan(&id, &updated)
	if err == nil {
		s.ID = id
		s.Updated = updated
	}
	rows, err := s.db.Query("SELECT `id`, `address`, `tags`, `updated` FROM `services` WHERE 1")
	if err == nil {
		for rows.Next() {
			sr := new(Service)
			var tags string
			err = rows.Scan(&sr.ID, &sr.Address, &tags, &sr.Updated)
			if err != nil {
				continue
			}
			json.Unmarshal([]byte(tags), &sr.Tags)
			s.services[sr.ID] = sr
		}
	}
}

// 服务注册
func (s *Service) Register() (int64, error) {
	defer s.onRegister(s.ID)
	if s.ID > 0 {
		return s.ID, nil
	}
	jsonTags, _ := json.Marshal(s.Tags)
	res, err := s.db.Exec("INSERT INTO `services`(`address`, `tags`, `updated`) VALUES (?,?,?)",
		s.Name+"-"+s.Address, string(jsonTags), time.Now().Unix())
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

// 服务注销
func (s *Service) Deregister() (error) {
	res, err := s.db.Exec("DELETE FROM `services` WHERE id=?",
		s.ID)
	if err != nil {
		return err
	}
	_, err = res.RowsAffected()
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) checkServiceDown() {
	time.Sleep(time.Second * 3)
	for {
		rows, err := s.db.Query("select id, updated from services where 1")
		if err != nil {
			time.Sleep(time.Second * 1)
			continue
		}
		for rows.Next() {
			var id, updated int64
			err = rows.Scan(&id, &updated)
			if err != nil {
				continue
			}
			sr, ok := s.services[id]
			if !ok {
				continue
			}
			oldStatus := sr.status
			if time.Now().Unix() - updated > 3 {
				sr.status = 0
			} else {
				sr.status = 1
			}
			// 如果是当前节点，更新下当前节点的状态
			if id == s.ID && s.status != sr.status {
				s.status = sr.status
			}
			if oldStatus == 0 && sr.status == 1 {
				// 上线
				s.onServiceUp(id)
			}
			if oldStatus == 1 && sr.status == 0 {
				// 下线
				s.onServiceDown(id)
			}
		}
		time.Sleep(time.Second * 1)
	}
}