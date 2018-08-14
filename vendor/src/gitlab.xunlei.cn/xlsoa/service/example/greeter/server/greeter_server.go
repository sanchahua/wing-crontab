package main

import (
	xlsoa "gitlab.xunlei.cn/xlsoa/service"
	pb "gitlab.xunlei.cn/xlsoa/service/proto_gen/xlsoa/example"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
	"net"
)

type GreeterServer struct{}

func (s *GreeterServer) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("Req: %v\n", req)
	return &pb.HelloReply{Message: "Hello " + req.Name}, nil
}

func main() {
	env := xlsoa.NewEnvironment()
	ctx := xlsoa.NewServerContext(env)
	svr := grpc.NewServer(
		grpc.UnaryInterceptor(ctx.GrpcUnaryServerInterceptor()),
		grpc.StreamInterceptor(ctx.GrpcStreamServerInterceptor()),
	)
	pb.RegisterGreeterServer(svr, &GreeterServer{})

	ls, err := net.Listen("tcp", ctx.GetAddr())
	if err != nil {
		log.Fatalf("Listen error: %v\n", err)
	}
	svr.Serve(ls)
}
