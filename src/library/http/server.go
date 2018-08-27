package http

import (
	"net/http"
	log "github.com/sirupsen/logrus"
	"github.com/emicklei/go-restful"
)

type HttpServer struct{
	Listen      string   // 监听ip 0.0.0.0
	httpHandler http.Handler
	ws *restful.WebService
	container *restful.Container
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

func SetHandle(pattern string, handler http.Handler) HttpServerOption {
	return func(http *HttpServer) {
		http.container.Handle(pattern, handler)
	}
}

// 初始化，系统自动执行
func NewHttpServer(address string, routes ...HttpServerOption) *HttpServer {
	//dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	//if err != nil {
	//	log.Panicf("获取当前目录错误：%+v", err)
	//}
	//currentPath := strings.Replace(dir, "\\", "/", -1)
	server := &HttpServer{
		Listen:      address,
		//httpHandler: http.FileServer(http.Dir(currentPath + "/web")),
		ws:          new(restful.WebService),
		container:   restful.NewContainer(),
	}
	server.container.Router(restful.CurlyRouter{})
	server.container.Add(server.ws)
	for _, f := range routes {
		f(server)
	}
	return server
}

func (server *HttpServer) Start() {
	go func() {
		httpServer := &http.Server{
			Addr:    server.Listen,
			Handler: server.container,
		}
		log.Fatal(httpServer.ListenAndServe())
	}()
}

func (server *HttpServer) Close() {

}


