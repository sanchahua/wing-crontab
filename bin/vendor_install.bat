@echo off
set current_path=%cd%
set vendor_path=%current_path%\vendor

::添加当前目录和当前目录下的vendor目录到GOPATH环境变量
set GOPATH=%vendor_path%;%current_path%

if not exist %vendor_path% (
 md %vendor_path%
 md %vendor_path%\src
)

echo installing... github.com/go-sql-driver/mysql
call go get github.com/go-sql-driver/mysql

echo installing... github.com/siddontang/go-mysql/canal
call go get github.com/siddontang/go-mysql/canal

echo installing... github.com/siddontang/go-mysql/replication
call go get github.com/siddontang/go-mysql/replication

echo installing... github.com/siddontang/go-mysql/mysql
call go get github.com/siddontang/go-mysql/mysql

echo installing... github.com/BurntSushi/toml
call go get github.com/BurntSushi/toml

echo installing... go-martini/martini
call go get github.com/go-martini/martini

echo installing... gorilla/websocket
call go get github.com/gorilla/websocket
echo installing... github.com/axgle/mahonia
call go get github.com/axgle/mahonia
echo installing... github.com/hashicorp/consul
call go get github.com/hashicorp/consul
echo installing... github.com/sirupsen/logrus
call go get github.com/sirupsen/logrus
echo installing... github.com/sevlyar/go-daemon
call go get github.com/sevlyar/go-daemon
echo installing... github.com/go-redis/redis
call go get github.com/go-redis/redis
echo installing... github.com/Shopify/sarama
call go get github.com/Shopify/sarama
echo installing... github.com/orcaman/concurrent-map
call go get github.com/orcaman/concurrent-map
echo installing... gopkg.in/robfig/cron.v2
call go get gopkg.in/robfig/cron.v2
echo installing... github.com/jilieryuyi/wing-go
call go get github.com/jilieryuyi/wing-go
echo installing... go get github.com/emicklei/go-restful
call go get github.com/emicklei/go-restful
call go get github.com/cihub/seelog
call go get github.com/go-yaml/yaml
call go get github.com/golang/protobuf/jsonpb
call go get github.com/json-iterator/go
call go get github.com/json-iterator/go/extra
call go get github.com/rakyll/statik/fs
call go get github.com/parnurzeal/gorequest
call go get github.com/huandu/goroutine
call go get github.com/bsm/go-guid

::xcopy  %vendor_path%\src\*.* %vendor_path% /s /e /y /q

echo install complete
