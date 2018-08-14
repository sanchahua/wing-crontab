Protobuf Define
============================


## 说明
所有服务的RPC协议定义文件，都放在这个仓库，对所有人开放。

## 目录规范
一般分三级，第一级为部门名，第二级为组名，第三级为服务名
```
{DEPARTMENT}/{GROUP}/{SERVICE}.proto
```

例如
```
xlsoa/core/auth.proto
xlsoa/core/key_sync.proto
xluser/core/session.proto
```

## 包名规范
协议的包名规则为

```
{DEPARTMENT}.{GROUP}.{SERVICE}
```


例如

```
package xlsoa.core.auth
package xlsoa.core.key_sync
package xluser.core.session
```
