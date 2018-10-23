package manager

import (
	"fmt"
	"errors"
	"strconv"
	"strings"
	"library/time"
)

type httpParamsEntity struct {
	// 定时任务表单
	Id interface{}      `json:"id"`
	CronSet string      `json:"cron_set"`
	Command string      `json:"command"`
	Remark string       `json:"remark"`
	Stop interface{}    `json:"stop"`
	StartTime string    `json:"start_time"`
	EndTime string      `json:"end_time"`
	IsMutex interface{} `json:"is_mutex"`
	Blame interface{}   `json:"blame"`

	// 用户相关表单
	UserName string     `json:"username"`
	Password string     `json:"password"`
	//realName, phone
	RealName string     `json:"real_name"`
	Phone interface{}   `json:"phone"`
	Enable interface{}  `json:"enable"`

	Powers []int64      `json:"powers"`
}
var ErrNil = errors.New("nil")
func (p *httpParamsEntity) GetPhone() string {
	return fmt.Sprintf("%v", p.Phone)
}

func (p *httpParamsEntity) ISEnable() bool {
	return fmt.Sprintf("%v", p.Enable) == "1"
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

func (p *httpParamsEntity) GetBlame() int64 {
	if p == nil {
		return 0
	}
	strBlame := fmt.Sprintf("%v", p.Blame)
	i, _ := strconv.ParseInt(strBlame, 10, 64)
	return i
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

