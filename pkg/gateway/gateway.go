package gateway

import (
	"context"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"

	"github.com/bochengyang/grpc-gateway-prac/pkg/helloworld"
)

func New(ctx context.Context, endpoint string) (http.Handler, error) {
	gw := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}

	if err := helloworld.RegisterGreeterHandlerFromEndpoint(ctx, gw, endpoint, opts); err != nil {
		return nil, err
	}

	mux := http.NewServeMux()
	mux.Handle("/", gw)

	return mux, nil
}
