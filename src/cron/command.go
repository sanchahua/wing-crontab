package cron
//
//import (
//	//log "github.com/cihub/seelog"
//	"gitlab.xunlei.cn/xllive/common/log"
//	"os/exec"
//	"time"
//	"sync/atomic"
//	time2 "library/time"
//	"sync"
//	"bytes"
//	"context"
//	"errors"
//	"os"
//	"fmt"
//)
//
//type Command struct {
//	lock *sync.RWMutex
//	ProcessNum int64
//	command string
//	Process map[int]*os.Process
//	onStart OnCommandRunStart
//	onEnd onCommandRunEnd
//}
//
//type OnCommandRunStart func(processId int, state, output, command, startTime string)
//type onCommandRunEnd func(processId int, state, output string, useTime int64, command, startTime string)
//
//func NewCommand(command string, onStart OnCommandRunStart, onEnd onCommandRunEnd) *Command {
//	c := &Command{
//		lock: new(sync.RWMutex),
//		command: command,
//		Process: make(map[int]*os.Process),
//		onStart: onStart,
//		onEnd: onEnd,
//	}
//	return c
//}
//
//func (row *Command) getProcessNum() int64 {
//	return atomic.LoadInt64(&row.ProcessNum)
//}
//
//func (row *Command) run() {
//	row.lock.Lock()
//	processNum := atomic.AddInt64(&row.ProcessNum, 1)
//	row.lock.Unlock()
//	var cmd *exec.Cmd
//	var err error
//	startTime := time2.GetDayTime()
//	log.Tracef( "##start run: %v=>[%v]", processNum, row.command)
//	start := time.Now().UnixNano()/1000000
//	cmd = exec.Command("bash", "-c", row.command)
//	var b bytes.Buffer
//	cmd.Stdout = &b
//	cmd.Stderr = &b
//
//	err = cmd.Start()
//	processId := 0
//	if cmd != nil && cmd.Process != nil {
//		processId = cmd.Process.Pid
//		row.lock.Lock()
//		row.Process[cmd.Process.Pid] = cmd.Process
//		row.lock.Unlock()
//		defer func() {
//			row.lock.Lock()
//			delete(row.Process, processId)
//			row.lock.Unlock()
//		}()
//	}
//
//	state := StateStart
//	output := ""
//	if err != nil {
//		state = StateStart+"-"+StateFail
//		output = err.Error()
//	}
//
//	row.onStart(processId, state, output, row.command, startTime)
//	err = cmd.Wait()
//
//	res := b.Bytes()
//
//	//res, err := cmd.CombinedOutput()
//	state = StateSuccess
//	if err != nil {
//		state = StateFail
//		res = append(res, []byte("  error: " + err.Error())...)
//		log.Errorf("runCommand fail, command=[%v], error=[%+v]", row.command, err)
//	}
//	log.Tracef( "##%v run end: %v, %v=>[%v]", processNum, processId, row.command)
//	atomic.AddInt64(&row.ProcessNum, -1)
//
//	useTime := int64(time.Now().UnixNano()/1000000 - start)
//	// todo 程序退出时，定时任务的日志可能会失败，因为这个时候数据库已经关闭，这个问题需要处理一下
//	// 即安全退出问题，kill -9没办法了
//	row.onEnd(processId, state, string(res), useTime, row.command, startTime)
//}
//
//
//// 支持超时设定的运行命令
//// 仅仅是运行
//// 不支持onStart、onEnd回调
//func (row *Command) runWithTimeout(duration time.Duration) ([]byte, int, error) {
//	ctx, _ := context.WithCancel(context.Background())
//	cmd := exec.CommandContext(ctx, "bash", "-c", row.command)
//	var b bytes.Buffer
//	cmd.Stdout = &b
//	cmd.Stderr = &b
//	err := cmd.Start()
//	if err != nil {
//		return nil, 0, err
//	}
//
//	c  := make(chan []byte)
//	er := make(chan error)
//
//	defer func() {
//		close(c)
//		close(er)
//	}()
//
//	processId := 0
//	if cmd != nil && cmd.Process != nil {
//		processId = cmd.Process.Pid
//	}
//
//	go func() {
//		defer func() {
//			if err := recover(); err != nil {
//				fmt.Println("err:", err)
//			}
//		}()
//		err := cmd.Wait()
//		if err != nil {
//			er <- err
//		} else {
//			c <- b.Bytes()
//		}
//	}()
//
//	select {
//	case r := <- c :
//		return r, processId, nil
//	case err := <- er:
//		return nil, processId, err
//	case <- time.After(duration):
//		err := cmd.Process.Kill()
//		if err != nil {
//			return nil, processId, errors.New("timeout with error: "+ err.Error())
//		}
//		return nil, processId, ErrTimeout
//	}
//	return nil, processId, ErrUnknown
//}
//
