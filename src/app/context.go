package app

import (
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
	"context"
)
// context
type Context struct {
	// canal context
	Ctx context.Context
	// canal context func
	Cancel context.CancelFunc
	// pid file path
	PidFile string
	cancelChan chan struct{}

	Config *Config
}

// new app context
func NewContext() *Context {
	config, _ := getAppConfig()
	ctx := &Context{
		cancelChan:make(chan struct{}),
		Config:config,
	}
	ctx.Ctx, ctx.Cancel = context.WithCancel(context.Background())
	go ctx.signalHandler()
	return ctx
}


func (ctx *Context) Stop() {
	ctx.cancelChan <- struct{}{}
}

func (ctx *Context) Done() <-chan struct{} {
	return ctx.cancelChan
}

func (ctx *Context) Context() context.Context {
	return ctx.Ctx
}

// wait for control + c signal
func (ctx *Context) signalHandler() {
	sc := make(chan os.Signal, 1)
	signal.Notify(sc,
		os.Kill,
		os.Interrupt,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	<-sc
	log.Warnf("get exit signal, service will exit later")
	ctx.cancelChan <- struct{}{}
}
