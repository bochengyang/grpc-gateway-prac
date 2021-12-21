package main

import (
	"context"
	"log"
	"net"

	"github.com/bochengyang/grpc-gateway-prac/pkg/helloworld"
	"github.com/mwitkow/grpc-proxy/proxy"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

var (
	director proxy.StreamDirector
)

// ServiceHandlers is used to implement helloworld.GreeterServer.
type ServiceHandlers struct{}

// SayHello implements helloworld.GreeterServer
func (s *ServiceHandlers) SayHello(ctx context.Context, in *helloworld.HelloRequest) (*helloworld.HelloReply, error) {
	log.Printf("Received: %v", in.Name)
	return &helloworld.HelloReply{Message: "Hello " + in.Name}, nil
}

const (
	port = ":50011"
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

	director := func(ctx context.Context, fullMethodName string) (context.Context, *grpc.ClientConn, error) {
		// // Make sure we never forward internal services.
		// if strings.HasPrefix(fullMethodName, "/com.example.internal.") {
		// 	return nil, nil, grpc.Errorf(codes.Unimplemented, "Unknown method")
		// }
		md, ok := metadata.FromIncomingContext(ctx)
		// Copy the inbound metadata explicitly.
		outCtx, _ := context.WithCancel(ctx)
		outCtx = metadata.NewOutgoingContext(outCtx, md.Copy())
		if ok {
			// // Decide on which backend to dial
			// if val, exists := md[":authority"]; exists && val[0] == "staging.api.example.com" {
			// 	// Make sure we use DialContext so the dialing can be cancelled/time out together with the context.
			// 	conn, err := grpc.DialContext(ctx, "api-service.staging.svc.local", grpc.WithCodec(proxy.Codec()))
			// 	return outCtx, conn, err
			// } else if val, exists := md[":authority"]; exists && val[0] == "api.example.com" {
			// 	conn, err := grpc.DialContext(ctx, "api-service.prod.svc.local", grpc.WithCodec(proxy.Codec()))
			// 	return outCtx, conn, err
			// }
			creds, err := credentials.NewServerTLSFromFile("./ssl/tls.crt", "./ssl/tls.key")
			conn, err := grpc.DialContext(ctx, "localhost:50001", grpc.WithTransportCredentials(creds), grpc.WithCodec(proxy.Codec()))
			return outCtx, conn, err
		}
		return nil, nil, grpc.Errorf(codes.Unimplemented, "Unknown method")
	}

	// // A gRPC server with the proxying codec enabled.
	// s := grpc.NewServer(grpc.Creds(creds), grpc.CustomCodec(proxy.Codec()))
	// // Register a TestService with 4 of its methods explicitly.
	// proxy.RegisterService(s, director,
	// 	"helloworld.Greeter",
	// 	"SayHello")
	s := grpc.NewServer(
		grpc.Creds(creds),
		grpc.CustomCodec(proxy.Codec()),
		grpc.UnknownServiceHandler(proxy.TransparentHandler(director)),
	)
	log.Println("gRPC server is running.")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
