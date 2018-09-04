package main
import "gitlab.xunlei.cn/xllive/common/consul"
func main() {
	con := consul.NewConsul(consul.DefaultConfig())
	key := "pw/roomuserserver/lock"
	if ok, _ := con.Lock(key, 1000); ok {
		//do something
		con.Unlock(key)
	}
}
