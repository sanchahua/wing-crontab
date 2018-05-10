package main

import (
	"fmt"
	"os"
)

func main() {
	var buffer = []byte("123456")
	buffer = append(buffer[:0], buffer[len(buffer):]...)
	fmt.Fprintf(os.Stderr, "%+v", buffer)

}
