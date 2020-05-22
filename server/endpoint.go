package server

import (
	"context"

	"github.com/go-kit/kit/endpoint"

	app "github.comcast.com/cpp/cpp-update-svc/service"
)

//EndPoint ...
type EndPoint struct {
	CreateFile   endpoint.Endpoint
	TestEndpoint endpoint.Endpoint
}

//MakeEndPoint ...
func MakeEndPoint(s app.FileProcessor) EndPoint {
	return EndPoint{
		CreateFile:   makeCreateFileEndpoint(s),
		TestEndpoint: makeTestEndpoint(s),
	}
}

func makeCreateFileEndpoint(s app.FileProcessor) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(CreateFileRequest)
		err := s.CreateFile(ctx, req.User, req.File)
		return CreateFileResponse{status: 200, message: "File successfully uploaded, we will notify you once it is processed."}, err
	}
}

func makeTestEndpoint(s app.FileProcessor) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		return CreateFileResponse{status: 200}, nil
	}
}
