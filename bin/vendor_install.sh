#!/usr/bin/env bash
current_path=$(cd `dirname $0`; pwd)
root_path=$(dirname $current_path)
vendor_path=$root_path"/vendor"

##添加当前目录和当前目录下的vendor目录到GOPATH环境变量
export GOPATH="$vendor_path:$root_path"
if [ ! -d "$vendor_path" ]; then
 mkdir "$vendor_path"
 mkdir "$vendor_path/src"
fi
echo "installing... github.com/go-sql-driver/mysql"
go get github.com/go-sql-driver/mysql
echo "installing... github.com/siddontang/go-mysql/canal"
go get github.com/siddontang/go-mysql/canal
echo "installing... github.com/siddontang/go-mysql/replication"
go get github.com/siddontang/go-mysql/replication
echo "installing... github.com/siddontang/go-mysql/mysql"
go get github.com/siddontang/go-mysql/mysql
echo "installing... github.com/BurntSushi/toml"
go get github.com/BurntSushi/toml
echo "installing... go-martini/martini"
go get github.com/go-martini/martini
echo "installing... gorilla/websocket"
go get github.com/gorilla/websocket
echo "installing... github.com/axgle/mahonia"
go get github.com/axgle/mahonia
echo "installing... github.com/hashicorp/consul"
go get github.com/hashicorp/consul
echo "installing... github.com/sirupsen/logrus"
go get github.com/sirupsen/logrus
echo "installing... github.com/sevlyar/go-daemon"
go get github.com/sevlyar/go-daemon
echo "installing... github.com/go-redis/redis"
go get github.com/go-redis/redis
echo "installing... github.com/Shopify/sarama"
go get github.com/Shopify/sarama
echo "installing... github.com/orcaman/concurrent-map"
go get github.com/orcaman/concurrent-map
echo "installing... gopkg.in/robfig/cron.v2"
go get gopkg.in/robfig/cron.v2
echo "installing... github.com/jilieryuyi/wing-go"
go get github.com/jilieryuyi/wing-go
echo "installing... go get github.com/emicklei/go-restful"
go get github.com/emicklei/go-restful
go get github.com/cihub/seelog
go get github.com/go-yaml/yaml
go get github.com/golang/protobuf/jsonpb
go get github.com/json-iterator/go
go get github.com/json-iterator/go/extra

find $vendor_path -name '*.git*' | xargs rm -rf
##cp -rf $vendor_path/src/* $vendor_path
##cp -rf $root_path/src/library $root_path/vendor/

echo "install complete"
