package server

import (
	"context"
	"net/http"

	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
	mux "github.com/gorilla/mux"
)

//NewHTTPServer ...
func NewHTTPServer(ctx context.Context, endpoints EndPoint, middlerware endpoint.Middleware, auth ...httptransport.RequestFunc) http.Handler {
	r := mux.NewRouter()

	r.Use(commonMiddleware)

	options := []httptransport.ServerOption{
		httptransport.ServerBefore(auth...),
	}

	r.Methods(http.MethodPost).Path("/uploadFile").Handler(httptransport.NewServer(
		middlerware(endpoints.CreateFile),
		decodeCreateRequest,
		encodeCreateRequest,
		options...,
	))

	r.Methods(http.MethodGet).Path("/test").Handler(httptransport.NewServer(
		endpoints.TestEndpoint,
		decodeTestRequest,
		endcodeTestRequest,
		options...,
	))

	return r
}

func commonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		resp.Header().Add("Content-Type", "application/json")
		next.ServeHTTP(resp, req)
	})
}
