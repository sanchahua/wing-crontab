package main

import (
	"fmt"
	"os"
)

func set(a []int) {
	a = make([]int, 3)
}
func main() {
	a := make([]int, 3)
	a[0] = 1
	a[1] = 2
	a[2] = 3

	fmt.Fprintf(os.Stderr, "%+v", a[1:])
	a = append(a[:2], a[3:]...)
	fmt.Println(a)
	//fmt.Fprintf(os.Stderr, "%+v", a[5:5])

}
