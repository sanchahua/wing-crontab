package data

import (
	"testing"
	"fmt"
)

func TestEsQueue_Put(t *testing.T) {
	q := NewQueue(64)

	for i:=0;i<64;i++ {
		fmt.Println(q.Put(i))
	}
	fmt.Println("\n")
	for i:=0;i<64;i++ {
		fmt.Println(q.Get())
		fmt.Println(q.Put(9999))
	}
}
