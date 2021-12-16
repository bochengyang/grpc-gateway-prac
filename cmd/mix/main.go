package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/bochengyang/grpc-gateway-prac/pkg/helloworld"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/soheilhy/cmux"
	"github.com/zenazn/goji/graceful"
	"google.golang.org/grpc"
)

// ServiceHandlers is used to implement helloworld.GreeterServer.
type ServiceHandlers struct{}

// SayHello implements helloworld.GreeterServer
func (s *ServiceHandlers) SayHello(ctx context.Context, in *helloworld.HelloRequest) (*helloworld.HelloReply, error) {
	log.Printf("Received: %v", in.Name)
	return &helloworld.HelloReply{Message: "Hello " + in.Name}, nil
}

const (
	port         = ":50001"
	grpcEndpoint = "localhost:50001"
)

// func grpcHandlerFunc(grpcServer *grpc.Server, otherHandler http.Handler) http.Handler {
// 	return h2c.NewHandler(
// 		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 			if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
// 				grpcServer.ServeHTTP(w, r)
// 			} else {
// 				otherHandler.ServeHTTP(w, r)
// 			}
// 		}),
// 		&http2.Server{})
// }

// func main() {
// 	// Another approach from https://github.com/philips/grpc-gateway-example/issues/22#issuecomment-490733965
// 	grpcS := grpc.NewServer()
// 	helloworld.RegisterGreeterServer(grpcS, &ServiceHandlers{})

// 	log.Println("gRPC server is running.")

// 	ctx := context.Background()
// 	ctx, cancel := context.WithCancel(ctx)
// 	defer cancel()

// 	gwS := runtime.NewServeMux()
// 	dialOpts := []grpc.DialOption{grpc.WithInsecure()}

// 	if err := helloworld.RegisterGreeterHandlerFromEndpoint(ctx, gwS, grpcEndpoint, dialOpts); err != nil {
// 		log.Fatalln(err)
// 	}

// 	http.ListenAndServe(port, grpcHandlerFunc(grpcS, gwS))
// }

func main() {

	// Start by setting up a port.
	l, err := net.Listen("tcp", port)
	if err != nil {
		panic(err)
	}

	m := cmux.New(l)
	grpcL := m.MatchWithWriters(cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc"))
	httpL := m.Match(cmux.HTTP1Fast())

	grpcS := grpc.NewServer()
	helloworld.RegisterGreeterServer(grpcS, &ServiceHandlers{})

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	gwS := runtime.NewServeMux()
	dialOpts := []grpc.DialOption{grpc.WithInsecure()}

	if err := helloworld.RegisterGreeterHandlerFromEndpoint(ctx, gwS, grpcEndpoint, dialOpts); err != nil {
		log.Fatalln(err)
	}

	eps := make(chan error, 2)

	// Start the listeners for each protocol
	go func() { eps <- grpcS.Serve(grpcL) }()
	// We use graceful as the server here to serve normal mux
	go func() { eps <- http.Serve(httpL, gwS) }()

	fmt.Printf("listening and serving (multiplexed) on: %s\n", port)
	if err := m.Serve(); err != nil {
		log.Fatalln(err.Error())
	}

	var failed bool
	if err != nil {
	}
	// Handle exiting like they do here: https://github.com/gdm85/grpc-go-multiplex/blob/master/greeter_multiplex_server/greeter_multiplex_server.go
	var i int
	for err := range eps {
		if err != nil {
		}
		i++
		if i == cap(eps) {
			close(eps)
			break
		}
	}
	if failed {
		os.Exit(1)
	}

	graceful.Wait()
}
