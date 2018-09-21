package manager

import "github.com/emicklei/go-restful"

func (m *CronManager) registerlogin(request *restful.Request, w *restful.Response) {
	p, err := ParseForm(request)
	if err != nil {
		m.outJson(w, HttpErrorParseFormFail, err.Error(), nil)
		return
	}
	id, err := m.userModel.Add(p.UserName, p.Password, p.RealName, p.GetPhone())
	if err != nil {
		m.outJson(w, HttpErrorAddUserFail, err.Error(), nil)
		return
	}
	m.outJson(w, HttpSuccess, err.Error(), id)
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
	userInfo.Password = "******"
	m.outJson(w, HttpSuccess, "ok", userInfo)
}
