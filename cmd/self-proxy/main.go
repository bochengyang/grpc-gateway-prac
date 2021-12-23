package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/mwitkow/grpc-proxy/proxy"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

const (
	port = ":50002"
)

func grpcHandlerFunc() http.Handler {
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
	return http.HandlerFunc(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			s.ServeHTTP(w, r)
		}),
	)
}

func proxyHandler(w http.ResponseWriter, r *http.Request) {
	// We can't have this set. And it only contains "/pkg/net/http/" anyway
	r.RequestURI = ""

	// Since the req.URL will not have all the information set,
	// such as protocol scheme and host, we create a new URL
	u, err := url.Parse(fmt.Sprintf("https://localhost:50001%s", r.URL))
	if err != nil {
		panic(err)
	}
	r.URL = u

	log.Println(fmt.Sprintf("%+v\n", r))
	requestDump, err := httputil.DumpRequest(r, true)
	if err != nil {
		log.Println(err.Error())
	}
	log.Println(string(requestDump))

	client := &http.Client{Timeout: time.Second * 10}

	res, err := client.Do(r)
	r.Body.Close()
	if err != nil {
		log.Fatalf("Failed: %s", err)
	}
	defer res.Body.Close()

	// Copy the response header to the header of ResponseWriter
	for k, vv := range res.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}

	w.WriteHeader(res.StatusCode)

	io.Copy(w, res.Body)
}

func main() {
	// mux := http.NewServeMux()
	// mux.HandleFunc("/", proxyHandler)
	// srv := &http.Server{
	// 	Addr:    port,
	// 	Handler: mux,
	// }
	// srv.ListenAndServeTLS("./ssl/tls.crt", "./ssl/tls.key")

	// Create gRPC Server
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	http.ServeTLS(lis, grpcHandlerFunc(), "./ssl/tls.crt", "./ssl/tls.key")
}
