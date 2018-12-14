package main

import(
	"bytes"
	"fmt"
"os/exec"
)

func main(){
	cmd := exec.Command("cmd", "/c", "php -v")

	var b bytes.Buffer
	cmd.Stdout = &b
	cmd.Stderr = &b

	cmd.Start()

	cmd.Wait()
	res := b.Bytes()

	fmt.Println(string(res))
}
