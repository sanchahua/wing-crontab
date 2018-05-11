package main

import (
	"fmt"
	"os"
)

func main() {
	//[49 50 51 52 53 54]
	var buffer = []byte("123456")

	fmt.Fprintf(os.Stderr, "%+v\n", buffer)

	content := make([]byte, 2)
	copy(content, buffer[:2])
	//content:= buffer[:2]
	buffer = append(buffer[:0], buffer[3:]...)
	fmt.Fprintf(os.Stderr, "%+v\n", content)
	fmt.Fprintf(os.Stderr, "%+v\n", buffer)

}
