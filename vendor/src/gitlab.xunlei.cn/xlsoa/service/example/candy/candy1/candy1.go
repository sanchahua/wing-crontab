package main

import (
	xlsoa "gitlab.xunlei.cn/xlsoa/service"
	pb "gitlab.xunlei.cn/xlsoa/service/proto_gen/xlsoa/example"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
	"net"
)

type Server struct {
	c pb.Candy2Client
}

func (s *Server) Get(ctx context.Context, req *pb.CandyGetRequest) (*pb.CandyGetReply, error) {
	log.Printf("Req: %v\n", req)

	// Notice: Use `ctx` from the Server.Get(), in order to make tracing correct.
	r, err := s.c.Get(ctx, &pb.CandyGetRequest{Name: "Request from candy1"})
	if err != nil {
		return &pb.CandyGetReply{Message: "Get fail from candy2 "}, nil
	}
	return &pb.CandyGetReply{Message: r.Message}, nil
}

func main() {
	// Create environment
	env := xlsoa.NewEnvironment()

	// Create candy2Client
	ctxC := xlsoa.NewClientContext(env, "xlsoa.example.candy2")
	conn, err := grpc.Dial(
		"xlsoa.example.candy2",
		grpc.WithInsecure(),
		grpc.WithDialer(ctxC.GrpcDialer()),
		grpc.WithPerRPCCredentials(ctxC.GrpcPerRPCCredentials()),
		grpc.WithUnaryInterceptor(ctxC.GrpcUnaryClientInterceptor()),
		grpc.WithStreamInterceptor(ctxC.GrpcStreamClientInterceptor()),
	)
	if err != nil {
		log.Fatal(err)
	}
	c := pb.NewCandy2Client(conn)

	// Create server
	ctxS := xlsoa.NewServerContext(env)
	svr := grpc.NewServer(
		grpc.UnaryInterceptor(ctxS.GrpcUnaryServerInterceptor()),
		grpc.StreamInterceptor(ctxS.GrpcStreamServerInterceptor()),
	)

	pb.RegisterCandy1Server(svr, &Server{c: c})
	ls, err := net.Listen("tcp", ctxS.GetAddr())
	if err != nil {
		log.Fatalf("Listen error: %v\n", err)
	}
	svr.Serve(ls)
}
