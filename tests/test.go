package main

import (
	"fmt"
	"time"
)

type A struct {
	a string
}

func main() {
	a := &A{}
	a.a = "hello"
	geta := func() string {
		return a.a
	}
	for i := 0; i < 5000; i++ {
		go func() {
			start := time.Now()
			av := geta()
			fmt.Println("use time ", time.Since(start))
			fmt.Println(av)
		}()
	}
}
