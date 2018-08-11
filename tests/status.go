package main

import "fmt"

func main() {
	status := 0
	status |= 1
	fmt.Println(status, status&1)
	status |= 1
	fmt.Println(status, status&1)
}
