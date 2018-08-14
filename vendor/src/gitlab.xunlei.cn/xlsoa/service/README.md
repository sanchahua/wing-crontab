
## Overview
3.0版本开始Sdk的设计原则上不侵入用户使用gRPC的方式，sdk只提供必要的组件和gRPC配套的Options。用户只需要按照gRPC本身的方式，并且按照Sdk的约定把相关Options制定到gRPC即可。

## 1. Examples

### Greeter
演示了一个简单的Client对Server调用的例子

```
     --------------------------------           (1)          --------------------------------
     | xlsoa.example.greeter.client |    ----------------->  |     xlsoa.example.greeter    |
     |                              |    <-----------------  |                              |
     --------------------------------           (2)          --------------------------------
```

### Candy
演示了一个中间代理服务器，对Client端的请求，转发到后端Server


```
     --------------------------------            (1)         --------------------------------         (2)          --------------------------------
     | xlsoa.example.candy.client   |    ----------------->  |     xlsoa.example.candy1     |  ----------------->  |    xlsoa.example.candy2      |
     |                              |    <-----------------  |                              |  <-----------------  |                              |
     --------------------------------            (4)         --------------------------------         (3)          --------------------------------
```

## 2. 组件
### Environment
每一个进程只需要初始化一个Environment对象，Environment负责加载并维护全局的上下文数据，例如`creds.json`、`soa.yml`等。

```go
env := NewEnvironment()
```

### Server Context
一个gRPC Server对应一个Server Context对象，负责维护Server的状态数据以及提供gRPC拦截器等。例如，对Client端的AccessToken鉴权。

```go
ctx := NewServerContext(env)
```

配合gRPC使用，需要在`grpc.NewServer()`指定以下Options:

```
grpc.UnaryInterceptor(ctx.GrpcUnaryServerInterceptor())           # Unary拦截器，负责AccessToken校验
grpc.StreamInterceptor(ctx.GrpcStreamServerInterceptor())         # Stream拦截器，负责AccessToken校验
```

并且，进程的Listen地址必须使用Server Context提供的地址:

```
net.Listen("tcp", ctx.GetAddr())
```


### Client Context
对每一个需要访问的Service，都需要创建一个独立的Client Context。跟Server Context类似，每一个Client Context负责维护访问的Service的状态数据以及提供gRPC拦截器等。例如，负责和CA进行AccessToken的申请、缓存。

如果要访问多个Service，应该为每一个访问都创建独立的Client Context。

```go
ctx1 := NewClientContext(env, "xlsoa.example.greeter") 
ctx2 := NewClientContext(env, "xlsoa.example.candy")  
```

配合gRPC使用，需要在`grpc.Dial()`指定以下Options:
```
grpc.WithDialer(ctx.GrpcDialer())                               # sdk需要对Service name进行Dial重定向，需要对gRPC的Dial进行hook
grpc.WithPerRPCCredentials(ctx.GrpcPerRPCCredentials())         # PerRPCCredentials提供AccessToken的承载能力
grpc.WithUnaryInterceptor(ctx.GrpcUnaryClientInterceptor())，   # Unary的拦截器
grpc.WithStreamInterceptor(ctx.GrpcStreamClientInterceptor())   # Stream的拦截器
```

## 3. 配置文件
### creds.json
SOA的身份秘钥文件，可以在SOA Portal下载，可以参考文档[Portal指南](http://xlsoa.gitlab.xunlei.cn/guide/portal-guide/)。

程序的运行目录为`workdir`，Sdk默认从以下路径搜索`creds.json`:
```
<workdir>/creds.json
<workdir>/conf/creds.json
```

***注意:***只有提供了`creds.json`，sdk才提供SOA的访问鉴权保证。如果没有提供，Client端将不会进行AccessToken的申请，Server端将不会对访问进行AccessToken校验。

### soa.yml
通过配置文件可以控制sdk的内部行为，***实际上用户部署的时候并不需要关心这些配置项，只有在调试的时候会用到部分配置项***，在后面的调试章节会详细提到。

一个完整的配置文件格式和说明:
```yaml
modules:
    xlsoa:
        environment:
            transport:
                addr: ""                    # Transport地址，一般为envoy的egress
            prometheus:                  
                listen:
                    addr: ""                # 默认启动一个http的地址提供prometheus数据拉取
                path: ""                    # prometheus拉取metrics的path，默认为'/metrics'
            hosts:                          # hosts可以指定service的直连地址，以及控制是否需要OAuth鉴权。如果在hosts里面匹配到，优先用这个配置，如果没有则用environment.transport.addr
            - service: ""                   # service名称
              addr: ""                      # service的地址，client端用这个地址直连。
              oauth: true|false             # 是否开启Oauth认证，如果true，client会到CA(CA的服务名为xlsoa.core.certificate)进行AccessToken申请。
            ...

        server:                             # 控制Server Context相关逻辑
            context:
                addr: ""                    # Server绑定地址
                oauth:
                    secure:
                        switch: "on|off"    # Server是否启动Oauth的鉴权，如果on，对Client的请求需要进行校验。
                        level: "OauthSecureLevelDegradeWhenException|OauthSecureLevelRigorous"  # Oauth的安全等级，默认为OauthSecureLevelRigorous
```


程序的运行目录为`workdir`，Sdk默认从以下路径搜索配置文件:
```
<workdir>/soa.yml
<workdir>/soa.yaml
<workdir>/xlsoa/conf/internal.yaml
```

在容器环境下，也支持环境变量指定:
```
export XLSOA_KUBESERVICE_CONFIG_FILE=<path-to-soa-config>
```

### 环境变量
可以通过环境变量指定的参数，以下为已经支持的环境变量
```
MODULES_XLSOA_ENVIRONMENT_TRANSPORT_ADDR                # 等同于`soa.yml`的`modules.xlsoa.environment.transport.addr`，下面的都一样，不再一一说明
MODULES_XLSOA_ENVIRONMENT_PROMETHEUS_LISTEN_ADDR
MODULES_XLSOA_ENVIRONMENT_PROMETHEUS_PATH

MODULES_XLSOA_SERVER_CONTEXT_ADDR
MODULES_XLSOA_SERVER_CONTEXT_OAUTH_SECURE_SWITCH
```

***注意:***环境变量的优先级低于配置文件`soa.yml`

## 4. 示例
### Client

以下代码只示例跟Sdk部分相关的初始化部分，完整的例子可以参考`example/greeter/client`的代码。
```go
env := NewEnvironment()                                                 
ctx := NewClientContext(env, "xlsoa.example.greeter")

conn, err := grpc.Dial(
		"xlsoa.example.greeter",
		grpc.WithInsecure(),
		grpc.WithDialer(ctx.GrpcDialer()),
		grpc.WithPerRPCCredentials(ctx.GrpcPerRPCCredentials()),
		grpc.WithUnaryInterceptor(ctx.GrpcUnaryClientInterceptor()),
		grpc.WithStreamInterceptor(ctx.GrpcStreamClientInterceptor()),
	)

```

说明:
- 为每一个需要访问的Service创建一个ClientContext，例如这里为`xlsoa.example.greeter`创建了一个ClientContext
- grpc.Dial()第一个参数为需要访问的SOA的Service name

### Server

以下代码只示例跟Sdk部分相关的初始化部分，完整的例子可以参考`example/greeter/server`的代码。
```go
env := xlsoa.NewEnvironment()
ctx := xlsoa.NewServerContext(env)

svr := grpc.NewServer(
		grpc.UnaryInterceptor(ctx.GrpcUnaryServerInterceptor()),
		grpc.StreamInterceptor(ctx.GrpcStreamServerInterceptor()),
	)

...
net.Listen("tcp", ctx.GetAddr())      # 必须监听Server Context提供的Address
...

```

## 5. How to 调试
我们写好一个程序，需要先调试、自测业务相关的逻辑，这时候一般不希望引入SOA的envoy、服务鉴权、Docker化等。我们支持简单的直连机制，进行方便的业务调试。

调试模式下，Server端和Client端实际上退化为简单的gRPC调用。

下图演示了一个典型的场景：
1. Server是一个调试版本的程序，这个程序不需要接入SOA
2. (A)路线，Client是一个调试客户端，需要访问Server进行调试
3. (B)路线，Client还需要依赖一个外部service，名称为ServiceX，但是ServiceX不提供调试版本，只提供一个部署版本。意味着访问ServiceX必须使用envoy进行路由，并且需要SOA的鉴权机制。

```
    -----------------------------------------------------------------
    | SOA                                                           |
    |                                                               |
    |                                                               |
    |                                                               |
    |      -----------------------                                  |
    |      |    ServiceX         |                                  |
    |      -----------------------                                  |
    |      xluser.core.session                                      |
    |          ^                                                    |
    |          |                     -----------------------        |
    |          |                     |          CA         |        |
    |          |                     -----------------------        |
    |          |                     xlsoa.core.certificate
    |          |                          ^                         |
    |          |                          |                         |
    -----------|--------------------------|--------------------------
               |   |-----------------------
               |   |
               |   |
        ------------------
        |  envoy sidecar |
        ------------------
               ^
               |
               | (B)
               |
        ----------------            (A)              ----------------
        |    Client    |---------------------------->|   Server     |
        ----------------                             ----------------
       xlsoa.example.greeter.client                  xlsoa.example.greeter
```

根据这几个场景分别介绍。

### Server
`xlsoa.example.greeter`是一个调试中的Server，可以选择关闭SOA的鉴权机制。

由于我们不需要SOA相关的鉴权逻辑，也可以不需要`creds.json`。

编辑配置文件`soa.yml`:
```
modules:
    xlsoa:
        server:
            context:
                addr: "<listen-address>"
                oauth:
                    secure:
                        switch: "<secure-switch>"
```

- `listen-address`: Server绑定监听的地址
- `secure-switch`: on|off。打开或关闭Server端SOA的OAuth鉴权

例如:
```
modules:
    xlsoa:
        server:
            context:
                addr: "127.0.0.1:59092"
                oauth:
                    secure:
                        switch: "off"

```

以上配置文件，指定Server绑定在`127.0.0.1:59092`，并且关闭SOA的OAuth校验。

配置完之后，直接运行Server程序即可。


### A路线
A线路是`xlsoa.example.greeter.client`需要访问`xlsoa.example.greeter`。由于`xlsoa.example.greeter`是一个调试版本的Server，没有开启SOA的鉴权。

可以通过配置`hosts`指定Server的直连地址，并且关闭SOA的鉴权机制。

这种模式下，由于我们不需要SOA相关的鉴权逻辑，也可以不需要`creds.json`。

编辑配置文件`soa.yml`:
```
modules:
    xlsoa:
        environment:
            hosts:
            - service: "<service-name>"
              addr: "<service-address>"
              oauth: "<service-oauth>"
```

- `service-name`: 要访问的Service名称，就是我们代码里面grpc.Dial()的目标名
- `service-address`: 该Service的访问地址
- `service-oauth`: true|false。指定该Service是否开启了SOA的Oauth校验。调试的Server端一般都会关闭Oauth校验，这里填false即可

例如:
```
modules:
    xlsoa:
        environment:
            hosts:
            - service: "xlsoa.example.greeter"
              addr: "127.0.0.1:59092"
              oauth: false
```

以上配置文件，指定Client访问`xlsoa.example.greeter`的地址为`127.0.0.1:59092`。并且关闭SOA的OAuth访问鉴权。


### B路线 
B线路是`xlsoa.example.greeter.client`需要访问`xluser.core.session`。`xluser.core.session`是一个正式部署的SOA Service，已经开启了SOA的鉴权，并且可以通过envoy的service mesh进行访问。

有2种方法对这种情况进行调试，并且根据情况这2种方法可以组合进行。

#### 方法1
使用envoy代理，我们已经提供了一个envoy的代理镜像，我们可以运行一个envoy的代理容器进行调试。

使用代理envoy的好处是，我们只需要关心envoy的地址，不需要关心其它Service的地址，envoy会帮我们进行目标路由。

关于如何运行这个代理envoy容器，可以参考[启动代理程序](https://gitlab.xunlei.cn/xlsoa/docs/wikis/how-to-deploy-xlsoa-service-in-docker/#4-%E5%AE%A2%E6%88%B7%E7%AB%AF%E7%A8%8B%E5%BA%8F)

通过配置文件指定`transport`地址，指向envoy的egress address即可。

编辑配置文件`soa.yml`:
```
modules:
    xlsoa:
        environment:
            transport:
                addr: "<envoy-egress-address>"
```

- `envoy-egress-address`: envoy的egress地址，除了`hosts`里面指定的地址，所有对外访问都会通过这个地址出去

假如我们启动的envoy代理程序的egress address为`127.0.0.1:59012`，则以下配置:
```
modules:
    xlsoa:
        environment:
            transport:
                addr: "127.0.0.1:59012"
```

以上配合文件，对外所有的访问，都使用`127.0.0.1:59012`，并且需要进行SOA的OAuth鉴权。

#### 方法2
如果不想单独启动一个envoy的代理，我们也可以通过配置`soa.yml`的`hosts`达到目的。

这种方式没有envoy帮我们进行路由，所以需要显示的配置需要访问的Service的地址。

除了配置`xluser.core.session`的地址，还需要配置SOA的`xlsoa.core.certificate`。

编辑配置文件`soa.yml`:
```
modules:
    xlsoa:
        environment:
            hosts:
            - service: "<ca-service-name>"
              addr: "<ca-address>"
            - service: "<service-name>"
              addr: "<service-address>"
              oauth: "<service-oauth>"

```

- `ca-service-name`: CA的服务名，为`xlsoa.core.certificate`。
- `ca-address`: SOA的CA地址，需要提供这个地址是因为访问`xlsoa.example.greeter`需要进行SOA的OAuth鉴权。
- `service-name`: 要访问的Service名称，就是我们代码里面grpc.Dial()的目标名
- `service-address`: 该Service的访问地址
- `service-oauth`: true|false。指定该Service是否开启了SOA的Oauth校验。调试的Server端一般都会关闭Oauth校验，这里填false即可

例如:
```
modules:
    xlsoa:
        environment:
            hosts:
            - service: "xlsoa.core.certificate"
              addr: "10.101.131.101:4000"
            - service: "xluser.core.session"
              addr: "127.0.0.1:59092"
              oauth: true
```

以上配置文件，指定Client访问`xluser.core.session`的地址为`127.0.0.1:59092`，需要进行OAuth鉴权。并且使用CA的地址为`10.101.131.101:4000`。

## 6. How to部署
程序经过调试之后，我们需要部署到一个测试环境或者是生产环境。部署阶段是一个相对正式的阶段，意味着Servie会部署到SOA的envoy service mesh、服务鉴权、Docker部署等。

详细的部署可以参考文档[SOA Getting Started#部署阶段](http://xlsoa.gitlab.xunlei.cn/guide/getting-started/#3)
