package mailer

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/smtp"
	"strconv"
	"strings"
	"time"
)

type MailtrapMailer struct {
	host      string
	port      int
	username  string
	password  string
	fromEmail string
}

func NewMailtrap(host string, port int, username, password, fromEmail string) *MailtrapMailer {
	if port <= 0 {
		port = 2525
	}

	return &MailtrapMailer{
		host:      host,
		port:      port,
		username:  username,
		password:  password,
		fromEmail: fromEmail,
	}
}

func (m *MailtrapMailer) Send(templateFile, username, email string, data any, isSandbox bool) error {
	if m.host == "" || m.username == "" || m.password == "" || m.fromEmail == "" {
		return errors.New("mailtrap config is incomplete")
	}

	tmpl, err := template.ParseFS(FS, "templates/"+templateFile)
	if err != nil {
		return err
	}

	subject := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return err
	}

	body := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(body, "body", data)
	if err != nil {
		return err
	}

	addr := net.JoinHostPort(m.host, strconv.Itoa(m.port))
	auth := smtp.PlainAuth("", m.username, m.password, m.host)

	safeSubject := strings.ReplaceAll(subject.String(), "\n", " ")
	safeSubject = strings.ReplaceAll(safeSubject, "\r", " ")
	toLine := fmt.Sprintf("%s <%s>", username, email)
	fromLine := fmt.Sprintf("%s <%s>", FromName, m.fromEmail)

	raw := strings.Join([]string{
		fmt.Sprintf("From: %s", fromLine),
		fmt.Sprintf("To: %s", toLine),
		fmt.Sprintf("Subject: %s", safeSubject),
		"MIME-Version: 1.0",
		`Content-Type: text/html; charset="UTF-8"`,
		"",
		body.String(),
	}, "\r\n")

	if isSandbox {
		log.Printf("mailer sandbox mode enabled for %s", email)
	}

	for i := 0; i < maxRetries; i++ {
		err = smtp.SendMail(addr, auth, m.fromEmail, []string{email}, []byte(raw))
		if err != nil {
			log.Printf("failed to send email to %s, attempt %d of %d", email, i+1, maxRetries)
			log.Printf("error: %v", err)

			time.Sleep(time.Second * time.Duration(i+1))
			continue
		}
		log.Printf("email sent to %s", email)
		return nil
	}

	return fmt.Errorf("failed to send email after %d attempts", maxRetries)
}
