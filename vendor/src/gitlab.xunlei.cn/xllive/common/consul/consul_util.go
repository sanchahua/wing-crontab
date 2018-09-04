package consul

import (
	"encoding/json"
	"errors"
	log "github.com/cihub/seelog"
	http "gitlab.xunlei.cn/xllive/common/net"
)
var IDErr = errors.New("ID key does not exists")
var SessionEmpty = errors.New("session empty")

// create a session, use for lock a key
func (con *Consul) createSession() (string, error) {
	request := http.NewHttp("http://" + con.serviceIp + "/v1/session/create")
	p := make(map[string] interface{})
	p["Name"] = "pw/consul/lock/service"
	p["LockDelay"] = "1s"
	p["TTL"] = "10s"
	p["Behavior"] = "delete"
	params, _:=json.Marshal(p)
	res, err := request.Put(params)
	if err != nil {
		return "", err
	}
	var arr interface{}
	err = json.Unmarshal(res, &arr)
	if err != nil {
		return "", err
	}
	m := arr.(map[string] interface{});
	id, ok := m["ID"]
	if !ok {
		return "", IDErr
	}
	log.Debugf("session id: %s", id.(string))
	return id.(string), nil
}

// key string 需要锁定的唯一key
// timeout 设定超时时间，单位为毫秒， 超时时间不能设置为0
func (con *Consul) Lock(key string, timeout int64) (bool, error) {
	con.initSession()
	if con.session == "" {
		return false, SessionEmpty
	}
	lockApi := "http://" + con.serviceIp +"/v1/kv/" + key + "?acquire=" + con.session
	request := http.NewHttp(lockApi)
	res, err := request.Put(nil)
	if err != nil {
		log.Errorf("lock error: %+v", err)
		con.session = ""
		return false, err
	}
	log.Debugf("lock return: %s", string(res))
	if string(res) == "true" {
		con.addLock(key, timeout, 0)
		return true, nil
	}
	con.session = ""
	return false, nil
}

// unlock
func (con *Consul) Unlock(key string) (bool, error) {
	con.initSession()
	if con.session == "" {
		return false, SessionEmpty
	}
	unlockApi := "http://" + con.serviceIp +"/v1/kv/" + key + "?release=" + con.session
	request := http.NewHttp(unlockApi)
	res, err := request.Put(nil)
	if err != nil {
		log.Errorf("Unlock error: %+v", err)
		return false, err
	}
	log.Debugf("unlock return: %s", string(res))
	if string(res) == "true" {
		con.clearLock(key, true)
		return true, nil
	}
	return false, nil
}

// 删除一个key，用以用来强制释放一个锁
func (con *Consul) Delete(key string) (bool, error) {
	con.initSession()
	if con.session == "" {
		return false, SessionEmpty
	}
	//defer con.writeMember()
	url := "http://" + con.serviceIp +"/v1/kv/" + key
	request := http.NewHttp(url)
	res, err := request.Delete()
	if err != nil {
		log.Debugf("delete err: %+v", err)
		return false, err
	}
	log.Debugf("delete %s return---%s", key, string(res))
	if string(res) == "true" {
		con.clearLock(key, false)
		con.client.Delete(PREFIX_LOCK_INFO + key)
		return true, nil
	}
	return false, nil
}
