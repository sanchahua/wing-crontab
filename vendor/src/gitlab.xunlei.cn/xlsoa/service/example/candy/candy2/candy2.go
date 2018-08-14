package main

import (
	"fmt"
	xlsoa "gitlab.xunlei.cn/xlsoa/service"
	pb "gitlab.xunlei.cn/xlsoa/service/proto_gen/xlsoa/example"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
	"net"
)

type Server struct{}

func (s *Server) Get(ctx context.Context, req *pb.CandyGetRequest) (*pb.CandyGetReply, error) {
	log.Printf("Req: %v\n", req)
	return &pb.CandyGetReply{Message: fmt.Sprintf("Got candy from candy2")}, nil
}

func main() {
	env := xlsoa.NewEnvironment()
	ctx := xlsoa.NewServerContext(env)
	svr := grpc.NewServer(
		grpc.UnaryInterceptor(ctx.GrpcUnaryServerInterceptor()),
		grpc.StreamInterceptor(ctx.GrpcStreamServerInterceptor()),
	)

	pb.RegisterCandy2Server(svr, &Server{})
	ls, err := net.Listen("tcp", ctx.GetAddr())
	if err != nil {
		log.Fatalf("Listen error: %v\n", err)
	}
	svr.Serve(ls)
}
