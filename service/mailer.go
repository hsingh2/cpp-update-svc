package app

import (
	"bytes"
	"net/smtp"
	"text/template"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.comcast.com/cpp/cpp-update-svc/config"
)

// Mailer - interface for mailer functions
type Mailer interface {
	SendMail(Message) error
}

//Message ...
type Message struct {
	EmailTo  string
	UserName string
	Type     string
	Body     []Error
}

type mail struct {
	cfg    config.Config
	logger log.Logger
}

// NewMailer - returns new mailer interface
func NewMailer(c config.Config, l log.Logger) Mailer {
	return &mail{
		logger: l,
		cfg:    c,
	}
}

func (m *mail) SendMail(msg Message) error {
	c, err := smtp.Dial(m.cfg.Mailer.MailHost)
	if err != nil {
		level.Error(m.logger).Log("Error dialing in SMTP", err)
		return err
	}

	err = c.Mail(m.cfg.Mailer.FromEmail)
	if err != nil {
		level.Error(m.logger).Log("Error sending mail", err)
		return err
	}

	err = c.Rcpt(msg.EmailTo)
	if err != nil {
		level.Error(m.logger).Log("Error sending mail rcpt", err)
		return err
	}

	wc, err := c.Data()
	if err != nil {
		level.Error(m.logger).Log("Error getting mail rcpt", err)
	}
	defer wc.Close()

	//parse template

	t, err := template.ParseFiles(m.cfg.Mailer.EmailTemplates[msg.Type])
	if err != nil {
		level.Error(m.logger).Log("Could not read email template file", err)
	}

	buf := new(bytes.Buffer)
	if err := t.Execute(buf, msg); err != nil {
		level.Error(m.logger).Log("Could not parse email template", err)
	}

	//create message
	message := "To: " + msg.EmailTo + "\r\n" +
		"From: " + m.cfg.Mailer.FromEmail + "\r\n" +
		"Subject: " + m.cfg.Mailer.Subject + "\r\n" +
		"Content-Type: text/html; charset=\"UTF-8\"\r\n" +
		"\r\n" + buf.String()

	wc.Write([]byte(message))

	return nil
}
