#!/bin/bash

## 更新proto的协议定义文件，重新生成代码

function build() {
    input=$1
    protoc ${input} --go_out=plugins=grpc:./
    if [ $? -ne 0 ]
    then
        echo -ne "\033[47;31m [ERROR] \033[0m"
        echo "Build ${input} fail."
        return 1
    fi
    echo -ne "\033[47;34m [PASSED] \033[0m"

    return 0
}

echo "Updating submodule ..."
git submodule init
git submodule update --remote

mkdir -pv proto_gen/xlsoa/core/
mkdir -pv proto_gen/xlsoa/example/

build "proto/xlsoa/core/certificate.proto" && mv -v proto/xlsoa/core/certificate.pb.go proto_gen/xlsoa/core/
build "proto/xlsoa/example/*.proto" && mv -v proto/xlsoa/example/*.pb.go proto_gen/xlsoa/example/
