package main

import (
"context"
"os/exec"
	"os"
	"bytes"
	"time"
	"fmt"
)

func main() {
	ctx, _ := context.WithCancel(context.Background())

	cmd := exec.CommandContext(ctx, "php", "/Users/yuyi/Code/go/wing-crontab/tests/runner.php")
	//cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	var b bytes.Buffer
	cmd.Stdout = &b
	cmd.Stderr = &b

	cmd.Start()
	//bufio.NewReader()
	//res, err := cmd.CombinedOutput()
	//fmt.Println(string(b.Bytes()))


	//time.Sleep(1 * time.Second)
	//fmt.Println("退出程序中...", cmd.Process.Pid)
	//cancel()

	c := make(chan struct{})
	go func() {
		cmd.Wait()
		fmt.Println(string(b.Bytes()))
		c <- struct{}{}
	}()

	select {
	case <- c :
		fmt.Println("run complete")
	case <- time.After(time.Second * 10):
		cmd.Process.Kill()
		fmt.Println("run timeout")
	}

}
