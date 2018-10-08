package manager

import (
	"fmt"
	"strings"
	shttp "net/http"
)

func (m *CronManager) readCookie(r *shttp.Request) map[string]string {
	cookie := r.Header.Get("Cookie")
	fmt.Println("cookie:", cookie)
	if cookie == "" {
		return nil
	}
	//Session=000000005baebf74f905216dbc000001; qaerwger=qertwer
	temp1 := strings.Split(cookie, ";")
	if len(temp1) <= 0 {
		return nil
	}
	var cookies = make(map[string]string)
	for _, v := range temp1 {
		t := strings.Split(v, "=")
		if len(t) < 2 {
			continue
		}
		cookies[strings.Trim(t[0], " ")] = strings.Trim(t[1], " ")
	}
	return cookies
}

