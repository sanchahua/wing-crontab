package manager

import (
	"github.com/emicklei/go-restful"
	"gitlab.xunlei.cn/xllive/common/log"
	"strconv"
	"time"
)

func (m *CronManager) register(request *restful.Request, w *restful.Response) {
	p, err := ParseForm(request)
	if err != nil {
		m.outJson(w, HttpErrorParseFormFail, err.Error(), nil)
		return
	}
	log.Tracef("%+v", *p)
	id, err := m.userModel.Add(p.UserName, p.Password, p.RealName, p.GetPhone())
	if err != nil {
		m.outJson(w, HttpErrorAddUserFail, err.Error(), nil)
		return
	}
	m.outJson(w, HttpSuccess, "ok", id)
}

func (m *CronManager) update(request *restful.Request, w *restful.Response) {
	strUserId := request.PathParameter("id")
	userId, err := strconv.ParseInt(strUserId, 10, 64)
	if err != nil {
		m.outJson(w, HttpErrorUserIdParseFail, err.Error(), nil)
		return
	}
	p, err := ParseForm(request)
	if err != nil {
		m.outJson(w, HttpErrorParseFormFail, err.Error(), nil)
		return
	}
	err = m.userModel.Update(userId, p.UserName, p.Password, p.RealName, p.GetPhone(), p.ISEnable())
	if err != nil {
		m.outJson(w, HttpErrorUpdateUserFail, err.Error(), nil)
		return
	}
	m.outJson(w, HttpSuccess, "ok", nil)
}

func (m *CronManager) enable(request *restful.Request, w *restful.Response) {
	strUserId := request.PathParameter("id")
	userId, err := strconv.ParseInt(strUserId, 10, 64)
	if err != nil {
		m.outJson(w, HttpErrorUserIdParseFail, err.Error(), nil)
		return
	}
	enable := request.PathParameter("enable")
	err = m.userModel.Enable(userId, enable == "1")
	if err != nil {
		m.outJson(w, HttpErrorUserEnableFail, err.Error(), nil)
		return
	}
	m.outJson(w, HttpSuccess, "ok", nil)
}

func (m *CronManager) sessionUpdate(request *restful.Request, w *restful.Response) {
	cookies := m.readCookie(request.Request)
	sessionid, ok := cookies["Session"]
	if !ok {
		m.outJson(w, HttpSessionNotFound, "session not found", nil)
		return
	}
	userId, err := m.session.GetUserId(sessionid)
	if err != nil {
		m.outJson(w, HttpSessionGetUserIdFail, "session get userid fail", nil)
		return
	}
	p, err := ParseForm(request)
	if err != nil {
		m.outJson(w, HttpErrorParseFormFail, err.Error(), nil)
		return
	}
	err = m.userModel.Update(userId, p.UserName, p.Password, p.RealName, p.GetPhone(), p.ISEnable())
	if err != nil {
		m.outJson(w, HttpErrorUpdateUserFail, err.Error(), nil)
		return
	}
	m.outJson(w, HttpSuccess, "ok", nil)
}

func (m *CronManager) sessionInfo(request *restful.Request, w *restful.Response) {
	cookies := m.readCookie(request.Request)
	sessionid, ok := cookies["Session"]
	if !ok {
		m.outJson(w, HttpSessionNotFound, "session not found", nil)
		return
	}
	userId, err := m.session.GetUserId(sessionid)
	if err != nil {
		m.outJson(w, HttpSessionGetUserIdFail, "session get userid fail", nil)
		return
	}
	info, err := m.userModel.GetUserInfo(userId)
	if err != nil {
		m.outJson(w, HttpErrorGetUserInfoFail, err.Error(), nil)
		return
	}
	if info == nil {
		m.outJson(w, HttpErrorUserNotExists, "user does not exists", nil)
		return
	}
	m.outJson(w, HttpSuccess, "ok", info)
}

func (m *CronManager) userInfo(request *restful.Request, w *restful.Response) {
	strUserId := request.PathParameter("id")
	userId, err := strconv.ParseInt(strUserId, 10, 64)
	if err != nil {
		m.outJson(w, HttpErrorUserIdParseFail, err.Error(), nil)
		return
	}
	info, err := m.userModel.GetUserInfo(userId)
	if err != nil {
		m.outJson(w, HttpErrorGetUserInfoFail, err.Error(), nil)
		return
	}
	if info == nil {
		m.outJson(w, HttpErrorUserNotExists, "user does not exists", nil)
		return
	}
	m.outJson(w, HttpSuccess, "ok", info)
}

//users
func (m *CronManager) users(request *restful.Request, w *restful.Response) {
	users, _ := m.userModel.GetUsers()
	m.outJson(w, HttpSuccess, "ok", users)
}

func (m *CronManager) userDelete(request *restful.Request, w *restful.Response) {
	strUserId := request.PathParameter("id")
	userId, err := strconv.ParseInt(strUserId, 10, 64)
	if err != nil {
		m.outJson(w, HttpErrorUserIdParseFail, err.Error(), nil)
		return
	}
	err = m.userModel.Delete(userId)
	if err != nil {
		m.outJson(w, HttpErrorDeleteUserFail, err.Error(), nil)
		return
	}
	m.outJson(w, HttpSuccess, "ok", nil)
}

func (m *CronManager) logout(request *restful.Request, w *restful.Response) {
	cookies := m.readCookie(request.Request)
	sessionid, ok := cookies["Session"]
	if ok {
		err := m.session.Clear(sessionid)
		if err != nil {
			log.Errorf("logout fail, error=[%v]", err)
		}
	}
	//m.outJson(w, HttpSuccess, "ok", nil)
	w.Header().Set("Refresh", "3; url=/ui/login.html")
	w.Write([]byte(JumpLoginCode))
}

func (m *CronManager) login(request *restful.Request, w *restful.Response) {
	p, err := ParseForm(request)
	if err != nil {
		m.outJson(w, HttpErrorParseFormFail, err.Error(), nil)
		return
	}
	userInfo, err := m.userModel.GetUserByUserName(p.UserName)
	if err != nil {
		m.outJson(w, HttpErrorGetUserByUserNameFail, err.Error(), nil)
		return
	}
	if userInfo == nil {
		m.outJson(w, HttpErrorUserNotFound, "用户不存在", nil)
		return
	}
	if userInfo.Password != userInfo.Password {
		m.outJson(w, HttpErrorPasswordError, "密码错误", nil)
		return
	}
	if !userInfo.Enable {
		m.outJson(w, HttpErrorUserDisabled, "用户已被禁用", nil)
		return
	}
	userInfo.Password = "******"
	sessionId, err := m.session.Store(userInfo.Id, time.Second * 60)
	if err != nil {
		m.outJson(w, HttpErrorStoreSessionFail, "创建session失败", nil)
		return
	}
	w.Header().Set("Set-Cookie", "Session="+sessionId+"; Path=/;")
	m.outJson(w, HttpSuccess, "ok", userInfo)
}
