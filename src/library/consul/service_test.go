package consul

import (
	"testing"
	"fmt"
	"time"
)

func TestNewService(t *testing.T) {
	address := "127.0.0.1:8500"
	lock := "test"
	NewService(
		address,
		lock,
		"test1",
		"127.0.0.1",
		7000,
		SetOnLeader(func(isLeader bool) {
			fmt.Println("test1 is leader:", isLeader)
		}))

	NewService(
		address,		lock,

		"test2",
		"127.0.0.1",
		7001,
		SetOnLeader(func(isLeader bool) {
			fmt.Println("test2 is leader:", isLeader)
		}))

	 NewService(
		address,		lock,

		 "test3",
		"127.0.0.1",
		7002,
		SetOnLeader(func(isLeader bool) {
			fmt.Println("test3 is leader:", isLeader)
		}))

	//sev1.Close()
	i := 0
	for {
		time.Sleep(time.Second)
		i++
		if i > 10 {
			break
		}
	}
}
