package main

import "fmt"

func main() {
	m := make(map[int] int)
	m[99] = 99
	m[1] = 1
	m[0] = 0

	for key, value := range m  {
		fmt.Println(key, "=>", value)
	}
}
