package event

import (
	"gitlab.xunlei.cn/xllive/common/log"
	"time"
	"encoding/json"
	"github.com/go-redis/redis"
)

const (
	EV_ADD int64 = 1 << iota
	EV_DELETE
	EV_UPDATE
	EV_START
	EV_STOP
	EV_DISABLE_MUTEX
	EV_ENABLE_MUTEX
	EV_KILL
	EV_OFFLINE
)

type Event struct {
	callback map[int64]EventCallback
	watchKey string
	redis *redis.Client
}
type EventCallback func(int64, ...int64)
func NewEvent(watchKey string, redis *redis.Client) *Event {
	return &Event{
		watchKey: watchKey,
		callback: make(map[int64]EventCallback),
		redis: redis,
	}
}

func (ev *Event) RegisterEventCallback(event int64, callback EventCallback) {
	ev.callback[event] = callback
}

func (ev *Event) Watch() {
	log.Tracef("start watchCron [%v]", ev.watchKey)
	var raw = make([]int64, 0)
	for {
		data, err := ev.redis.BRPop(time.Second * 3, ev.watchKey).Result()
		if err != nil {
			if err != redis.Nil {
				log.Errorf("watchCron redis.BRPop fail, error=[%v]", err)
			}
			continue
		}
		log.Tracef("watchCron data=[%v]", data)
		if len(data) < 2 {
			log.Errorf("watchCron data len fail, error=[%v]", err)
			continue
		}
		err = json.Unmarshal([]byte(data[1]), &raw)
		if err != nil {
			log.Errorf("watchCron json.Unmarshal fail, error=[%v]", err)
			continue
		}
		if len(raw) < 2 {
			log.Errorf("watchCron raw len fail, error=[%v]", err)
			continue
		}
		event := raw[0]
		id := raw[1]

		call, ok := ev.callback[event]
		if ok {
			call(id, raw...)
		}
	}
}