package consul

import (
	consulkv "github.com/armon/consul-kv"
	log "github.com/sirupsen/logrus"
	http "net/http"
	"time"
	"sync"
)

const (
	PREFIX_LOCK_INFO = "pw/lock/"
)

type ConsulLocks struct {
	Key string
	Timeout int64
	StartLockTime int64
}

type Consul struct {
	client *consulkv.Client
	serviceIp string //比如 127.0.0.1：8500
	session string
	lock *sync.Mutex
	locks map[string] *ConsulLocks
}

func DefaultConfig() *ConsulConfig {
	return &ConsulConfig{
		ServiceIp: "127.0.0.1:8500",
	}
}

//全局只需要new一次就可以了
func NewConsul(config *ConsulConfig) *Consul{
	con := &Consul{
		serviceIp: config.ServiceIp,
		lock:      new(sync.Mutex),
		session:   "",
		locks:     make(map[string] *ConsulLocks),
	}
	var err error
	con.initSession()
	kvConfig := &consulkv.Config{
		Address:    config.ServiceIp,
		HTTPClient: http.DefaultClient,
	}
	con.client, err = consulkv.NewClient(kvConfig)
	if err != nil {
		log.Panicf("new consul client with error: %+v", err)
	}
	//启动程序的时候读取所有的lock列表，加载到locks
	con.init()
	//检测死锁
	go con.checkDeadLock()
	go con.watch()
	go con.checkInit()
	return con
}

func (con *Consul) initSession() {
	if con.session != "" {
		return
	}
	session, err := con.createSession()
	if err != nil {
		log.Errorf("create consul session with error: %+v", err)
	} else {
		con.session = session
	}
}

func (con *Consul) init() {
	_, pairs, err := con.client.List(PREFIX_LOCK_INFO)
	if err != nil {
		log.Errorf("read list with error: %+v", err)
		return
	}
	for _, v := range pairs {
		key := string(v.Value[16:])
		con.lock.Lock()
		_, ok := con.locks[key]
		con.lock.Unlock()
		if !ok {
			//如果锁不存在当前的缓存里面，增加到缓存
			timeout := int64(v.Value[0]) | int64(v.Value[1])<<8 |
				int64(v.Value[2])<<16 | int64(v.Value[3])<<24 |
				int64(v.Value[4])<<32 | int64(v.Value[5])<<40 |
				int64(v.Value[6])<<48 | int64(v.Value[7])<<56
			startLockTime := int64(v.Value[8]) | int64(v.Value[9])<<8 |
				int64(v.Value[10])<<16 | int64(v.Value[11])<<24 |
				int64(v.Value[12])<<32 | int64(v.Value[13])<<40 |
				int64(v.Value[14])<<48 | int64(v.Value[15])<<56
			con.addLock(key, timeout, startLockTime)
		}
	}
}

func (con *Consul) checkInit() {
	for {
		con.init()
		time.Sleep(3 * time.Second)
	}
}

func (con *Consul) addLock(key string, timeout int64, startLockTime int64) {
	con.lock.Lock()
	defer con.lock.Unlock()
	l, ok := con.locks[key]
	if startLockTime <= 0 {
		startLockTime = int64(time.Now().UnixNano()/1000000)
	}
	if ok {
		l.Key = key
		l.Timeout = timeout
		l.StartLockTime = startLockTime
	} else {
		l = &ConsulLocks{
			Key : key,
			Timeout:timeout,
			StartLockTime:startLockTime,//int64(time.Now().UnixNano()/1000000),
		}
		con.locks[key] = l
	}
	k  := []byte(key)
	kl := len(k)
	r := make([]byte, 16 + kl)
	i := 0
	r[i] = byte(timeout); i++
	r[i] = byte(timeout >> 8); i++
	r[i] = byte(timeout >> 16); i++
	r[i] = byte(timeout >> 24); i++
	r[i] = byte(timeout >> 32); i++
	r[i] = byte(timeout >> 40); i++
	r[i] = byte(timeout >> 48); i++
	r[i] = byte(timeout >> 56); i++
	r[i] = byte(l.StartLockTime); i++
	r[i] = byte(l.StartLockTime >> 8); i++
	r[i] = byte(l.StartLockTime >> 16); i++
	r[i] = byte(l.StartLockTime >> 24); i++
	r[i] = byte(l.StartLockTime >> 32); i++
	r[i] = byte(l.StartLockTime >> 40); i++
	r[i] = byte(l.StartLockTime >> 48); i++
	r[i] = byte(l.StartLockTime >> 56); i++
	for _, b := range k {
		r[i] = b; i++
	}
	// put a kv lock info
	con.client.Put(PREFIX_LOCK_INFO + key, r, 0)
}

func (con *Consul) clearLock(key string, isdelete bool) {
	con.lock.Lock()
	delete(con.locks, key)
	con.lock.Unlock()
	//删除consul里面的key锁
	if !isdelete {
		return
	}
	if err := con.client.Delete(PREFIX_LOCK_INFO + key); err != nil {
		log.Errorf("client.Delete key %s with error: %+v", key, err)
	}
	if err := con.client.Delete(key); err != nil {
		log.Errorf("client.Delete key %s with error: %+v", key, err)
	}
}

// 死锁检测
func (con *Consul) checkDeadLock() {
	for {
		for _, v := range con.locks {
			current := int64(time.Now().UnixNano()/1000000)
			if current - v.StartLockTime >= v.Timeout {
				log.Warnf("key %s lock timeout, try to delete", v.Key)
				con.clearLock(v.Key, true)
			}
		}
		//精确度为10毫秒
		time.Sleep(time.Millisecond * 10)
	}
}

// 锁变化监听
// 其实即使监听新增加的锁，将新增加的锁加入到当前的锁缓存
func (con *Consul) watch() {
	for {
		meta, _, err := con.client.List(PREFIX_LOCK_INFO)
		if err != nil {
			log.Errorf("watch chang with error：%#v", err)
			time.Sleep(time.Second)
			continue
		}
		if meta == nil {
			time.Sleep(time.Second)
			continue
		}
		_, pairs, err := con.client.WatchList(PREFIX_LOCK_INFO, meta.ModifyIndex)
		if err != nil {
			log.Errorf("watch chang with error：%#v, %+v", err, pairs)
			time.Sleep(time.Second)
			continue
		}
		if pairs == nil {
			time.Sleep(time.Second)
			continue
		}
		for _, v :=range pairs {
			key := string(v.Value[16:])
			con.lock.Lock()
			_, ok := con.locks[key]
			con.lock.Unlock()
			if !ok {
				//如果锁不存在当前的缓存里面，增加到缓存
				timeout := int64(v.Value[0]) | int64(v.Value[1])<<8 |
					int64(v.Value[2])<<16 | int64(v.Value[3])<<24 |
					int64(v.Value[4])<<32 | int64(v.Value[5])<<40 |
					int64(v.Value[6])<<48 | int64(v.Value[7])<<56
				startLockTime := int64(v.Value[8]) | int64(v.Value[9])<<8 |
					int64(v.Value[10])<<16 | int64(v.Value[11])<<24 |
					int64(v.Value[12])<<32 | int64(v.Value[13])<<40 |
					int64(v.Value[14])<<48 | int64(v.Value[15])<<56
				con.addLock(key, timeout, startLockTime)
			}
		}
		time.Sleep(time.Millisecond * 10)
	}
}