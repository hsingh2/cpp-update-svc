package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.comcast.com/cpp/cpp-update-svc/config"
	"github.comcast.com/cpp/cpp-update-svc/server"
	app "github.comcast.com/cpp/cpp-update-svc/service"

	kitjwt "github.com/go-kit/kit/auth/jwt"

	"github.com/dgrijalva/jwt-go"
)

const (
	defaultPort = "8080"
	configfile  = "config.json"
	authfile    = "publicKey.pem"
)

var (
	logger log.Logger
	srvs   app.FileProcessor
	cfg    config.Config
	errs   chan error
)

func init() {
	//initialize logger
	logger = log.NewLogfmtLogger(os.Stderr)
	logger = log.NewSyncLogger(logger)
	logger = log.With(logger, "servicename", "cpp-update-svc", "time", log.DefaultTimestampUTC, "caller", log.DefaultCaller)

	//initialize the configs
	configs, err := ioutil.ReadFile(configfile)
	if err != nil {
		level.Error(logger).Log("error", "invalid config.json file supplied")
		os.Exit(1)
	}

	if err := json.Unmarshal(configs, &cfg); err != nil {
		level.Error(logger).Log("error", err, "message", "cannot marshal config.json file into config struct")
		os.Exit(1)
	}

	//load authkey
	pubKey, err := ioutil.ReadFile(authfile)
	if err != nil {
		level.Error(logger).Log("error", "no publicKey.pem file supplied")
		os.Exit(1)
	}

	//Parse pubkey
	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(pubKey)
	if err != nil {
		level.Error(logger).Log("error", "no publicKey.pem file supplied")
		os.Exit(1)
	}
	cfg.PubKey = publicKey

	//retrieve the OS environment variables
	//servicepassword
	val, ok := os.LookupEnv("CPP_SERVICE_PASSWORD")
	if !ok {
		level.Error(logger).Log("error", "password an environment variable does not exist")
		os.Exit(1)
	}
	cfg.OSConf.CPPServicePassword = val
	//clientsecret
	val, ok = os.LookupEnv("CLIENT_SECRET")
	if !ok {
		level.Error(logger).Log("error", "secret environment variable does not exist")
		os.Exit(1)
	}
	cfg.OSConf.ClientSecret = val

	val, ok = os.LookupEnv("CPP_ARTIFACTORY_API_KEY")
	if !ok {
		level.Error(logger).Log("error", "artifactory_api_secret environment variable does not exist")
		os.Exit(1)
	}
	cfg.HTTP.APISecret = val
}

func main() {
	level.Info(logger).Log("msg", "service started ....")
	defer level.Info(logger).Log("msg", "service ended.")

	//init context
	ctx := context.Background()

	//spin the server
	go func() {
		//init service
		srvs = app.NewService(cfg, logger)

		//make endpoints
		endpoints := server.MakeEndPoint(srvs)

		//add authentication
		authClient := server.NewAuthClient(cfg, logger)
		authBefore := []kithttp.RequestFunc{kitjwt.HTTPToContext()}

		//create server
		handler := server.NewHTTPServer(ctx, endpoints, authClient.ValidateAuth(), authBefore...)

		errs <- http.ListenAndServe(":8080", handler)
	}()

	level.Error(logger).Log("exit error : ", <-errs)
}
