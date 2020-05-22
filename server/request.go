package server

import (
	"context"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"time"

	app "github.comcast.com/cpp/cpp-update-svc/service"
)

type (
	//CreateFileRequest ...
	CreateFileRequest struct {
		User app.User
		File *multipart.FileHeader
	}

	//CreateFileResponse ...
	CreateFileResponse struct {
		status  int
		message string `json:"message"`
	}
)

//decoder for the incoming create request
func decodeCreateRequest(ctx context.Context, req *http.Request) (interface{}, error) {

	//parsing form data
	email := req.FormValue("email")
	name := req.FormValue("name")
	ntID := req.FormValue("userNtID")

	//parsing multipart file
	if err := req.ParseMultipartForm(20 << 20); err != nil {
		return nil, errors.New("no form data uploaded from frontend")
	}

	//get fileheader for the form file
	_, fileHead, err := req.FormFile("bulk_updates")
	if err != nil {
		return nil, errors.New("no file uploaded")
	}

	//check if uploaded file is CSV, if not then return
	if fileHead.Header.Get("Content-Type") != "text/csv" {
		return nil, errors.New("file format not supported")
	}

	//return the createfileRequest
	return CreateFileRequest{User: app.User{UserNTID: ntID, Name: name, Email: email}, File: fileHead}, nil
}

//encoder for the response
func encodeCreateRequest(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	return json.NewEncoder(w).Encode("File upload completed. We will notify you once processing is done.")
}

func decodeTestRequest(ctx context.Context, req *http.Request) (interface{}, error) {
	return nil, nil
}

func endcodeTestRequest(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	time.Sleep(1 * time.Second)
	return json.NewEncoder(w).Encode(response)
}
