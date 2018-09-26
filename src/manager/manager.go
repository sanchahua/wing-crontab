package manager

import (
	"cron"
	mcron "models/cron"
	"database/sql"
	seelog "gitlab.xunlei.cn/xllive/common/log"
	"library/http"
	modelLog "models/log"
	shttp "net/http"
	_ "statik"
	"github.com/rakyll/statik/fs"
	"time"
	time2 "library/time"
	"models/statistics"
	"github.com/go-redis/redis"
	"fmt"
	"os"
	"encoding/json"
	"service"
	"gitlab.xunlei.cn/xllive/common/log"
	"models/user"
)

type CronManager struct {
	cronController *cron.Controller
	cronModel *mcron.DbCron
	logModel  *modelLog.DbLog
	httpServer *http.HttpServer
	logKeepDay int64
	statisticsModel *statistics.Statistics
	service *service.Service
	serviceId int64
	redis *redis.Client
	watchKey string
	userModel *user.User
}

const (
	EV_ADD = 1
	EV_DELETE = 2
	EV_UPDATE = 3
	EV_START = 4
	EV_STOP = 5
	EV_DISABLE_MUTEX = 6
	EV_ENABLE_MUTEX = 7
	EV_KILL = 8
)

func NewManager(
	service *service.Service,
	redis *redis.Client,
	RedisKeyPrex string, db *sql.DB,
	listen string, logKeepDay int64) *CronManager {
	cronModel := mcron.NewCron(db)
	logModel  := modelLog.NewLog(db)
	statisticsModel := statistics.NewStatistics(db)
	cronController := cron.NewController(redis, RedisKeyPrex, logModel, statisticsModel)
	userModel := user.NewUser(db)

	//这里还需要一个线程，watch定时任务的增删改查，用来改变自身的配置
	name, err := os.Hostname()
	if err != nil {
		seelog.Errorf("%v", err)
		panic(1)
	}
	watchKey := name + "-" + listen

	m := &CronManager{
		cronController:cronController,
		cronModel:cronModel,
		logModel: logModel,
		logKeepDay: logKeepDay,
		statisticsModel: statisticsModel,
		service: service,
		redis: redis,
		watchKey: watchKey,
		userModel: userModel,
	}
	m.init()
	statikFS, err := fs.New()
	if err != nil {
		panic(err)
		return nil
	}

	// restful api 路由
	m.httpServer = http.NewHttpServer(
		listen,
		http.SetRoute("GET",  "/log/list/{cron_id}/{search_fail}/{page}/{limit}", m.logs),
		http.SetRoute("GET",  "/cron/list",                   m.cronList),
		http.SetRoute("GET",  "/cron/stop/{id}",              m.stopCron),
		http.SetRoute("GET",  "/cron/start/{id}",             m.startCron),
		http.SetRoute("GET",  "/cron/mutex/false/{id}",       m.mutexFalse),
		http.SetRoute("GET",  "/cron/mutex/true/{id}",        m.mutexTrue),
		http.SetRoute("GET",  "/cron/delete/{id}",            m.deleteCron),
		http.SetRoute("POST", "/cron/update/{id}",            m.updateCron),
		http.SetRoute("POST", "/cron/add",                    m.addCron),
		http.SetRoute("GET",  "/cron/info/{id}",              m.cronInfo),
		http.SetRoute("GET",  "/index",                       m.index),
		http.SetRoute("GET",  "/charts/{days}",               m.charts),
		http.SetRoute("GET",  "/cron/run/{id}/{timeout}",     m.cronRun),
		http.SetRoute("GET",  "/cron/kill/{id}/{process_id}", m.cronKill),
		http.SetRoute("GET",  "/cron/log/detail/{id}",        m.cronLogDetail),

		http.SetRoute("GET",  "/user/info/{id}",              m.userInfo),
		http.SetRoute("POST",  "/user/delete/{id}",            m.userDelete),
		http.SetRoute("POST", "/user/login",                  m.login),
		http.SetRoute("POST", "/user/register",               m.register),
		http.SetRoute("POST", "/user/update/{id}",            m.update),

		http.SetHandle("/ui/", shttp.StripPrefix("/ui/", shttp.FileServer(statikFS))),
	)
	m.httpServer.Start()
	//cronController.SetAvgMaxData()
	go m.logManager()
	go m.checkDateTime()
	go m.updateAvgMax()
	go m.watchCron()
	return m
}

// 广播通知相关事件
func (m *CronManager) broadcast(ev, id int64, p...int64) {
	// 查询服务列表 逐个push redis队列广播通知数据变化
	services, err := m.service.GetServices()
	if err != nil {
		seelog.Errorf("broadcast m.service.GetServices fail, error=[%v]", err)
		return
	}
	log.Tracef("broadcast ev=[%v], id=[%v], p=[%v], serviceId=[%v]", ev, id, p, m.serviceId)
	for _, sv := range services {
		if sv.ID == m.serviceId {
			continue
		}

		// 这里判断一下服务的有效性
		// 如果失效的就不推送了
		if sv.Status != 1 {
			continue
		}

		var data []byte
		var err error
		if ev == EV_KILL {
			data, err = json.Marshal([]int64{ev, id, p[0]})
		} else {
			data, err = json.Marshal([]int64{ev, id})
		}
		if err != nil {
			seelog.Errorf("broadcast json.Marshal fail, error=[%v]", err)
			continue
		}

		//这里还需要一个线程，watch定时任务的增删改查，用来改变自身的配置
		log.Tracef("push [%v] to [%v]", string(data), sv.Address)
		err = m.redis.RPush(sv.Address, string(data)).Err()
		if err != nil {
			seelog.Errorf("broadcast m.redis.RPush fail, error=[%v]", err)
		}
	}
}

func (m *CronManager) watchCron() {
	//[event, id]
	log.Tracef("start watchCron [%v]", m.watchKey)
	var raw = make([]int64, 0)
	for {
		data, err := m.redis.BRPop(time.Second * 3, m.watchKey).Result()
		if err != nil {
			if err != redis.Nil {
				seelog.Errorf("watchCron redis.BRPop fail, error=[%v]", err)
			}
			continue
		}
		log.Tracef("watchCron data=[%v]", data)
		if len(data) < 2 {
			seelog.Errorf("watchCron data len fail, error=[%v]", err)
			continue
		}
		err = json.Unmarshal([]byte(data[1]), &raw)
		if err != nil {
			seelog.Errorf("watchCron json.Unmarshal fail, error=[%v]", err)
			continue
		}
		if len(raw) < 2 {
			seelog.Errorf("watchCron raw len fail, error=[%v]", err)
			continue
		}
		ev := raw[0]
		id := raw[1]
		switch ev {
		case EV_ADD:
			log.Tracef("watchCron new add id=[%v]", id)
			// 新增定时任务
			info, err := m.cronModel.Get(id)
			if err != nil {
				seelog.Errorf("watchCron EV_ADD m.cronModel.Get fail, id=[%v], error=[%v]", id, err)
			} else {
				m.cronController.Add(info)
			}
		case EV_DELETE:
			log.Tracef("watchCron delete id=[%v]", id)
			// 删除定时任务
			m.cronController.Delete(id)
		case EV_UPDATE:
			log.Tracef("watchCron update id=[%v]", id)
			// 更新定时任务
			info, err := m.cronModel.Get(id)
			if err != nil {
				seelog.Errorf("watchCron EV_UPDATE m.cronModel.Get fail, id=[%v], error=[%v]", id, err)
			} else {
				m.cronController.Delete(id)
				m.cronController.Add(info)
			}
		case EV_START:
			log.Tracef("watchCron start id=[%v]", id)
			m.cronController.Stop(id, false)
		case EV_STOP:
			log.Tracef("watchCron stop id=[%v]", id)
			m.cronController.Stop(id, true)
		case EV_DISABLE_MUTEX:
			log.Tracef("watchCron mutex false id=[%v]", id)
			m.cronController.Mutex(id, false)
		case EV_ENABLE_MUTEX:
			log.Tracef("watchCron mutex true id=[%v]", id)
			m.cronController.Mutex(id, true)
		case EV_KILL:
			log.Tracef("watchCron kill id=[%v]", id)
			if len(raw) == 3 {
				m.cronController.Kill(id, int(raw[2]))
			}
		}
	}

}

func (m *CronManager) SetServiceId(serviceId int64) {
	m.serviceId = serviceId
	m.cronController.SetServiceId(serviceId)
}

func (m *CronManager) updateAvgMax() {
	// 周期性的收集平均运行时长和最大运行时长数据
	for {
		m.cronController.SetAvgMaxData()
		time.Sleep(time.Second * 60)
	}
}

// 系统时间的修改会对系统造成致命错误
// 这里检测时间变化，对cron进行reload操作避免bug
func (m *CronManager) checkDateTime() {
	t := time.Now().Unix()
	for {
		time.Sleep(time.Second)
		d := time.Now().Unix() - t
		t = time.Now().Unix()
		if d > 3 || d < 0 {
			fmt.Fprintf(os.Stderr,"%v", "########################system time is change######################\r\n")
			m.cronController.RestartCron()
		}
	}
}

func (m *CronManager) SetLeader(isLeader bool) {
	m.cronController.SetLeader(isLeader)
}

func (m *CronManager) logManager() {
	logKeepDay := m.logKeepDay
	if logKeepDay < 1 {
		logKeepDay = 1
	}
	// 日志清理操作，每60秒执行一次
	for {
		m.logModel.DeleteByStartTime(time2.TimeFormat(time.Now().Unix()-logKeepDay*86400))
		time.Sleep(time.Second * 60)
	}
}

func (m *CronManager) init() {
	// 启动后，先进性服务注册
	// 然后获取节点数量，当前节点id=节点数量
	// 从数据库查询 定时任务id%节点数量==0的定时任务，载入当前的定时任务管理器


	// 节点数量变化产生的流程
	// 如果获取到节点退出事件
	// id也需要重新生成，注意并发加锁与id唯一性即可
	// 如果当前的节点id == 节点数量
	// 从数据库查询 定时任务id%节点数量==0的定时任务，载入当前的定时任务管理器
	// 如果当前的节点id < 节点数量
	// 从数据库查询 定时任务id%节点数量==节点id 的定时任务，载入当前的定时任务管理器
	// 最后判断，当前的定时任务管理器内的定时任务是否在查询列表内，不在则清除
	// 然后再判断返回列表的定时任务，如果该定时任务已存在当前的定时任务管理器，不予处理
	// 不存在，则加入到当前的定时任务管理器中

	list, err := m.cronModel.GetList()
	if err != nil {
		seelog.Errorf("init fail, m.cronModel.GetList fail, error=[%v]", err)
		return
	}
	for _, data := range list {
		log.Tracef("add cron: %+v", *data)
		m.cronController.Add(data)
	}
}

func (m *CronManager) Start() {
	m.cronController.StartCron()
}

func (m *CronManager) Stop() {
	m.cronController.StartCron()
}

