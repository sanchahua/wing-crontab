package service

import (
	//grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	//"github.com/pkg/errors"
	"gitlab.xunlei.cn/xlsoa/service/log"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"net"
	"time"
)

type ClientContext struct {
	env         *Environment
	serviceName string
	callCreds   *oAuthCallCredentials
	tracer      *zipkinTracer
}

func NewClientContext(env *Environment, serviceName string) *ClientContext {
	c := &ClientContext{
		env:         env,
		serviceName: serviceName,
	}

	// Log debug
	log.Printf("[xlsoa] [client context] ServiceName '%v'\n", c.serviceName)
	log.Printf("[xlsoa] [client context] Addresss '%v'\n", c.env.GetServiceAddr(c.serviceName))
	log.Printf("[xlsoa] [client context] OauthSecure '%v'\n", c.env.CheckServiceOauthSecure(c.serviceName))

	c.initOauthSecure()
	c.tracer = NewZipkinTracer()

	return c
}

func (c *ClientContext) Close() {
	if c.callCreds != nil {
		c.callCreds.Close()
		c.callCreds = nil
	}
}

func (c *ClientContext) initOauthSecure() {
	var err error

	if !c.env.CheckServiceOauthSecure(c.serviceName) {
		return
	}

	if c.callCreds, err = NewOauthCallCredentials(c); err != nil {
		log.Printf("[xlsoa] [client context] [Warn] NewOauthCallCredentials error: %v. No soa access token will be acquired.\n", err)
		return
	}
}

func (c *ClientContext) GetEnv() *Environment {
	return c.env
}

func (c *ClientContext) GetServiceName() string {
	return c.serviceName
}

func (c *ClientContext) GrpcDialer() func(string, time.Duration) (net.Conn, error) {
	return c.env.GrpcDialer()
}

func (c *ClientContext) GrpcUnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return nil
}

func (c *ClientContext) GrpcStreamClientInterceptor() grpc.StreamClientInterceptor {
	return nil
}

func (c *ClientContext) GrpcPerRPCCredentials() credentials.PerRPCCredentials {
	return c
}

// interface credentials.PerRPCCredentials
func (c *ClientContext) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {

	var err error
	var mdX map[string]string
	md := make(map[string]string)

	// Oauth
	if c.callCreds != nil {
		if mdX, err = c.callCreds.GetRequestMetadata(ctx, uri...); err != nil {
		} else {
			for k, v := range mdX {
				md[k] = v
			}
		}
	}

	// Tracer
	if mdX, err = c.tracer.GetRequestMetadata(ctx, uri...); err != nil {
	} else {
		for k, v := range mdX {
			md[k] = v
		}
	}

	return md, nil
}

// interface credentials.PerRPCCredentials
func (c *ClientContext) RequireTransportSecurity() bool {
	return false
}
