package manager

import (
	"cron"
	mcron "models/cron"
	"database/sql"
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
	"session"
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
	//watchKey string
	userModel *user.User
	session *session.Session

	powers Powers//map[int64]string
	//leader int64
}
const (
	EV_ADD           = 1
	EV_DELETE        = 2
	EV_UPDATE        = 3
	EV_START         = 4
	EV_STOP          = 5
	EV_DISABLE_MUTEX = 6
	EV_ENABLE_MUTEX  = 7
	EV_KILL          = 8
	EV_OFFLINE       = 9
	EV_LEADER        = 10
	JumpLoginCode = "<a href=\"/ui/login.html\" id=\"location\">3秒后跳到登录页面，点击去登录</a><script>var s = 3;window.setInterval(function () {s--;document.getElementById(\"location\").innerText=s+\"秒后跳到登录页面，点击去登录\"}, 1000);</script>"
)

func NewManager(
	service *service.Service,
	redis *redis.Client,
	RedisKeyPrex string, db *sql.DB,
	listen string, logKeepDay int64) *CronManager {

	cronModel       := mcron.NewCron(db)
	logModel        := modelLog.NewLog(db)
	statisticsModel := statistics.NewStatistics(db)
	userModel       := user.NewUser(db)
	cronController  := cron.NewController(service, redis, RedisKeyPrex, logModel, statisticsModel, userModel)


	m := &CronManager{
		cronController:cronController,
		cronModel:cronModel,
		logModel: logModel,
		logKeepDay: logKeepDay,
		statisticsModel: statisticsModel,
		service: service,
		redis: redis,
		//watchKey: watchKey,
		userModel: userModel,
		session: session.NewSession(redis),
		powers: make(Powers, 0),//map[int64]string),
	}
	m.init()
	m.powersInit()
	statikFS, err := fs.New()
	if err != nil {
		panic(err)
		return nil
	}

	// restful api 路由
	m.httpServer = http.NewHttpServer(
		listen,
		// 日志列表
		http.SetRoute("GET",  "/log/list/{cron_id}/{search_fail}/{page}/{limit}", m.midLogs),
		// 定时任务列表
		http.SetRoute("GET",  "/cron/list",             m.midCronList),
		// 停止定时任务
		http.SetRoute("GET",  "/cron/stop/{id}",        m.midStopCron),
		// 开始定时任务
		http.SetRoute("GET",  "/cron/start/{id}",       m.midStartCron),
		// 取消互斥
		http.SetRoute("GET",  "/cron/mutex/false/{id}", m.midMutexFalse),
		// 设为互斥
		http.SetRoute("GET",  "/cron/mutex/true/{id}",  m.midMutexTrue),
		// 删除定时任务
		http.SetRoute("GET",  "/cron/delete/{id}",      m.midDeleteCron),
		// 更新定时任务
		http.SetRoute("POST", "/cron/update/{id}",      m.midUpdateCron),
		// 增加新的定时任务
		http.SetRoute("POST", "/cron/add",              m.midAddCron),
		// 定时任务详情
		http.SetRoute("GET",  "/cron/info/{id}",        m.midCronInfo),
		// 首页相关统计信息
		http.SetRoute("GET",  "/index",                 m.midIndex),
		// 查询图表，首页使用
		http.SetRoute("GET",  "/charts/{days}",         m.midCharts),
		//charts/avg_run_time/'+that.days+
		http.SetRoute("GET",  "/charts/avg_run_time/{days}",         m.midAvgRunTimeCharts),
		// 手动运行定时任务，一般用户测试
		http.SetRoute("GET",  "/cron/run/{id}/{timeout}", m.midCronRun),
		// 杀死进程
		http.SetRoute("GET",  "/cron/kill/{id}/{process_id}", m.midCronKill),
		// 查询日志详情
		http.SetRoute("GET",  "/cron/log/detail/{id}", m.midCronLogDetail),
		// 查询用户列表
		http.SetRoute("GET",  "/users",              m.midUsers),
		// 通用查询用户信息接口
		http.SetRoute("GET",  "/user/info/{id}",     m.midUserInfo),
		// 删除用户接口
		http.SetRoute("POST",  "/user/delete/{id}",  m.midUserDelete),
		http.SetRoute("POST",  "/user/powers/{id}",  m.midUserPowers),
		// 登录api
		http.SetRoute("POST", "/user/login",          m.login),
		// 退出登录api
		http.SetRoute("GET",  "/user/logout",         m.logout),
		// 查询在线用户信息，根据cookie查询
		http.SetRoute("GET",  "/user/session/info",   m.sessionInfo),
		// 更新在线用户信息，个人中心用户更新自己的信息使用
		http.SetRoute("POST", "/user/session/update", m.sessionUpdate),
		// （注册）添加新用户api
		http.SetRoute("POST", "/user/register",       m.midRegister),
		// 通用更新用户信息接口
		http.SetRoute("POST", "/user/update/{id}",    m.midUpdateUser),
		// 启用/禁用用户账号接口
		http.SetRoute("POST", "/user/enable/{id}/{enable}", m.midEnable),
		http.SetRoute("POST", "/user/admin/{id}/{admin}", m.midAdmin),
		http.SetRoute("GET",  "/powers/{id}",   m.powersList),
		http.SetRoute("GET",  "/page/power/check", m.pagePowerCheck),

		http.SetRoute("GET",  "/services", m.midServices),
		http.SetRoute("POST",  "/services/offline/{id}", m.midNodeOffline),
		http.SetRoute("POST",  "/services/online/{id}", m.midNodeOnline),

		http.SetHandle("/ui/", m.ui(shttp.FileServer(statikFS))),
	)
	m.httpServer.Start()
	//cronController.SetAvgMaxData()
	// 负责清理过期日志
	go m.logManager()
	// 检测系统是否有发生修改时间行为
	go m.checkDateTime()
	// 更新平均时长和最大运行时长
	go m.updateAvgMax()
	// 检测同步事件
	go m.watchCron()
	return m
}

// 广播通知相关事件
func (m *CronManager) broadcast(ev, id int64, p...int64) {
	// 查询服务列表 逐个push redis队列广播通知数据变化
	services, err := m.service.GetServices()
	if err != nil {
		log.Errorf("broadcast m.service.GetServices fail, error=[%v]", err)
		return
	}
	log.Tracef("broadcast ev=[%v], id=[%v], p=[%v], serviceId=[%v]", ev, id, p, m.serviceId)
	for _, sv := range services {
		if sv.ID == m.serviceId {
			continue
		}

		// 这里判断一下服务的有效性
		// 如果失效的就不推送了
		// only not push offline node
		if sv.Offline == 1 {
			continue
		}

		var data []byte
		var err error

		sendData := make([]int64, 0)
		sendData = append(sendData, ev)
		sendData = append(sendData, id)
		sendData = append(sendData, p...)


		//if ev == EV_KILL {
			data, err = json.Marshal(sendData)
		//} else {
		//	data, err = json.Marshal([]int64{ev, id})
		//}
		if err != nil {
			log.Errorf("broadcast json.Marshal fail, error=[%v]", err)
			continue
		}

		watch := fmt.Sprintf("wing-crontab/watch/event/%v", sv.ID)
		//这里还需要一个线程，watch定时任务的增删改查，用来改变自身的配置
		log.Tracef("push [%v] to [%v]", string(data), watch)
		err = m.redis.RPush(watch, string(data)).Err()
		if err != nil {
			log.Errorf("broadcast m.redis.RPush fail, error=[%v]", err)
		}
	}
}

func (m *CronManager) watchCron() {
	//[event, id]
	watch := fmt.Sprintf("wing-crontab/watch/event/%v", m.service.ID)
	log.Tracef("start watchCron [%v]", watch)
	var raw = make([]int64, 0)
	for {
		data, err := m.redis.BRPop(time.Second * 3, watch).Result()
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
		ev := raw[0]
		id := raw[1]
		switch ev {
		case EV_OFFLINE:
			log.Infof("receive event offline, serviceid=[%v], offline=[%v]", id, raw[2] == 1)
			m.service.SetOffline(id, raw[2] == 1)
		//case EV_LEADER:
		//	m.service.OnKeepLeader(id)
		case EV_ADD:
			log.Tracef("watchCron new add id=[%v]", id)
			// 新增定时任务
			info, err := m.cronModel.Get(id)
			if err != nil {
				log.Errorf("watchCron EV_ADD m.cronModel.Get fail, id=[%v], error=[%v]", id, err)
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
				log.Errorf("watchCron EV_UPDATE m.cronModel.Get fail, id=[%v], error=[%v]", id, err)
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
		if m.service.IsOffline() {
			time.Sleep(time.Second)
			continue
		}
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

//func (m *CronManager) SetLeader(isLeader bool) {
//	//atomic.StoreInt64(&m.leader, 1)
//	//m.cronController.SetLeader(isLeader)
//}

func (m *CronManager) logManager() {
	logKeepDay := m.logKeepDay
	if logKeepDay < 1 {
		logKeepDay = 1
	}
	// 日志清理操作，每60秒执行一次
	for {
		// only leader do this
		if !m.service.IsLeader() || m.service.IsOffline() {
			time.Sleep(time.Second * 60)
			continue
		}
		m.logModel.DeleteByStartTime(time2.TimeFormat(time.Now().Unix()-logKeepDay*86400))
		time.Sleep(time.Second * 60)
	}
}

func (m *CronManager) init() {
	list, err := m.cronModel.GetList()
	if err != nil {
		log.Errorf("init fail, m.cronModel.GetList fail, error=[%v]", err)
		return
	}
	for _, data := range list {
		log.Tracef("add cron: %+v", *data)
		m.cronController.Add(data)
	}
	m.cronController.Ready()
}

func (m *CronManager) Start() {
	m.cronController.StartCron()
}

func (m *CronManager) Stop() {
	m.cronController.StartCron()
}

