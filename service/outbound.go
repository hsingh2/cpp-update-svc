package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.comcast.com/cpp/cpp-update-svc/config"
)

type outClient struct {
	hClt   http.Client
	cfg    config.Config
	logger log.Logger
	auth   OAuth
}

//Payload ...
type Payload struct {
	LocationAttributeMD5 string `json:"locationAttributeMD5"`
	ItrcDBID             string `json:"itrcDbId"`
	PIID                 string `json:"piId"`
	NewPIID              string `json:"newPiId,omitempty"`
	//for the new record call
	IsVerify     bool   `json:"isVerified,omitempty"`
	SourceSystem string `json:"sourceSystem"`
	UpdatedBy    string `json:"updatedBy"`
}

//DataServiceRequest ....
type DataServiceRequest struct {
	Payload
	RequestType string
}

//Error is outbound error struct
type Error struct {
	Payload `json:"Payload"`
	Err     error `json:"Error"`
}

//NewOutboundClient ...
func NewOutboundClient(conf config.Config, logger log.Logger) HTTPClient {
	return &outClient{
		hClt:   http.Client{Timeout: time.Duration(conf.HTTP.TimeOut) * time.Second},
		cfg:    conf,
		logger: log.With(logger, "outbound", "http"),
		auth:   NewOAuthClient(conf, logger),
	}
}

//HTTPClient ...
type HTTPClient interface {
	HTTPDo(request DataServiceRequest) *Error
	UploadFile(string) *Error
}

func (outClt *outClient) UploadFile(fname string) *Error {
	r, w := io.Pipe()
	m := multipart.NewWriter(w)

	go func() {
		defer w.Close()
		defer m.Close()

		part, err := m.CreateFormFile("upload-data", fname)
		if err != nil {
			return
		}
		file, err := os.Open(fname)
		if err != nil {
			return
		}
		defer file.Close()
		if _, err = io.Copy(part, file); err != nil {
			return
		}
	}()

	//get filename by splitting fullname
	dir := strings.Split(fname, "/")

	request, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s%s", outClt.cfg.HTTP.FileUploadURL, dir[1]), r)
	if err != nil {
		return &Error{Err: err}
	}

	//add headers
	request.Header.Add("x-api-key", outClt.cfg.HTTP.APISecret)
	//retry file upload for 5 times
	for i := 0; i < 5; i++ {
		//make http call
		response, err := outClt.hClt.Do(request)
		if err != nil {
			level.Error(outClt.logger).Log("could not put file in artifactory :", fname)
		}

		if response.StatusCode == http.StatusCreated {
			if err := os.Remove(fname); err != nil {
				level.Info(outClt.logger).Log("could not delete file from container:", fname)
			}
			break
		}
		time.Sleep(time.Duration(10 * time.Second))
	}

	level.Info(outClt.logger).Log("successfully uploaded file to artifactory")

	return nil
}

func (outClt *outClient) HTTPDo(request DataServiceRequest) *Error {
	body, err := json.Marshal(request.Payload)
	if err != nil {
		return &Error{request.Payload, err}
	}

	hRequest, err := http.NewRequest(http.MethodPost, outClt.cfg.HTTP.DataServiceURL[request.RequestType], bytes.NewBuffer(body))
	if err != nil {
		return &Error{request.Payload, err}
	}

	//add headers, authorization
	hRequest.Header.Add("Authorization", fmt.Sprintf("Bearer %s", outClt.auth.GetNewAuthToken()))

	//do http call
	response, err := outClt.hClt.Do(hRequest)
	if err != nil {
		level.Error(outClt.logger).Log("failed to call data service :", err, "request :", request.Payload)
		return &Error{request.Payload, fmt.Errorf("failed to do http call to data service")}
	}

	//check status code, if not 200 or 201 return error
	if !(response.StatusCode == http.StatusOK || response.StatusCode == http.StatusCreated) {
		return &Error{request.Payload, fmt.Errorf("data service returned bad httpcode :%d", response.StatusCode)}
	}

	return nil
}
