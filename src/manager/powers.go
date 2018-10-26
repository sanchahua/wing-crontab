package manager

import (
	"github.com/emicklei/go-restful"
	"fmt"
	"strconv"
)

// 以下值可能会因调整顺序而发生改变
// 所以所有的判断均不能写死某某值来识别
const (
	PLogList = 1 << iota
	PCronList
	PCronStop
	PCronStart
	PMutexFalse
	PMutexTrue
	PCronDelete
	PCronUpdate
	PCronAdd
	PCronInfo
	PIndex
	PCharts
	PAvgRunTimeCharts
	PCronRun
	PKill
	PLogDetail
	PUsers
	PUserInfo
	PUserDelete
	PUserRegister
	PUserUpdate
	PUserEnable
	PUserAdmin
	PUserPowers
	PServices
	PNodeOffline
	PNodeOnline
	//这几个人人都应该具备的权限
	//PLogin
	//PLogout
	//PSessionInfo
	//PSessionUpdate
	//PUi
// ######## 页面权限 ########
	// 增加定时任务页面
	P_PageCronAdd
	// 定时任务管理页面
	P_PageCronList
	// 定时任务编辑页面
	P_PageCronEdit
	// 定时任务运行日志页面
	P_PageCronLogs
	// 定时任务日志详情页面
	P_PageCronLogDetail
	P_PageUsers
	// 添加用户页面
	P_PageUserAdd
	// 编辑用户页面
	P_PageUserEdit
	// 用户权限分配页面
	P_PageUserPowers
)
type Power struct {
	Id int64 `json:"id"`
	Name string `json:"name"`
	Checked bool `json:"checked"`
}

type Powers []*Power

func (a Powers) Len() int {
	return len(a)
}
func (a Powers) Swap(i, j int){
	a[i], a[j] = a[j], a[i]
}
func (a Powers) Less(i, j int) bool {
	return a[j].Id > a[i].Id
}

func (m *CronManager) powersInit() {
	m.powers = append(m.powers, &Power{
		Id: PLogList,
		Name: "日志列表（只读，可开放）",
	})
	m.powers = append(m.powers, &Power{
		Id: PCronList,
		Name: "定时任务列表（只读，可开放）",
	})

	m.powers = append(m.powers, &Power{
		Id: PCronStop,
		Name: "停止定时任务（写操作，谨慎开放）",
	})

	m.powers = append(m.powers, &Power{
		Id: PCronStart,
		Name: "开始定时任务（写操作，谨慎开放）",
	})

	m.powers = append(m.powers, &Power{
		Id: PMutexFalse,
		Name: "取消互斥（写操作，谨慎开放）",
	})

	m.powers = append(m.powers, &Power{
		Id: PMutexTrue,
		Name: "设为互斥（写操作，谨慎开放）",
	})

	m.powers = append(m.powers, &Power{
		Id: PCronDelete,
		Name: "删除定时任务（写操作，严格开放）",
	})

	m.powers = append(m.powers, &Power{
		Id: PCronUpdate,
		Name: "更新定时任务（写操作，严格开放）",
	})

	m.powers = append(m.powers, &Power{
		Id: PCronAdd,
		Name: "添加定时任务（写操作，严格开放）",
	})

	m.powers = append(m.powers, &Power{
		Id: PCronInfo,
		Name: "查询定时任务详细信息（只读，可开放）",
	})

	m.powers = append(m.powers, &Power{
		Id: PIndex,
		Name: "首页统计信息接口（只读，一般必须开放）",
	})

	m.powers = append(m.powers, &Power{
		Id: PCharts,
		Name: "首页图表接口（只读，一般必须开放）",
	})

	//PAvgRunTimeCharts
	m.powers = append(m.powers, &Power{
		Id: PAvgRunTimeCharts,
		Name: "首页平均运行时长图表接口（只读，一般必须开放）",
	})

	m.powers = append(m.powers, &Power{
		Id: PCronRun,
		Name: "运行定时任务（具有风险性，严格开放）",
	})

	m.powers = append(m.powers, &Power{
		Id: PKill,
		Name: "杀死正在运行的定时任务进程（具有风险性，严格开放）",
	})

	m.powers = append(m.powers, &Power{
		Id: PLogDetail,
		Name: "查询日志详细信息（只读，可开放）",
	})

	m.powers = append(m.powers, &Power{
		Id: PUsers,
		Name: "查询用户列表接口（只读，谨慎开放）",
	})

	m.powers = append(m.powers, &Power{
		Id: PUserInfo,
		Name: "用户信息查询接口（只读，谨慎开放）",
	})

	m.powers = append(m.powers, &Power{
		Id: PUserDelete,
		Name: "用户删除接口（写操作，严格开放）",
	})

	m.powers = append(m.powers, &Power{
		Id: PUserRegister,
		Name: "添加新用户（写操作，严格开放）",
	})

	m.powers = append(m.powers, &Power{
		Id: PUserUpdate,
		Name: "更新用户信息（写操作，严格开放）",
	})

	m.powers = append(m.powers, &Power{
		Id: PUserEnable,
		Name: "禁用/启用用户账号（写操作，谨慎开放）",
	})

	m.powers = append(m.powers, &Power{
		Id: PUserAdmin,
		Name: "设置/取消管理员（写操作，严格开放）",
	})

	m.powers = append(m.powers, &Power{
		Id: PUserPowers,
		Name: "设置/编辑用户权限（写操作，严格开放）",
	})

	m.powers = append(m.powers, &Power{
		Id: PServices,
		Name: "服务器集群节点列表（只读操作，可开放）",
	})

	m.powers = append(m.powers, &Power{
		Id: PNodeOffline,
		Name: "下线服务器",
	})

	m.powers = append(m.powers, &Power{
		Id: PNodeOnline,
		Name: "上线服务器",
	})

	m.powers = append(m.powers, &Power{
		Id: P_PageCronAdd,
		Name: "增加定时任务页面",
	})

	m.powers = append(m.powers, &Power{
		Id: P_PageCronList,
		Name: "定时任务管理页面",
	})

	m.powers = append(m.powers, &Power{
		Id: P_PageCronEdit,
		Name: "定时任务编辑页面",
	})

	m.powers = append(m.powers, &Power{
		Id: P_PageCronLogs,
		Name: "定时任务运行日志页面",
	})

	m.powers = append(m.powers, &Power{
		Id: P_PageCronLogDetail,
		Name: "定时任务日志详情页面",
	})

	//P_PageUsers
	m.powers = append(m.powers, &Power{
		Id: P_PageUsers,
		Name: "用户管理页面",
	})
	m.powers = append(m.powers, &Power{
		Id: P_PageUserAdd,
		Name: "添加用户页面",
	})

	m.powers = append(m.powers, &Power{
		Id: P_PageUserEdit,
		Name: "编辑用户页面",
	})

	m.powers = append(m.powers, &Power{
		Id: P_PageUserPowers,
		Name: "用户权限分配页面",
	})
	//sort.Sort(m.powers)
}

func (m *CronManager) powersList(request *restful.Request, w *restful.Response) {

	strUserId := request.PathParameter("id")
	userId, err := strconv.ParseInt(strUserId, 10, 64)
	if err != nil {
		m.outJson(w, HttpErrorUserIdParseFail, err.Error(), nil)
		return
	}

	userinfo, err := m.userModel.GetUserInfo(userId)
	if err != nil {
		m.outJson(w, HttpErrorGetUserInfoFail, err.Error(), nil)
		return
	}

	for _, v := range m.powers {
		fmt.Printf("userid=[%v--%v], power id=[%v], checked=[%v]\r\n", userinfo.Id, userinfo.Powers, v.Id, v.Id & userinfo.Powers > 0)
		v.Checked = v.Id & userinfo.Powers > 0
	}

	m.outJson(w, HttpSuccess, "ok", m.powers)
}

// 页面权限判断，仅仅用来控制展示或隐藏某些标签
func (m *CronManager) pagePowerCheck(request *restful.Request, w *restful.Response) {
	id := request.QueryParameter("id")
	switch id {
		// 增加定时任务页面
	case "P_PageCronAdd":
		if m.hasPower(request, P_PageCronAdd) {
			m.outJson(w, HttpSuccess, "ok", nil)
		} else {
			m.outJson(w, HttpErrorNoPower, "no power for visit", nil)
		}
		// 定时任务管理页面
	case "P_PageCronList":
		if m.hasPower(request, P_PageCronList) {
			m.outJson(w, HttpSuccess, "ok", nil)
		} else {
			m.outJson(w, HttpErrorNoPower, "no power for visit", nil)
		}
		// 定时任务编辑页面
	case "P_PageCronEdit":
		if m.hasPower(request, P_PageCronEdit) {
			m.outJson(w, HttpSuccess, "ok", nil)
		} else {
			m.outJson(w, HttpErrorNoPower, "no power for visit", nil)
		}
		// 定时任务运行日志页面
	case "P_PageCronLogs":
		if m.hasPower(request, P_PageCronLogs) {
			m.outJson(w, HttpSuccess, "ok", nil)
		} else {
			m.outJson(w, HttpErrorNoPower, "no power for visit", nil)
		}
		// 定时任务日志详情页面
	case "P_PageCronLogDetail":
		if m.hasPower(request, P_PageCronLogDetail) {
			m.outJson(w, HttpSuccess, "ok", nil)
		} else {
			m.outJson(w, HttpErrorNoPower, "no power for visit", nil)
		}
	case "P_PageUsers":
		if m.hasPower(request, P_PageUsers) {
			m.outJson(w, HttpSuccess, "ok", nil)
		} else {
			m.outJson(w, HttpErrorNoPower, "no power for visit", nil)
		}
		// 添加用户页面
	case "P_PageUserAdd":
		if m.hasPower(request, P_PageUserAdd) {
			m.outJson(w, HttpSuccess, "ok", nil)
		} else {
			m.outJson(w, HttpErrorNoPower, "no power for visit", nil)
		}
		// 编辑用户页面
	case "P_PageUserEdit":
		if m.hasPower(request, P_PageUserEdit) {
			m.outJson(w, HttpSuccess, "ok", nil)
		} else {
			m.outJson(w, HttpErrorNoPower, "no power for visit", nil)
		}
		// 用户权限分配页面
	case "P_PageUserPowers":
		if m.hasPower(request, P_PageUserPowers) {
			m.outJson(w, HttpSuccess, "ok", nil)
		} else {
			m.outJson(w, HttpErrorNoPower, "no power for visit", nil)
		}
	default:
		m.outJson(w, HttpErrorNoPower, "no power for visit", nil)
	}
}