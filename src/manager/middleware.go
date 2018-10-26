package manager

import (
	"github.com/emicklei/go-restful"
	"strings"
	"net/url"
	shttp "net/http"
	"models/user"
	"gitlab.xunlei.cn/xllive/common/log"
)

// 路由中间件，主要用于是否登录、权限检验
func (m *CronManager) midLogs(request *restful.Request, response *restful.Response) {
	if !m.sessionValid(request.Request) {
		m.login3(request.Request, response)
		return
	}
	if !m.csrfCheck(request.Request) {
		m.outJson(response, HttpErrorCsrfFail, "csrf fail", nil)
		return
	}
	if m.hasPower(request, PLogList) {
		m.logs(request, response)
	} else {
		m.outJson(response, HttpErrorNoPower, "无权访问/操作", nil)
	}
}

func (m *CronManager) midCronList(request *restful.Request, response *restful.Response) {
	if !m.sessionValid(request.Request) {
		m.login3(request.Request, response)
		return
	}
	if !m.csrfCheck(request.Request) {
		m.outJson(response, HttpErrorCsrfFail, "csrf fail", nil)
		return
	}
	if m.hasPower(request, PCronList) {
		m.cronList(request, response)
	} else {
		m.outJson(response, HttpErrorNoPower, "无权访问/操作", nil)
	}
}

func (m *CronManager) midStopCron(request *restful.Request, response *restful.Response) {
	if !m.sessionValid(request.Request) {
		m.login3(request.Request, response)
		return
	}
	if !m.csrfCheck(request.Request) {
		m.outJson(response, HttpErrorCsrfFail, "csrf fail", nil)
		return
	}
	if m.hasPower(request, PCronStop) {
		m.stopCron(request, response)
	} else {
		m.outJson(response, HttpErrorNoPower, "无权访问/操作", nil)
	}
}

func (m *CronManager) midStartCron(request *restful.Request, response *restful.Response) {
	if !m.sessionValid(request.Request) {
		m.login3(request.Request, response)
		return
	}
	if !m.csrfCheck(request.Request) {
		m.outJson(response, HttpErrorCsrfFail, "csrf fail", nil)
		return
	}
	if m.hasPower(request, PCronStart) {
		m.startCron(request, response)
	} else {
		m.outJson(response, HttpErrorNoPower, "无权访问/操作", nil)
	}
}

func (m *CronManager) midMutexFalse(request *restful.Request, response *restful.Response) {
	if !m.sessionValid(request.Request) {
		m.login3(request.Request, response)
		return
	}
	if !m.csrfCheck(request.Request) {
		m.outJson(response, HttpErrorCsrfFail, "csrf fail", nil)
		return
	}
	if m.hasPower(request, PMutexFalse) {
		m.mutexFalse(request, response)
	} else {
		m.outJson(response, HttpErrorNoPower, "无权访问/操作", nil)
	}
}

func (m *CronManager) midMutexTrue(request *restful.Request, response *restful.Response) {
	if !m.sessionValid(request.Request) {
		m.login3(request.Request, response)
		return
	}
	if !m.csrfCheck(request.Request) {
		m.outJson(response, HttpErrorCsrfFail, "csrf fail", nil)
		return
	}
	if m.hasPower(request, PMutexTrue) {
		m.mutexTrue(request, response)
	} else {
		m.outJson(response, HttpErrorNoPower, "无权访问/操作", nil)
	}
}

func (m *CronManager) midDeleteCron(request *restful.Request, response *restful.Response) {
	if !m.sessionValid(request.Request) {
		m.login3(request.Request, response)
		return
	}
	if !m.csrfCheck(request.Request) {
		m.outJson(response, HttpErrorCsrfFail, "csrf fail", nil)
		return
	}
	if m.hasPower(request, PCronDelete) {
		m.deleteCron(request, response)
	} else {
		m.outJson(response, HttpErrorNoPower, "无权访问/操作", nil)
	}
}

func (m *CronManager) midUpdateCron(request *restful.Request, response *restful.Response) {
	if !m.sessionValid(request.Request) {
		m.login3(request.Request, response)
		return
	}
	if !m.csrfCheck(request.Request) {
		m.outJson(response, HttpErrorCsrfFail, "csrf fail", nil)
		return
	}
	if m.hasPower(request, PCronUpdate) {
		m.updateCron(request, response)
	} else {
		m.outJson(response, HttpErrorNoPower, "无权访问/操作", nil)
	}
}

func (m *CronManager) midAddCron(request *restful.Request, response *restful.Response) {
	if !m.sessionValid(request.Request) {
		m.login3(request.Request, response)
		return
	}
	if !m.csrfCheck(request.Request) {
		m.outJson(response, HttpErrorCsrfFail, "csrf fail", nil)
		return
	}
	if m.hasPower(request, PCronAdd) {
		m.addCron(request, response)
	} else {
		m.outJson(response, HttpErrorNoPower, "无权访问/操作", nil)
	}
}

func (m *CronManager) midCronInfo(request *restful.Request, response *restful.Response) {
	if !m.sessionValid(request.Request) {
		m.login3(request.Request, response)
		return
	}
	if !m.csrfCheck(request.Request) {
		m.outJson(response, HttpErrorCsrfFail, "csrf fail", nil)
		return
	}
	if m.hasPower(request, PCronInfo) {
		m.cronInfo(request, response)
	} else {
		m.outJson(response, HttpErrorNoPower, "无权访问/操作", nil)
	}
}

func (m *CronManager) midIndex(request *restful.Request, response *restful.Response) {
	if !m.sessionValid(request.Request) {
		m.login3(request.Request, response)
		return
	}
	if !m.csrfCheck(request.Request) {
		m.outJson(response, HttpErrorCsrfFail, "csrf fail", nil)
		return
	}
	if m.hasPower(request, PIndex) {
		m.index(request, response)
	} else {
		m.outJson(response, HttpErrorNoPower, "无权访问/操作", nil)
	}
}

func (m *CronManager) midCharts(request *restful.Request, response *restful.Response) {
	if !m.sessionValid(request.Request) {
		m.login3(request.Request, response)
		return
	}
	if !m.csrfCheck(request.Request) {
		m.outJson(response, HttpErrorCsrfFail, "csrf fail", nil)
		return
	}
	if m.hasPower(request, PCharts) {
		m.charts(request, response)
	} else {
		m.outJson(response, HttpErrorNoPower, "无权访问/操作", nil)
	}
}

func (m *CronManager) midAvgRunTimeCharts(request *restful.Request, response *restful.Response) {
	if !m.sessionValid(request.Request) {
		m.login3(request.Request, response)
		return
	}
	if !m.csrfCheck(request.Request) {
		m.outJson(response, HttpErrorCsrfFail, "csrf fail", nil)
		return
	}
	if m.hasPower(request, PAvgRunTimeCharts) {
		m.avgRunTimeCharts(request, response)
	} else {
		m.outJson(response, HttpErrorNoPower, "无权访问/操作", nil)
	}
}

func (m *CronManager) midCronRun(request *restful.Request, response *restful.Response) {
	if !m.sessionValid(request.Request) {
		m.login3(request.Request, response)
		return
	}
	if !m.csrfCheck(request.Request) {
		m.outJson(response, HttpErrorCsrfFail, "csrf fail", nil)
		return
	}
	if m.hasPower(request, PCronRun) {
		m.cronRun(request, response)
	} else {
		m.outJson(response, HttpErrorNoPower, "无权访问/操作", nil)
	}
}

func (m *CronManager) midCronKill(request *restful.Request, response *restful.Response) {
	if !m.sessionValid(request.Request) {
		m.login3(request.Request, response)
		return
	}
	if !m.csrfCheck(request.Request) {
		m.outJson(response, HttpErrorCsrfFail, "csrf fail", nil)
		return
	}
	if m.hasPower(request, PKill) {
		m.cronKill(request, response)
	} else {
		m.outJson(response, HttpErrorNoPower, "无权访问/操作", nil)
	}
}

func (m *CronManager) midCronLogDetail(request *restful.Request, response *restful.Response) {
	if !m.sessionValid(request.Request) {
		m.login3(request.Request, response)
		return
	}
	if !m.csrfCheck(request.Request) {
		m.outJson(response, HttpErrorCsrfFail, "csrf fail", nil)
		return
	}
	if m.hasPower(request, PLogDetail) {
		m.cronLogDetail(request, response)
	} else {
		m.outJson(response, HttpErrorNoPower, "无权访问/操作", nil)
	}
}

func (m *CronManager) midUsers(request *restful.Request, response *restful.Response) {
	if !m.sessionValid(request.Request) {
		m.login3(request.Request, response)
		return
	}
	if !m.csrfCheck(request.Request) {
		m.outJson(response, HttpErrorCsrfFail, "csrf fail", nil)
		return
	}
	if m.hasPower(request, PUsers) {
		m.users(request, response)
	} else {
		m.outJson(response, HttpErrorNoPower, "无权访问/操作", nil)
	}
}

func (m *CronManager) midEnable(request *restful.Request, response *restful.Response) {
	if !m.sessionValid(request.Request) {
		m.login3(request.Request, response)
		return
	}
	if !m.csrfCheck(request.Request) {
		m.outJson(response, HttpErrorCsrfFail, "csrf fail", nil)
		return
	}
	if m.hasPower(request, PUserEnable) {
		m.enable(request, response)
	} else {
		m.outJson(response, HttpErrorNoPower, "无权访问/操作", nil)
	}
}

func (m *CronManager) midAdmin(request *restful.Request, response *restful.Response) {
	if !m.sessionValid(request.Request) {
		m.login3(request.Request, response)
		return
	}
	if !m.csrfCheck(request.Request) {
		m.outJson(response, HttpErrorCsrfFail, "csrf fail", nil)
		return
	}
	if m.hasPower(request, PUserAdmin) {
		m.admin(request, response)
	} else {
		m.outJson(response, HttpErrorNoPower, "无权访问/操作", nil)
	}
}

func (m *CronManager) midServices(request *restful.Request, response *restful.Response) {
	if !m.sessionValid(request.Request) {
		m.login3(request.Request, response)
		return
	}
	if !m.csrfCheck(request.Request) {
		m.outJson(response, HttpErrorCsrfFail, "csrf fail", nil)
		return
	}
	if m.hasPower(request, PServices) {
		m.services(request, response)
	} else {
		m.outJson(response, HttpErrorNoPower, "无权访问/操作", nil)
	}
}

func (m *CronManager) midNodeOffline(request *restful.Request, response *restful.Response) {
	if !m.sessionValid(request.Request) {
		m.login3(request.Request, response)
		return
	}
	if !m.csrfCheck(request.Request) {
		m.outJson(response, HttpErrorCsrfFail, "csrf fail", nil)
		return
	}
	if m.hasPower(request, PNodeOffline) {
		m.nodeOffline(request, response)
	} else {
		m.outJson(response, HttpErrorNoPower, "无权访问/操作", nil)
	}
}

//midNodeOnline
func (m *CronManager) midNodeOnline(request *restful.Request, response *restful.Response) {
	if !m.sessionValid(request.Request) {
		m.login3(request.Request, response)
		return
	}
	if !m.csrfCheck(request.Request) {
		m.outJson(response, HttpErrorCsrfFail, "csrf fail", nil)
		return
	}
	if m.hasPower(request, PNodeOnline) {
		m.nodeOnline(request, response)
	} else {
		m.outJson(response, HttpErrorNoPower, "无权访问/操作", nil)
	}
}

func (m *CronManager) midUpdateUser(request *restful.Request, response *restful.Response) {
	if !m.sessionValid(request.Request) {
		m.login3(request.Request, response)
		return
	}
	if !m.csrfCheck(request.Request) {
		m.outJson(response, HttpErrorCsrfFail, "csrf fail", nil)
		return
	}
	if m.hasPower(request, PUserUpdate) {
		m.update(request, response)
	} else {
		m.outJson(response, HttpErrorNoPower, "无权访问/操作", nil)
	}
}

func (m *CronManager) midRegister(request *restful.Request, response *restful.Response) {
	if !m.sessionValid(request.Request) {
		m.login3(request.Request, response)
		return
	}
	if !m.csrfCheck(request.Request) {
		m.outJson(response, HttpErrorCsrfFail, "csrf fail", nil)
		return
	}
	if m.hasPower(request, PUserRegister) {
		m.register(request, response)
	} else {
		m.outJson(response, HttpErrorNoPower, "无权访问/操作", nil)
	}
}

func (m *CronManager) midUserInfo(request *restful.Request, response *restful.Response) {
	if !m.sessionValid(request.Request) {
		m.login3(request.Request, response)
		return
	}
	if !m.csrfCheck(request.Request) {
		m.outJson(response, HttpErrorCsrfFail, "csrf fail", nil)
		return
	}
	if m.hasPower(request, PUserInfo) {
			m.userInfo(request, response)
	} else {
		m.outJson(response, HttpErrorNoPower, "无权访问/操作", nil)
	}
}

func (m *CronManager) midUserDelete(request *restful.Request, response *restful.Response) {
	if !m.sessionValid(request.Request) {
		m.login3(request.Request, response)
		return
	}
	if !m.csrfCheck(request.Request) {
		m.outJson(response, HttpErrorCsrfFail, "csrf fail", nil)
		return
	}
	if m.hasPower(request, PUserDelete) {
		m.userDelete(request, response)
	} else {
		m.outJson(response, HttpErrorNoPower, "无权访问/操作", nil)
	}
}

// 设置/编辑用户权限
func (m *CronManager) midUserPowers(request *restful.Request, response *restful.Response) {
	if !m.sessionValid(request.Request) {
		m.login3(request.Request, response)
		return
	}
	if !m.csrfCheck(request.Request) {
		m.outJson(response, HttpErrorCsrfFail, "csrf fail", nil)
		return
	}
	if m.hasPower(request, PUserPowers) {
		m.userPowers(request, response)
	} else {
		m.outJson(response, HttpErrorNoPower, "无权访问/操作", nil)
	}
}

func (m *CronManager) hasPower(request *restful.Request, power int64) bool {
	uinfo := m.uinfo(request)
	if uinfo == nil {
		log.Tracef("hasPower uinfo is nil")
		return false
	}
	log.Tracef("hasPower powers=[%v], uinfo.Powers & PUserPowers=[%v], admin=[%v]", uinfo.Powers, uinfo.Powers & power, uinfo.Admin)
	return uinfo.Powers & power > 0 || uinfo.Admin
}

// 静态文件入口，嵌入式
func (m *CronManager) ui(h shttp.Handler) shttp.Handler {
	return shttp.HandlerFunc(func(w shttp.ResponseWriter, r *shttp.Request) {
		// 只有登录页面和静态资源无需校验登录状态
		if !strings.HasPrefix(r.URL.Path, "/ui/login.html") &&
		   !strings.HasPrefix(r.URL.Path, "/ui/static/") &&
           !m.sessionValid(r) {
			  m.login3(r, w)
			  return
		}
		if p := strings.TrimPrefix(r.URL.Path, "/ui/"); len(p) < len(r.URL.Path) {
			r2         := new(shttp.Request)
			*r2         = *r
			r2.URL      = new(url.URL)
			*r2.URL     = *r.URL
			r2.URL.Path = p
			h.ServeHTTP(w, r2)
		} else {
			shttp.NotFound(w, r)
		}
	})
}

func (m *CronManager) uinfo(request *restful.Request) *user.Entity {
	cookies := m.readCookie(request.Request)
	sessionid, ok := cookies["Session"]
	if !ok {
		return nil
	}
	userId, err := m.session.GetUserId(sessionid)
	if err != nil {
		return nil
	}
	uinfo, err := m.userModel.GetUserInfo(userId)
	if err != nil || uinfo == nil {
		return nil
	}
	return uinfo
}

func (m *CronManager) login3(r *shttp.Request, response shttp.ResponseWriter) {
	if "ajax" == r.Header.Get("Client") {
		m.outJson(response, HttpErrorNeedLogin, "need login", nil)
	} else {
		response.Header().Set("Refresh", "3; url=/ui/login.html")
		response.Write([]byte(JumpLoginCode))
	}
}