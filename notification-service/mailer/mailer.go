// Package mailer implementa o envio de e-mails de notificação.
package mailer

import (
	"fmt"
	"log"
	"net/smtp"
)

// Mailer define a interface de envio de e-mails.
type Mailer interface {
	Send(to, subject, body string) error
}

// LogMailer registra os e-mails no log quando o SMTP não está configurado.
type LogMailer struct {
	from string
}

// Send registra o e-mail no log em vez de enviá-lo.
func (m *LogMailer) Send(to, subject, body string) error {
	log.Printf("[MAIL] From: %s | To: %s | Subject: %s\n%s", m.from, to, subject, body)
	return nil
}

// SMTPMailer envia e-mails via SMTP.
type SMTPMailer struct {
	host string
	port string
	user string
	pass string
	from string
}

// Send envia um e-mail via SMTP usando PlainAuth.
func (m *SMTPMailer) Send(to, subject, body string) error {
	addr := fmt.Sprintf("%s:%s", m.host, m.port)
	auth := smtp.PlainAuth("", m.user, m.pass, m.host)
	msg := []byte(fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s",
		m.from, to, subject, body,
	))
	return smtp.SendMail(addr, auth, m.from, []string{to}, msg)
}

// NewMailer cria um Mailer adequado conforme a configuração SMTP.
// Retorna LogMailer se o host estiver vazio.
func NewMailer(host, port, user, pass, from string) Mailer {
	if host == "" {
		return &LogMailer{from: from}
	}
	return &SMTPMailer{
		host: host,
		port: port,
		user: user,
		pass: pass,
		from: from,
	}
}
