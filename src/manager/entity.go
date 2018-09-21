package manager

import (
	"fmt"
	"errors"
	"strconv"
	"strings"
	"library/time"
)

type httpParamsEntity struct {
	// 数据库的基本属性
	Id interface{}      `json:"id"`
	CronSet string      `json:"cron_set"`
	Command string      `json:"command"`
	Remark string       `json:"remark"`
	Stop interface{}    `json:"stop"`
	StartTime string    `json:"start_time"`
	EndTime string      `json:"end_time"`
	IsMutex interface{} `json:"is_mutex"`
	Blame string        `json:"blame"`

	UserName string     `json:"user_name"`
	Password string     `json:"password"`
	//realName, phone
	RealName string     `json:"real_name"`
	Phone interface{}   `json:"phone"`
}
var ErrNil = errors.New("nil")
func (p *httpParamsEntity) GetPhone() string {
	return fmt.Sprintf("%v", p.Phone)
}
func (p *httpParamsEntity) IsStop() bool {
	if p == nil {
		return false
	}
	return fmt.Sprintf("%v", p.Stop) == "1"
}

func (p *httpParamsEntity) GetCronSet() string {
	if p == nil {
		return ""
	}
	p.CronSet = strings.Trim(p.CronSet, " ")
	return p.CronSet
}

func (p *httpParamsEntity) GetCommand() string {
	if p == nil {
		return ""
	}
	p.Command = strings.Trim(p.Command, " ")
	return p.Command
}

func (p *httpParamsEntity) GetRemark() string {
	if p == nil {
		return ""
	}
	p.Remark = strings.Trim(p.Remark, " ")
	return p.Remark
}

func (p *httpParamsEntity) GetStartTime() (string, error) {
	if p == nil {
		return "", ErrNil
	}
	if p.StartTime == "" {
		return "", nil
	}
	t := time.StrToTime(p.StartTime)
	if t <= 0 {
		return "", errors.New("convert fail")
	}
	return p.StartTime, nil
}

func (p *httpParamsEntity) GetEndTime() (string, error) {
	if p == nil {
		return "", ErrNil
	}
	if p.StartTime == "" {
		return "", nil
	}
	t := time.StrToTime(p.EndTime)
	if t <= 0 {
		return "", errors.New("convert fail")
	}
	return p.EndTime, nil
}

func (p *httpParamsEntity) GetId() (int64, error) {
	if p == nil {
		return 0, ErrNil
	}
	return strconv.ParseInt(fmt.Sprintf("%v", p.Id), 10, 64)
}

func (p *httpParamsEntity) Mutex() bool {
	if p == nil {
		return false
	}
	return fmt.Sprintf("%v", p.IsMutex) == "1"
}

