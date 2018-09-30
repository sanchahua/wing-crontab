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
	"session"
	"github.com/emicklei/go-restful"
	"strings"
	"net/url"
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
	session *session.Session
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

	JumpLoginCode = "<a href=\"/ui/login.html\" id=\"location\">3秒后跳到登录页面，点击去登录</a><script>var s = 3;window.setInterval(function () {s--;document.getElementById(\"location\").innerText=s+\"秒后跳到登录页面，点击去登录\"}, 1000);</script>"
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
		session: session.NewSession(redis),
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
		
		// 日志列表
		http.SetRoute("GET",  "/log/list/{cron_id}/{search_fail}/{page}/{limit}", func(request *restful.Request, response *restful.Response) {
			if !m.sessionValid(request.Request) {
				response.Header().Set("Refresh", "3; url=/ui/login.html")
				response.Write([]byte(JumpLoginCode))
				return
			}
			m.logs(request, response)
		}),

		// 定时任务列表
		http.SetRoute("GET",  "/cron/list", func(request *restful.Request, response *restful.Response) {
			if !m.sessionValid(request.Request) {
				response.Header().Set("Refresh", "3; url=/ui/login.html")
				response.Write([]byte(JumpLoginCode))
				return
			}
			m.cronList(request, response)
		}),

		// 停止定时任务
		http.SetRoute("GET",  "/cron/stop/{id}", func(request *restful.Request, response *restful.Response) {
			if !m.sessionValid(request.Request) {
				response.Header().Set("Refresh", "3; url=/ui/login.html")
				response.Write([]byte(JumpLoginCode))
				return
			}
			m.stopCron(request, response)
		}),

		// 开始定时任务
		http.SetRoute("GET",  "/cron/start/{id}", func(request *restful.Request, response *restful.Response) {
			if !m.sessionValid(request.Request) {
				response.Header().Set("Refresh", "3; url=/ui/login.html")
				response.Write([]byte(JumpLoginCode))
				return
			}
			m.startCron(request, response)
		}),

		// 取消互斥
		http.SetRoute("GET",  "/cron/mutex/false/{id}", func(request *restful.Request, response *restful.Response) {
			if !m.sessionValid(request.Request) {
				response.Header().Set("Refresh", "3; url=/ui/login.html")
				response.Write([]byte(JumpLoginCode))
				return
			}
			m.mutexFalse(request, response)
		}),

		// 设为互斥
		http.SetRoute("GET",  "/cron/mutex/true/{id}", func(request *restful.Request, response *restful.Response) {
			if !m.sessionValid(request.Request) {
				response.Header().Set("Refresh", "3; url=/ui/login.html")
				response.Write([]byte(JumpLoginCode))
				return
			}
			m.mutexTrue(request, response)
		}),

		// 删除定时任务
		http.SetRoute("GET",  "/cron/delete/{id}", func(request *restful.Request, response *restful.Response) {
			if !m.sessionValid(request.Request) {
				response.Header().Set("Refresh", "3; url=/ui/login.html")
				response.Write([]byte(JumpLoginCode))
				return
			}
			m.deleteCron(request, response)
		}),

		// 更新定时任务
		http.SetRoute("POST", "/cron/update/{id}", func(request *restful.Request, response *restful.Response) {
			if !m.sessionValid(request.Request) {
				response.Header().Set("Refresh", "3; url=/ui/login.html")
				response.Write([]byte(JumpLoginCode))
				return
			}
			m.updateCron(request, response)
		}),

		// 增加新的定时任务
		http.SetRoute("POST", "/cron/add", func(request *restful.Request, response *restful.Response) {
			if !m.sessionValid(request.Request) {
				response.Header().Set("Refresh", "3; url=/ui/login.html")
				response.Write([]byte(JumpLoginCode))
				return
			}
			m.addCron(request, response)
		}),

		// 定时任务详情
		http.SetRoute("GET",  "/cron/info/{id}", func(request *restful.Request, response *restful.Response) {
			if !m.sessionValid(request.Request) {
				response.Header().Set("Refresh", "3; url=/ui/login.html")
				response.Write([]byte(JumpLoginCode))
				return
			}
			m.cronInfo(request, response)
		}),
		
		// 首页相关统计信息
		http.SetRoute("GET",  "/index", func(request *restful.Request, response *restful.Response) {
			if !m.sessionValid(request.Request) {
				response.Header().Set("Refresh", "3; url=/ui/login.html")
				response.Write([]byte(JumpLoginCode))
				return
			}
			m.index(request, response)
		}),
		
		// 查询图表，首页使用
		http.SetRoute("GET",  "/charts/{days}", func(request *restful.Request, response *restful.Response) {
			if !m.sessionValid(request.Request) {
				response.Header().Set("Refresh", "3; url=/ui/login.html")
				response.Write([]byte(JumpLoginCode))
				return
			}
			m.charts(request, response)
		}),
		
		// 手动运行定时任务，一般用户测试
		http.SetRoute("GET",  "/cron/run/{id}/{timeout}", func(request *restful.Request, response *restful.Response) {
			if !m.sessionValid(request.Request) {
				response.Header().Set("Refresh", "3; url=/ui/login.html")
				response.Write([]byte(JumpLoginCode))
				return
			}
			m.cronRun(request, response)
		}),

		// 杀死进程
		http.SetRoute("GET",  "/cron/kill/{id}/{process_id}", func(request *restful.Request, response *restful.Response) {
			if !m.sessionValid(request.Request) {
				response.Header().Set("Refresh", "3; url=/ui/login.html")
				response.Write([]byte(JumpLoginCode))
				return
			}
			m.cronKill(request, response)
		}),

		// 查询日志详情
		http.SetRoute("GET",  "/cron/log/detail/{id}", func(request *restful.Request, response *restful.Response) {
			if !m.sessionValid(request.Request) {
				response.Header().Set("Refresh", "3; url=/ui/login.html")
				response.Write([]byte(JumpLoginCode))
				return
			}
			m.cronLogDetail(request, response)
		}),

		// 查询用户列表
		http.SetRoute("GET",  "/users", func(request *restful.Request, response *restful.Response) {
			if !m.sessionValid(request.Request) {
				response.Header().Set("Refresh", "3; url=/ui/login.html")
				response.Write([]byte(JumpLoginCode))
				return
			}
			m.users(request, response)
		}),

		// 通用查询用户信息接口
		http.SetRoute("GET",  "/user/info/{id}", func(request *restful.Request, response *restful.Response) {
			if !m.sessionValid(request.Request) {
				response.Header().Set("Refresh", "3; url=/ui/login.html")
				response.Write([]byte(JumpLoginCode))
				return
			}
			m.userInfo(request, response)
		}),

		// 删除用户接口
		http.SetRoute("POST",  "/user/delete/{id}", func(request *restful.Request, response *restful.Response) {
			if !m.sessionValid(request.Request) {
				response.Header().Set("Refresh", "3; url=/ui/login.html")
				response.Write([]byte(JumpLoginCode))
				return
			}
			m.userDelete(request, response)
		}),

		// 登录api
		http.SetRoute("POST", "/user/login",          m.login),
		// 退出登录api
		http.SetRoute("GET",  "/user/logout",         m.logout),
		// 查询在线用户信息，根据cookie查询
		http.SetRoute("GET",  "/user/session/info",   m.sessionInfo),
		// 更新在线用户信息，个人中心用户更新自己的信息使用
		http.SetRoute("POST", "/user/session/update", m.sessionUpdate),

		// （注册）添加新用户api
		http.SetRoute("POST", "/user/register", func(request *restful.Request, response *restful.Response) {
			if !m.sessionValid(request.Request) {
				response.Header().Set("Refresh", "3; url=/ui/login.html")
				response.Write([]byte(JumpLoginCode))
				return
			}
			m.register(request, response)
		}),

		// 通用更新用户信息接口
		http.SetRoute("POST", "/user/update/{id}", func(request *restful.Request, response *restful.Response) {
			if !m.sessionValid(request.Request) {
				response.Header().Set("Refresh", "3; url=/ui/login.html")
				response.Write([]byte(JumpLoginCode))
				return
			}
			m.update(request, response)
		}),

		// 启用/禁用用户账号接口
		http.SetRoute("POST", "/user/enable/{id}/{enable}", func(request *restful.Request, response *restful.Response) {
			if !m.sessionValid(request.Request) {
				response.Header().Set("Refresh", "3; url=/ui/login.html")
				response.Write([]byte(JumpLoginCode))
				return
			}
			m.enable(request, response)
		}),

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

func (m *CronManager) ui(h shttp.Handler) shttp.Handler {
	prefix:="/ui/"
	return shttp.HandlerFunc(func(w shttp.ResponseWriter, r *shttp.Request) {
		// 只有登录页面和静态资源无需校验登录状态
		if !strings.HasPrefix(r.URL.Path, "/ui/login.html") &&
			!strings.HasPrefix(r.URL.Path, "/ui/static/") {
			if !m.sessionValid(r) {
				w.Header().Set("Refresh", "3; url=/ui/login.html")
				w.Write([]byte(JumpLoginCode))
				return
			}
		}
		if p := strings.TrimPrefix(r.URL.Path, prefix); len(p) < len(r.URL.Path) {
			r2 := new(shttp.Request)
			*r2 = *r
			r2.URL = new(url.URL)
			*r2.URL = *r.URL
			r2.URL.Path = p
			h.ServeHTTP(w, r2)
		} else {
			shttp.NotFound(w, r)
		}
	})
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

