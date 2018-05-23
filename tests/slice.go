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
	set(a)
	fmt.Fprintf(os.Stderr, "%+v", a)
	//fmt.Fprintf(os.Stderr, "%+v", a[5:5])

}
