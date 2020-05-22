package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	stdjwt "github.com/dgrijalva/jwt-go"
	"github.com/go-kit/kit/auth/jwt"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.comcast.com/cpp/cpp-update-svc/config"
)

const (
	AuthUserID contextKey = "UserId"
)

var (
	HttpNotAllowed  = errors.New("http not allowed")
	XFFProtoMissing = errors.New("X-Forwarded-Proto header missing")
	ErrForbidden    = errors.New("User not authorized for operation")
	ErrInvalidToken = errors.New("token contains an invalid number of segments")
)

// authorizor - struct for authorizor values
type authorizor struct {
	config     config.Config
	logger     log.Logger
	httpClient http.Client
}

// Profile - struct for profile values
type Profile struct {
	Name     string   `json:"name"`
	MemberOf []string `json:"comcast_groups"`
}

// Claims - struct for jwt claims
type Claims struct {
	LastName   string `json:"COMCAST_LNAME"`
	UserName   string `json:"COMCAST_USERNAME"`
	Email      string `json:"COMCAST_EMAIL"`
	ObjGUID    string `json:"COMCAST_OBJGUID"`
	ObjGUIDb64 string `json:"COMCAST_OBJGUID_BASE64"`
	GUID       string `json:"COMCAST_GUID"`
	stdjwt.StandardClaims
}

//NewAuthClient ...
func NewAuthClient(cfg config.Config, logger log.Logger) Auth {
	return &authorizor{cfg, log.With(logger, "server-auth", "oauth"), http.Client{}}
}

//Auth ...
type Auth interface {
	ValidateAuth() endpoint.Middleware
}

type contextKey string

// ValidateAuth - returns endpoint for auth validation
func (a *authorizor) ValidateAuth() endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {

			// xffProto, ok := ctx.Value(kithttp.ContextKeyRequestXForwardedProto).(string)
			// if strings.ToLower(xffProto) != "https" {
			// 	//http.Redirect()
			// 	err := errors.New("https required")
			// 	level.Error(a.logger).Log("err", err)
			// 	return nil, HttpNotAllowed
			// }

			token, ok := ctx.Value(jwt.JWTTokenContextKey).(string)
			if !ok {
				err := errors.New("Token Not Found")
				level.Error(a.logger).Log("err", err)
				return nil, ErrInvalidToken
			}

			claims, err := parseToken(a, token)
			if err != nil {
				level.Error(a.logger).Log("err", err)
				return nil, ErrInvalidToken
			}

			level.Info(a.logger).Log("UserNTID", claims.UserName, "UserEmail", claims.Email)

			if claims.ExpiresAt == 0 {
				level.Error(a.logger).Log("err", "JWT ExpiresAt invalid")
				return nil, ErrInvalidToken
			}

			if claims.UserName == "" {
				level.Error(a.logger).Log("err", "JWT Username invalid")
				return nil, ErrInvalidToken
			}

			if len(a.config.Auth.Groups) > 0 {
				req, _ := http.NewRequest(http.MethodGet, a.config.Auth.ProfileURL, nil)
				req.Header.Set("Authorization", fmt.Sprintf("%s %s", "Bearer", token))

				u := url.Values{}
				u.Set("group", strings.Join(a.config.Auth.Groups, ","))
				req.URL.RawQuery = u.Encode()

				resp, err := a.httpClient.Do(req)
				if err != nil {
					level.Error(a.logger).Log("err", "Unable to retrieve user authorization")
					return nil, errors.New("Unable to retrieve user authorization")
				}
				defer resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					return nil, ErrForbidden
				}
			}

			ctx = context.WithValue(ctx, AuthUserID, claims.UserName)
			return next(ctx, request)
		}
	}
}

func parseToken(a *authorizor, tokenString string) (claims *Claims, err error) {
	token, err := stdjwt.ParseWithClaims(tokenString, &Claims{},
		func(token *stdjwt.Token) (interface{}, error) { return a.config.PubKey, nil })
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	// Check validity of token
	if !token.Valid {
		return nil, errors.New("Token is invalid")
	}

	// Check claims have been parsed
	if claims, ok := token.Claims.(*Claims); ok {
		return claims, err
	}
	return nil, errors.New("Unable to parse claims")
}
