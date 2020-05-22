package config

import "crypto/rsa"

//Config ...
type Config struct {
	PubKey *rsa.PublicKey
	Auth   authConfig `json:"auth,omitempty"`
	Mailer mailConfig `json:"mail,omitempty"`
	HTTP   httpConfig `json:"http,omitempty"`
	OSConf envConfig
}

//authentication client config
type authConfig struct {
	ProfileURL       string   `json:"profileURL"`
	Groups           []string `json:"groups"`
	ClientID         string   `json:"clientID"`
	Username         string   `json:"username"`
	AuthorizationURL string   `json:"authURL"`
	TokenURL         string   `json:"tokenURL"`
	Scopes           []string `json:"scopes"`
	RenewInterval    int      `json:"renewInterval"`
}

//outbound http client config
type httpConfig struct {
	TimeOut        int               `json:"timeout"`
	APISecret      string            `json:"apiSecret"`
	DataServiceURL map[string]string `json:"dataServiceURL"`
	FileUploadURL  string            `json:"fileUploadURL"`
}

//email notification client config
type mailConfig struct {
	FromEmail      string            `json:"fromEmail,omitempty"`
	MailHost       string            `json:"mailhost,omitempty"`
	Subject        string            `json:"subject,omitempty"`
	EmailTemplates map[string]string `json:"templates,omitempty"`
}

//envConfig stores the enviornment level configuration passed
type envConfig struct {
	ClientSecret       string
	CPPServicePassword string
}
