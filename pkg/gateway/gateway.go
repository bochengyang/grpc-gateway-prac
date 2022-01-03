package gateway

import (
	"context"
	"crypto/x509"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/bochengyang/grpc-gateway-prac/pkg/helloworld"
)

func New(ctx context.Context, endpoint string) (http.Handler, error) {
	gw := runtime.NewServeMux()
	pool, _ := x509.SystemCertPool()
	// error handling omitted
	creds := credentials.NewClientTLSFromCert(pool, "")
	opts := []grpc.DialOption{grpc.WithTransportCredentials(creds)}

	if err := helloworld.RegisterGreeterHandlerFromEndpoint(ctx, gw, endpoint, opts); err != nil {
		return nil, err
	}

	mux := http.NewServeMux()
	mux.Handle("/", gw)

	return mux, nil
}
