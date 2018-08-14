package main

import (
	xlsoa "gitlab.xunlei.cn/xlsoa/service"
	pb "gitlab.xunlei.cn/xlsoa/service/proto_gen/xlsoa/example"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
	"time"
)

func main() {
	env := xlsoa.NewEnvironment()
	ctx := xlsoa.NewClientContext(env, "xlsoa.example.candy1")

	conn, err := grpc.Dial(
		"xlsoa.example.candy1",
		grpc.WithInsecure(),
		grpc.WithDialer(ctx.GrpcDialer()),
		grpc.WithPerRPCCredentials(ctx.GrpcPerRPCCredentials()),
		grpc.WithUnaryInterceptor(ctx.GrpcUnaryClientInterceptor()),
		grpc.WithStreamInterceptor(ctx.GrpcStreamClientInterceptor()),
	)
	if err != nil {
		log.Fatal(err)
	}

	c := pb.NewCandy1Client(conn)
	for {
		r, err := c.Get(context.Background(), &pb.CandyGetRequest{Name: "Request from candy client"})
		if err != nil {
			log.Printf("CandyGet error: %v\n", err)
		} else {
			log.Printf("CandyGet: %v\n", r.Message)
		}
		time.Sleep(1 * time.Second)
	}
}
