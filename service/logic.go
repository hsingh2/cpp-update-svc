package app

import (
	"bufio"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"os"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.comcast.com/cpp/cpp-update-svc/config"

	"github.com/gocarina/gocsv"
)

//should come from the config
const (
	numJobs      = 100
	sourceSystem = "bulkupload"
	tempDirPath  = "temporary_files"
)

//each time you open up a file and you start a worker for the Notifier
//read jobs
var (
	jobs chan FileRequest
)

//bulkFile is the struct to load the file attributes
type bulkFile struct {
	ITRCDBID    string `csv:"itrc_db_id,omitempty"`
	LocationMD5 string `csv:"location_md5,omitempty"`
	UpdatedBy   string `csv:"UpdatedBy,omitempty"`
	PITag       string `csv:"PiTag,omitempty"`
	NewPITag    string `csv:"NewPiTag,omitempty"`
	NewComment  bool   `csv:"NewComment,omitempty"`
}

type service struct {
	cfg       config.Config
	logger    log.Logger
	outClient HTTPClient
	mailer    Mailer
}

//NewService ...
func NewService(config config.Config, logger log.Logger) FileProcessor {
	//init the file reading jobs
	jobs = make(chan FileRequest, numJobs)

	//return the service
	srvs := &service{
		cfg:       config,
		logger:    logger,
		outClient: NewOutboundClient(config, logger),
		mailer:    NewMailer(config, logger),
	}

	//start file reader
	go ReadFile(srvs, jobs)

	return srvs
}

//CreateFile ...
func (s service) CreateFile(ctx context.Context, user User, reqFile *multipart.FileHeader) error {
	logger := log.With(s.logger, "method", "CreateFile")

	source, err := reqFile.Open()
	if err != nil {
		level.Error(logger).Log("failed to open request multipart file: ", err.Error())
		return err
	}
	defer source.Close()

	//read header of the file and validate it before creating a temp file and writing data to it
	r := csv.NewReader(bufio.NewReader(source))
	header, err := r.Read()
	if err != nil {
		level.Error(logger).Log("failed to read the multipart file, currupt data: ", err.Error())
		return err
	}

	//validate the header if supported or not - length and the column names
	if err := isValidFile(header); err != nil {
		level.Info(logger).Log("file header is not correct format")
		return err
	}

	//rewind the file
	source.Seek(0, 0)

	//create a directory for temp files
	if _, err := os.Stat(tempDirPath); os.IsNotExist(err) {
		os.Mkdir(tempDirPath, os.ModePerm)
	}

	//create a temporary file and copy the multipart form file to it
	temporary, err := ioutil.TempFile(tempDirPath, fmt.Sprintf("bulk_%s_%s_*.csv", user.UserNTID, time.Now().Format("200601021504")))
	if err != nil {
		level.Error(logger).Log("failed to create temporary file: ", err.Error())
		return err
	}
	defer temporary.Close()

	if _, err := io.Copy(temporary, source); err != nil {
		level.Error(logger).Log("failed to copy sourcefile: ", err.Error())
		return err
	}

	//add job to the to the file reading queue
	jobs <- FileRequest{temporary.Name(), user}

	return nil
}

//ReadFile ...
func ReadFile(s *service, jobs chan FileRequest) error {
	logger := log.With(s.logger, "method", "ReadFile")

	//job here is the read request for the file. this is a worker which keeps on reading the file
	for job := range jobs {

		file, err := os.Open(job.FileName)
		if err != nil {
			level.Error(logger).Log("failed to read file to process: ", err.Error())
			return err
		}
		defer file.Close()

		//each time you start reading a file, open a channel which listen to errors coming from the
		//and then close it once the file reading is done to notify and signal the event
		resultSet := make([]Error, 0)

		//create an array of the type records to read the fle
		records := []*bulkFile{}

		if err := gocsv.UnmarshalFile(file, &records); err != nil {
			level.Error(logger).Log(fmt.Sprintf("failed to unmarshall file uploaded by: %s, filename: %s", job.Email, job.FileName), err.Error())
			//upload bad file
			s.outClient.UploadFile(job.FileName)
			//send email about bad file
			s.mailer.SendMail(Message{EmailTo: job.Email, UserName: job.User.Name, Type: "error"})
			return err
		}

		//rewind
		if _, err := file.Seek(0, 0); err != nil {
			level.Error(logger).Log(fmt.Sprintf("failed to rewind csv file %s:", job.FileName), err.Error())
			return err
		}

		level.Info(logger).Log("started reading file :", job.FileName)
		for _, record := range records {
			if record == nil {
				continue
			}

			//read each record, create a dataservice request and make a outbound call
			request, err := createOutBoundRequest(job.User, *record)
			if err != nil {
				resultSet = append(resultSet, Error{request.Payload, err})
			} else {
				if err := s.outClient.HTTPDo(request); err != nil {
					resultSet = append(resultSet, *err)
				}
			}
		}

		//Results can then be mailed to the customer once the process is completed.
		//Also we can push the errors to a file and write it to the cloud foundary
		//drop an email to the customer
		go func(fileName string, user User, result []Error) {
			level.Info(s.logger).Log("finished processing file: ", fileName)
			//send mail
			go s.mailer.SendMail(Message{EmailTo: user.Email, UserName: user.Name, Type: "success", Body: result})

			//trigger file upload
			go s.outClient.UploadFile(fileName)

		}(job.FileName, job.User, resultSet)
	}

	return nil
}

// createOutBoundRequest wraps the payload and request type
func createOutBoundRequest(user User, record bulkFile) (DataServiceRequest, error) {
	dataRequest := DataServiceRequest{
		Payload: Payload{
			LocationAttributeMD5: record.LocationMD5,
			ItrcDBID:             record.ITRCDBID,
			PIID:                 record.PITag,
			SourceSystem:         sourceSystem,
			UpdatedBy:            user.UserNTID,
		},
	}
	//skip records with
	if record.ITRCDBID == "" || record.LocationMD5 == "" || record.UpdatedBy == "" {
		return dataRequest, errors.New("missing one or more maindatory field itrcdbid, md5, updateby")
	}

	//bad request if both NewPIID and NewComment passed
	if len(record.NewPITag) > 0 && record.NewComment {
		return dataRequest, errors.New("both newpiid and newcomment are passed")
	}

	//this is the updateComment request
	if len(record.NewPITag) > 0 {
		dataRequest.NewPIID = record.NewPITag
		dataRequest.RequestType = UpdateCommentRequest
		return dataRequest, nil
	}

	//check NewComment flag for bool value which triggers addCommentRequest
	if record.NewComment {
		dataRequest.RequestType = AddCommentRequest
		return dataRequest, nil

	}

	//if not update or create request, then its verify request
	dataRequest.IsVerify = true
	dataRequest.RequestType = VerifyCommentRequest

	return dataRequest, nil
}
