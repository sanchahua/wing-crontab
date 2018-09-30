package user

import (
	"database/sql"
	"strings"
	"gitlab.xunlei.cn/xllive/common/log"
	"errors"
	"library/time"
)

type Entity struct {
	//SELECT `id`, `user_name`, `password`, `real_name`,
	//`phone`, `created`, `updated` FROM `users` WHERE 1
	Id       int64  `json:"id"`
	UserName string `json:"user_name"`
	Password string `json:"password"`
	RealName string `json:"real_name"`
	Phone    string `json:"phone"`
	Created  string `json:"created"`
	Updated  string `json:"updated"`
	Enable   bool   `json:"enable"`
}

type User struct {
	db *sql.DB `json:"-"`
}

func NewUser(db *sql.DB) *User {
	return &User{
		db: db,
	}
}

// 根据用户名查询用户信息
// 一般登录使用
// userName 可以使用户名，也可以是手机号
func (u *User) GetUserByUserName(userName string) (*Entity, error) {
	userName = strings.Trim(userName, " ")
	if userName == "" {
		log.Errorf("GetUserByUserName fail, error=[userName invalid]")
		return nil, errors.New("userName invalid")
	}
	sqlStr := "select `id`, `user_name`, `password`, " +
		"`real_name`, `phone`, `created`, `updated`, `enable` " +
		"from users where " +
		"`user_name`=? or `phone`=?"
	data := u.db.QueryRow(sqlStr, userName, userName)
	var (
		row Entity
	)
	err := data.Scan(&row.Id, &row.UserName, &row.Password,
		&row.RealName, &row.Phone, &row.Created, &row.Updated, &row.Enable)
	if err != nil && err != sql.ErrNoRows {
		log.Errorf("GetUserByUserName data.Scan fail, sql=[%v], userName=[%v], error=[%v]", sqlStr, userName, err)
		return &row, err
	}
	if err == sql.ErrNoRows {
		return nil, nil
	}
	log.Infof("GetUserByUserName success, sql=[%v], userName=[%v], return=[%v]", sqlStr, userName, row)
	return &row, nil
}

func (u *User) Enable(id int64, enable bool) error {
	sqlStr := "UPDATE `users` SET `updated`=?,`enable`=? WHERE id=?"
	iEnable := 0
	if enable {
		iEnable = 1
	}
	_, err := u.db.Exec(sqlStr, time.GetDayTime(), iEnable, id)
	return err
}

func (u *User) Update(id int64, userName, password, realName, phone string) error {
	if id <=0 {
		return errors.New("id param error")
	}
	if userName == "" {
		return errors.New("user name param error")
	}
	if password == "" {
		return errors.New("password param error")
	}
	// 检验userName、phone是否已存在
	sqlStr := "select `id` from users where `id`!=? and (`user_name`=? or `phone`=?)"
	data := u.db.QueryRow(sqlStr, id, userName, phone)
	var (
		exid int64
	)
	err := data.Scan(&exid)
	if err != sql.ErrNoRows {
		log.Errorf("Update data.Scan fail, sql=[%v], userName=[%v], error=[%v]", sqlStr, userName, err)
		return errors.New(userName + "或者" + phone + "已存在")
	}
	sqlStr = "UPDATE `users` SET `user_name`=?,`password`=?,`real_name`=?, `phone`=?, `updated`=? WHERE id=?"
	_, err = u.db.Exec(sqlStr, userName, password, realName, phone, time.GetDayTime(), id)
	return err
}

func  (u *User) GetUsers() ([]*Entity, error)  {
	sqlStr := "SELECT `id`, `user_name`, `password`, `real_name`, `phone`, `created`, `updated`, `enable` FROM `users` WHERE 1"
	rows, err := u.db.Query(sqlStr)
	if err != nil {
		return nil, err
	}
	var data = make([]*Entity, 0)
	for rows.Next() {
		var e Entity
		var enable int
		err = rows.Scan(&e.Id, &e.UserName, &e.Password, &e.RealName, &e.Phone, &e.Created, &e.Updated, &enable)
		if err != nil {
			continue
		}
		e.Enable = enable == 1
		e.Password = "******"
		data = append(data, &e)
	}
	return data, nil
}

func  (u *User) GetUserInfo(id int64) (*Entity, error)  {
	sqlStr := "SELECT `id`, `user_name`, `password`, `real_name`, `phone`, `created`, `updated`, `enable` FROM `users` WHERE id=?"
	row := u.db.QueryRow(sqlStr, id)
	var e Entity
	var enable int
	err := row.Scan(&e.Id, &e.UserName, &e.Password, &e.RealName, &e.Phone, &e.Created, &e.Updated, &enable)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	e.Enable = enable == 1
	//e.Password = "******"
	return &e, nil
}

func  (u *User) Delete(id int64) (error)  {
	sqlStr := "DELETE FROM `users` WHERE id=?"
	_, err := u.db.Exec(sqlStr, id)
	if err != nil {
		return err
	}
	return nil
}

func (u *User) Add(userName, password, realName, phone string) (int64, error) {
	// 判断用户名是否已被占用
	userinfo, err := u.GetUserByUserName(userName)
	if err != nil {
		return 0, err
	}
	if userinfo != nil {
		return 0, errors.New(userName + "已经存在")
	}
    // 判断手机号是否已被使用
	userinfo, err = u.GetUserByUserName(phone)
	if err != nil {
		return 0, err
	}
	if userinfo != nil {
		return 0, errors.New(phone + "已经存在")
	}
	password = strings.Trim(password, " ")
	if password == "" {
		return 0, errors.New("密码不能为空")
	}
	sqlStr := "INSERT INTO `users`(`user_name`, `password`, `real_name`, " +
		"`phone`, `created`, `updated`) " +
		"VALUES (?, ?, ?, ?, ?, ?)"
	created := time.GetDayTime()
	res, err := u.db.Exec(sqlStr, userName, password, realName, phone, created, created)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}
