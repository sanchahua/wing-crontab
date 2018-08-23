package http

import (
	"net/http"
	log "github.com/sirupsen/logrus"
	"path/filepath"
	"os"
	"strings"
	"github.com/emicklei/go-restful"
)

type HttpServer struct{
	Listen      string   // 监听ip 0.0.0.0
	httpHandler http.Handler
	ws *restful.WebService
}

//type RouteFunc func(request *restful.Request, w *restful.Response)
type HttpServerOption func(http *HttpServer)
func SetRoute(m string, r string, f restful.RouteFunction) HttpServerOption {
	return func(http *HttpServer) {
		switch (m) {
		case "POST":
			http.ws.Route(http.ws.POST(r).To(f))
		case "GET":
			http.ws.Route(http.ws.GET(r).To(f))
		}
	}
}

// 初始化，系统自动执行
func NewHttpServer(address string, routes ...HttpServerOption) *HttpServer {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Panicf("获取当前目录错误：%+v", err)
	}
	currentPath := strings.Replace(dir, "\\", "/", -1)
	//config, err := app.GetConfig()
	//if err != nil {
	//	log.Panicf("%+v", err)
	//}
	server := &HttpServer{
		Listen : address,//config.HttpListen,
		httpHandler : http.FileServer(http.Dir(currentPath + "/web")),
		ws:new(restful.WebService),
	}
	for _, f := range routes {
		f(server)
	}
	return server
}

func (server *HttpServer) Start() {
	//go func() {
	//	log.Infof("http服务器启动...")
	//	log.Infof("http服务监听: %s", server.Listen)
	//	http.HandleFunc("/", server.onRequest)
	//	http.HandleFunc("/cron/add", server.add)
	//	http.HandleFunc("/cron/list", server.list) 		        // 查询数据库的定时任务列表
	//	//http.HandleFunc("/cron/list/running", server.running)		// 查询正在运行的定时任务列表
	//	http.HandleFunc("/cron/lock", server.lock) //锁定所有的定时任务，用来测试死锁
	//	err := http.ListenAndServe(server.Listen, nil)
	//	if err != nil {
	//		log.Fatalf("http服务启动失败: %v", err)
	//	}
	//}()
	go func() {
		wsContainer := restful.NewContainer()
		wsContainer.Router(restful.CurlyRouter{})
		//u := UserResource{map[string]User{}}
		//u.Register(wsContainer)

		//ws := new(restful.WebService)
		//ws.
		//	Path("/users").
		//	Consumes(restful.MIME_XML, restful.MIME_JSON).
		//	Produces(restful.MIME_JSON, restful.MIME_XML) // you can specify this per route as well

		//ws.Route(ws.GET("/cron/list").To(server.list))
		//ws.Route(ws.GET("/cron/stop/{id}").To(server.stop))
		//ws.Route(ws.GET("/cron/start/{id}").To(server.start))
		//ws.Route(ws.GET("/cron/delete/{id}").To(server.delete))
		//ws.Route(ws.POST("/cron/update").To(server.update))
		//ws.Route(ws.POST("/cron/add").To(server.add))
		//ws.Route(ws.GET("/cron/unlock/{id}").To(server.unlock))
		//ws.Route(ws.GET("/cron/lock/{id}").To(server.lock))

		//ws.Route(ws.POST("").To(u.updateUser))
		//ws.Route(ws.PUT("/{user-id}").To(u.createUser))
		//ws.Route(ws.DELETE("/{user-id}").To(u.removeUser))

		wsContainer.Add(server.ws)

		log.Printf("start http server: %s", server.Listen)
		httpserver := &http.Server{Addr: server.Listen, Handler: wsContainer}
		log.Fatal(httpserver.ListenAndServe())
	}()
}

func (server *HttpServer) Close() {

}


