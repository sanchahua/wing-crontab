package main

import (
	"fmt"
	"os"
)

func set(a []int) {
	a = make([]int, 3)
}
func main() {
	a := make([]int, 1)
	a[0] = 1

	fmt.Fprintf(os.Stderr, "%+v", a[1:])
	//fmt.Fprintf(os.Stderr, "%+v", a[5:5])

}
