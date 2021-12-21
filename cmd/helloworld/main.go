package main

import (
	"context"
	"log"
	"net"

	"github.com/bochengyang/grpc-gateway-prac/pkg/helloworld"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// ServiceHandlers is used to implement helloworld.GreeterServer.
type ServiceHandlers struct{}

// SayHello implements helloworld.GreeterServer
func (s *ServiceHandlers) SayHello(ctx context.Context, in *helloworld.HelloRequest) (*helloworld.HelloReply, error) {
	log.Printf("Received: %v", in.Name)
	return &helloworld.HelloReply{Message: "Hello " + in.Name}, nil
}

const (
	port = ":50001"
)

func main() {
	// Create gRPC Server
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	// Create tls based credential.
	creds, err := credentials.NewServerTLSFromFile("./ssl/tls.crt", "./ssl/tls.key")
	if err != nil {
		log.Fatalf("failed to create credentials: %v", err)
	}
	s := grpc.NewServer(grpc.Creds(creds))
	log.Println("gRPC server is running.")
	helloworld.RegisterGreeterServer(s, &ServiceHandlers{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
