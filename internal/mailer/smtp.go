package mailer

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"time"

	gomail "gopkg.in/gomail.v2"
)

//go:embed templates/*.html
var templateFS embed.FS

type Mailer struct {
	fromEmail string
	dialer    *gomail.Dialer
}

type ActivationData struct {
	Username       string
	ActivationLink string
}

func NewMailer(host string, port int, fromEmail string) *Mailer {
	dialer := gomail.NewDialer(host, port, "", "")
	return &Mailer{
		fromEmail: fromEmail,
		dialer:    dialer,
	}
}

func (m *Mailer) Send(templateFile string, username string, email string, data any) error {
	tmpl, err := template.ParseFS(templateFS, "templates/"+templateFile)
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("execute template: %w", err)
	}

	msg := gomail.NewMessage()
	msg.SetHeader("From", m.fromEmail)
	msg.SetHeader("To", email)
	msg.SetHeader("Subject", "Welcome to Social!")
	msg.SetBody("text/html", buf.String())

	var sendErr error
	for i := range 3 {
		if sendErr = m.dialer.DialAndSend(msg); sendErr == nil {
			return nil
		}
		time.Sleep(time.Duration(i+1) * time.Second)
	}

	return sendErr
}
