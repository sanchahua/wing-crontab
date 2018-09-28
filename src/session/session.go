package session

import (
	"github.com/bsm/go-guid"
	"encoding/hex"
	"time"
	"github.com/go-redis/redis"
	"encoding/json"
)

type Session struct {
	redis *redis.Client
}

func NewSession(redis *redis.Client) *Session {
	s := &Session{
		redis: redis,
	}
	return s
}

// 返回sessionid，这个id会写到cookie返回给客户端
// 客户端下次请求带上此sessionid，验证是否在线，因为一个用户支持多端登录
// timeout 单位为秒，超过该事件没有活动的session会被清理掉
func (s *Session) Store(userid int64, timeout time.Duration) (string, error) {
	id := guid.New128()
	sessionid := hex.EncodeToString(id.Bytes())
	data, err := json.Marshal(userid)
	if err != nil {
		return "", err
	}
	err = s.redis.Set(sessionid, string(data), timeout).Err()
	if err != nil {
		return "", err
	}
	return sessionid, nil
}

func (s *Session) Update(sessionid string, timeout time.Duration) error {
	return s.redis.Expire(sessionid, timeout).Err()
}

// 用来检验一个session的有效性
func (s *Session) Valid(sessionid string) (bool, error) {
	v, err := s.redis.Exists(sessionid).Result()
	if err != nil {
		return false, err
	}
	return v >= 1, nil
}

func (s *Session) GetUserId(sessionid string) (int64, error) {
	v, err := s.redis.Get(sessionid).Int64()
	if err != nil {
		return 0, err
	}

	return v, nil
}