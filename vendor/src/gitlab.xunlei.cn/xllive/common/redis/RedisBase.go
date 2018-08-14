package redis

import (
	"github.com/heipacker/redis"
	"time"
	"fmt"
	"github.com/golang/protobuf/proto"
	"gitlab.xunlei.cn/xllive/common/utils"
)

const Nil = redis.Nil

type RedisBase struct{
	addr      string
	password  string
	client    *redis.Client
}

func NewRedisBase(addr, password string) (*RedisBase, error) {
	r := new(RedisBase)
	if err := r.init(addr, password); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *RedisBase)init(addr, password string) error {

	r.addr = addr
	r.password = password
	r.client = redis.NewClient(&redis.Options{Addr:r.addr, Password: r.password, DB:0})
	_, err := r.client.Ping().Result()
	if err != nil {
		return err
	}

	return nil
}

func (r *RedisBase)Lock(key string, expiration time.Duration) error {
	v, err := r.client.Incr(key).Result()
	if err != nil {
		return err
	}
	if v != 1 {
		return fmt.Errorf("Incr fail, v='%d'", v)
	}

	tm := time.Now().Add(expiration)
	if err = r.ExpireAt(key, tm); err != nil {
		r.client.Del(key)
		return err
	}

	 return nil
}

func (r *RedisBase)Unlock(key string) error {
	 _, err := r.client.Del(key).Result()
	 return err
}

func (r *RedisBase)Expire(key string, expiration time.Duration) error {
	b, err := r.client.Expire(key, expiration).Result()
	if err != nil {
		return err
	}

	if b == false {
		return fmt.Errorf("Expire fail, expiration='%d'", expiration)
	}

	return nil
}

func (r *RedisBase)ExpireAt(key string, tm time.Time) error {
	b, err := r.client.ExpireAt(key, tm).Result()
	if err != nil {
		return err
	}

	if b == false {
		return fmt.Errorf("ExpireAt fail, tm='%v'", tm)
	}

	return nil
}


func (r * RedisBase)HGetJson(key, field string, v interface{}) error {
	rsp, err := r.client.HGet(key, field).Result()
	if err != nil {
		return err
	}

	if err = utils.JsonGoUnmarshal([]byte(rsp), v); err != nil {
		return err
	}
	return nil
}

func (r * RedisBase)HSetJson(key, field string, v interface{}) error {

	var strValue string
	if v != nil {
		if bytesValue, err := utils.JsonOmitMarshal(v); err != nil {
			return err
		} else {
			strValue = string(bytesValue)
		}
	}

	if _, err := r.client.HSet(key, field, strValue).Result(); err != nil {
		return err
	}

	return nil
}

func (r * RedisBase)HSetJsonpb(key, field string, pb proto.Message) error {

	var strValue string
	if pb != nil {
		if bytesValue, err := utils.JsonpbMarshal(pb, true); err != nil {
			return err
		} else {
			strValue = string(bytesValue)
		}
	}

	if _, err := r.client.HSet(key, field, strValue).Result(); err != nil {
		return err
	}

	return nil
}

func (r *RedisBase)HGetAll(key string) (map[string]string, error) {
	return r.client.HGetAll(key).Result()
}

func (r *RedisBase)HSetAll(key string, values map[string]string) error {
	for field, value := range values {
		if _, err := r.client.HSet(key, field, value).Result(); err != nil {
			return err
		}
	}
	return nil
}

func (r *RedisBase)HDel(key string, fields ...string) error {
	_, err := r.client.HDel(key, fields...).Result()
	return err
}

func (r *RedisBase)Del(keys ...string) error {
	_, err := r.client.Del(keys...).Result()
	return err
}

func (r *RedisBase)SetJson(key string, value interface{}, expiration time.Duration) error {

	bytesValue, err := utils.JsonOmitMarshal(value)
	if err != nil {
		return err
	}

	strValue := string(bytesValue)
	if err := r.client.Set(key, strValue, expiration).Err(); err != nil {
		return err
	}

	if err := r.Expire(key, expiration); err != nil {
		return err
	}

	return nil
}

