Configure Manager
==========================================
配置管理中心


## 设计文档
[ConfigureManager设计概要](http://gitlab.xunlei.cn/xlsoa/docs/wikis/design-configure-manager)

## Simple usage

### 1. Define your configuration data struct

```
// Names should be exported.
type myConfig struct {
	Name   string `yaml:"name"`
	Server struct {
		Addr string `yaml:"addr"`
		Port int `yaml:"port"`
	} `yaml:"server"`
	Mysql struct {
		Host     string `yaml:"host"`
		Port     int `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
	} `yaml:"mysql"`
	Token struct {
		Ttl int32 `yaml:"token"`
	}
}
```

### 2. Load

```
c = config.New(name, opts...)
if loader, err = c.Load(); err != nil {
    log.Fatal(err)
}

```

### 3. Get & Populate
```
// Get ROOT
if v, err := loader.Get(config.ROOT); err != nil {
    log.Fatal(err)
}
if v == nil {
    log.Fatal("Not exists")
}

var o = &myConfig{}
if err = v.Populate(o); err != nil {
    log.Fatalf("Populate error: %v\n", err)
}
log.Println(o)

```

### 4. Watch update

```
ch, _ := loader.Watch(config.ROOT)
for {
    select {
	case <-ch:
	    v, err = loader.Get(config.ROOT)
		if err != nil || v == nil {
		    continue
	    }

        var o = &myConfig{}
		err = v.Populate(o)
		if err != nil {
		    log.Fatalf("Updated config Populate error: %v\n", err)
		    break
	    }
		log.Println(o)

    }
}

```


## Environment support

```
// Default server address.
XLSOA_CONFIG_ADDR

// Default datacenter name.
XLSOA_DC

// Default node name.
XLSOA_NODE
```


## 过滤器
配置项支持过滤器，下面的例子表示只对机房dc1的节点node1上面的实例instance1有效。

```
config/serviceA/mysql/[dc=dc1,node=node1,instance=instance1]host = "localhost"
```

dc、node、instance可以组合指定，如果有多个可适配配置项，‘最大匹配’的那一项会被生效。
没有任何过滤器的配置项，默认为‘对所有都有效’。

