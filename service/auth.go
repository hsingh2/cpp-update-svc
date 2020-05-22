package app

import (
	"context"
	"os"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.comcast.com/cpp/cpp-update-svc/config"
	"golang.org/x/oauth2"
)

var (
	authToken *oauth2.Token
)

const (
	renewTokenDuration = 6
)

//OAuthClient ...
type OAuthClient struct {
	auth          *oauth2.Config
	Username      string
	Password      string
	RenewInterval int
	logger        log.Logger
}

//NewOAuthClient wrapper around oauth2 client package...
func NewOAuthClient(conf config.Config, logger log.Logger) OAuth {
	clt := &OAuthClient{
		&oauth2.Config{
			ClientID:     conf.Auth.ClientID,
			ClientSecret: conf.OSConf.ClientSecret,
			Endpoint: oauth2.Endpoint{
				AuthURL:  conf.Auth.AuthorizationURL,
				TokenURL: conf.Auth.TokenURL},
			Scopes: conf.Auth.Scopes,
		},
		conf.Auth.Username,
		conf.OSConf.CPPServicePassword,
		conf.Auth.RenewInterval,
		log.With(logger, "auth", "oauth"),
	}
	//call init token for the first time
	clt.initToken()

	return clt
}

//OAuth ...
type OAuth interface {
	GetNewAuthToken() string
}

//instantiate the token
func (clt *OAuthClient) initToken() {
	token, err := clt.auth.PasswordCredentialsToken(context.Background(), clt.Username, clt.Password)
	if err != nil {
		level.Error(clt.logger).Log("service failed", "error caused initializing oauth token client", "error", err)
		os.Exit(1)
	}

	//initialize token
	authToken = token

	//run the timer to update token every renewInterval
	go func(clt *OAuthClient) {
		ticker := time.NewTicker(time.Duration(clt.RenewInterval) * time.Hour)
		for range ticker.C {
			token, err := clt.auth.PasswordCredentialsToken(context.Background(), clt.Username, clt.Password)
			if err != nil {
				level.Error(clt.logger).Log("error caused renewing oauth token", err)
			}
			authToken = token
		}
	}(clt)
}

//GetNewAuthToken returns a new auth token from auth client ...
func (clt *OAuthClient) GetNewAuthToken() string {
	if authToken != nil {
		return authToken.AccessToken
	}
	level.Error(clt.logger).Log("authclient is nil")
	return ""
}
