package manager

import (
	"time"
	shttp "net/http"
)

// 保持session的有效性
func (m *CronManager) sessionValid(r *shttp.Request) bool {
	cookies := m.readCookie(r)
	sessionid, ok := cookies["Session"]
	if ok {
		if v, _ := m.session.Valid(sessionid); !v {
			return false
		}
		m.session.Update(sessionid, time.Second * 60)
		return true
	}
	return false
}

func (m *CronManager) csrfCheck(r *shttp.Request) bool {
	cookies := m.readCookie(r)
	sessionid, _ := cookies["Session"]
	return r.Header.Get("Session") == sessionid
}
